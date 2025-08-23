import { Component, Input, Output, EventEmitter, OnInit, inject, signal, OnDestroy } from '@angular/core';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { Subject, debounceTime, distinctUntilChanged, switchMap, takeUntil, of } from 'rxjs';
import { ExternalIntegrationService } from '../../../services/external-integration.service';
import { IssueResolveResult } from '../../../models/external-issue.model';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';

// PrimeNG imports
import { InputText } from 'primeng/inputtext';
import { AutoComplete } from 'primeng/autocomplete';
import { Chip } from 'primeng/chip';
import { ProgressSpinner } from 'primeng/progressspinner';

@Component({
  selector: 'app-description-input',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    TranslatePipe,
    InputText,
    AutoComplete,
    Chip,
    ProgressSpinner
  ],
  template: `
    <div class="description-input-container">
      <div class="input-wrapper">
        <p-autoComplete
          [formControl]="control"
          [suggestions]="suggestions()"
          (completeMethod)="search($event)"
          (onInput)="onInputChange($event)"
          [placeholder]="placeholder"
          [disabled]="disabled"
          styleClass="w-full"
          field="value"
          [dropdown]="false"
          [multiple]="false">
        </p-autoComplete>
        
        @if (isResolving()) {
          <div class="resolving-indicator">
            <p-progressSpinner 
              [style]="{ width: '16px', height: '16px' }"
              strokeWidth="4">
            </p-progressSpinner>
            <span class="resolving-text">{{ 'timeEntry.form.resolvingIssue' | translate }}</span>
          </div>
        }
      </div>
      
      @if (resolvedIssue()) {
        <div class="resolved-issue">
          <p-chip 
            [label]="getIssueChipLabel()"
            icon="pi pi-external-link"
            styleClass="issue-chip"
            [removable]="true"
            (onRemove)="clearResolvedIssue()">
          </p-chip>
        </div>
      }
      
      @if (pendingIssue()) {
        <div class="pending-issue">
          <p-chip 
            [label]="getPendingIssueLabel()"
            icon="pi pi-clock"
            severity="warning"
            styleClass="pending-chip">
          </p-chip>
        </div>
      }
    </div>
  `,
  styles: [`
    .description-input-container {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }
    
    .input-wrapper {
      position: relative;
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }
    
    .resolving-indicator {
      display: flex;
      align-items: center;
      gap: 0.25rem;
      font-size: 0.75rem;
      color: #6b7280;
    }
    
    .resolved-issue,
    .pending-issue {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
    }
    
    :host ::ng-deep .issue-chip {
      background-color: #10b981;
      color: white;
    }
    
    :host ::ng-deep .pending-chip {
      background-color: #f59e0b;
      color: white;
    }
    
    :host ::ng-deep .issue-chip .p-chip-text {
      cursor: pointer;
    }
    
    .resolving-text {
      font-size: 0.75rem;
    }
  `]
})
export class DescriptionInputComponent implements OnInit, OnDestroy {
  @Input() control!: FormControl;
  @Input() projectId?: string;
  @Input() placeholder = 'Enter description...';
  @Input() disabled = false;
  
  @Output() issueResolved = new EventEmitter<IssueResolveResult>();
  @Output() issueCleared = new EventEmitter<void>();

  private readonly externalService = inject(ExternalIntegrationService);
  private readonly translateService = inject(TranslateService);
  private readonly destroy$ = new Subject<void>();
  
  suggestions = signal<string[]>([]);
  resolvedIssue = signal<IssueResolveResult | null>(null);
  pendingIssue = signal<string | null>(null);
  isResolving = signal(false);

  private suggestionSubject = new Subject<string>();
  private issueDetectionSubject = new Subject<string>();

  ngOnInit() {
    this.setupAutocompleteSuggestions();
    this.setupIssueDetection();
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private setupAutocompleteSuggestions() {
    this.suggestionSubject.pipe(
      debounceTime(150),
      distinctUntilChanged(),
      switchMap(query => {
        if (!this.projectId || query.length < 2) {
          return of([]);
        }
        return this.externalService.getDescriptionSuggestions(this.projectId, query);
      }),
      takeUntil(this.destroy$)
    ).subscribe({
      next: (response) => {
        this.suggestions.set(response.suggestions);
      },
      error: (error) => {
        console.error('Error fetching suggestions:', error);
        this.suggestions.set([]);
      }
    });
  }

  private setupIssueDetection() {
    this.issueDetectionSubject.pipe(
      debounceTime(300),
      distinctUntilChanged(),
      switchMap(input => {
        if (!this.projectId || !input.trim()) {
          return of(null);
        }

        const issuePattern = this.externalService.detectIssuePattern(input);
        if (!issuePattern) {
          return of(null);
        }

        this.isResolving.set(true);
        return this.externalService.resolveIssue(this.projectId, input);
      }),
      takeUntil(this.destroy$)
    ).subscribe({
      next: (result) => {
        this.isResolving.set(false);
        
        if (result) {
          if (result.status === 'resolved') {
            this.resolvedIssue.set(result);
            this.pendingIssue.set(null);
            this.issueResolved.emit(result);
          } else if (result.status === 'pending') {
            this.pendingIssue.set(result.key);
            this.resolvedIssue.set(null);
          }
        } else {
          this.clearAllIssues();
        }
      },
      error: (error) => {
        console.error('Error resolving issue:', error);
        this.isResolving.set(false);
        this.clearAllIssues();
      }
    });
  }

  search(event: any) {
    const query = event.query;
    this.suggestionSubject.next(query);
  }

  onInputChange(event: any) {
    const input = event.target.value;
    this.issueDetectionSubject.next(input);
  }

  getIssueChipLabel(): string {
    const issue = this.resolvedIssue();
    if (issue?.title && issue.key) {
      return `${issue.key} · ${issue.title}`;
    }
    return issue?.key || '';
  }

  getPendingIssueLabel(): string {
    const pending = this.pendingIssue();
    return pending ? 
      `${pending} (${this.translateService.instant('timeEntry.form.pending')})` : 
      '';
  }

  clearResolvedIssue() {
    this.resolvedIssue.set(null);
    this.issueCleared.emit();
  }

  private clearAllIssues() {
    this.resolvedIssue.set(null);
    this.pendingIssue.set(null);
  }
}