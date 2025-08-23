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
  templateUrl: './description-input.component.html',
  styleUrl: './description-input.component.css'
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