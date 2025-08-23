import { Component, Input, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import { ExternalIntegrationService } from '../../../services/external-integration.service';
import { UserExternalAccountService } from '../../../services/user-external-account.service';
import { ConnectProjectToAccountRequest } from '../../../models/external-connection.model';
import { UserExternalAccount } from '../../../models/user-external-account.model';
import { Project } from '../../../models/project.model';
import { MessageService } from 'primeng/api';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { environment } from '../../../../environments/environment';

// PrimeNG imports
import { DropdownModule } from 'primeng/dropdown';
import { InputText } from 'primeng/inputtext';
import { Button } from 'primeng/button';
import { Card } from 'primeng/card';
import { Message } from 'primeng/message';
import { ProgressSpinner } from 'primeng/progressspinner';

@Component({
  selector: 'app-external-integration',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    ReactiveFormsModule,
    TranslatePipe,
    DropdownModule,
    InputText,
    Button,
    Card,
    Message,
    ProgressSpinner
  ],
  providers: [MessageService],
  template: `
    <p-card [header]="'project.external.title' | translate">
      <div class="external-integration-form">
        @if (isConnected()) {
          <!-- Connected state -->
          <p-message severity="success" [text]="getConnectionMessage()"></p-message>
          <div class="connection-actions">
            <p-button 
              [label]="'project.external.sync' | translate"
              icon="pi pi-refresh"
              (onClick)="syncIssues()"
              [loading]="isSyncing()"
              styleClass="p-button-outlined">
            </p-button>
            <p-button 
              [label]="'project.external.disconnect' | translate"
              icon="pi pi-unlink"
              (onClick)="disconnect()"
              [loading]="isProcessing()"
              severity="danger"
              styleClass="p-button-outlined">
            </p-button>
          </div>
        } @else {
          <!-- Connection form -->
          @if (isLoadingAccounts()) {
            <div class="loading-container">
              <p-progressSpinner></p-progressSpinner>
              <span>{{ 'user.externalAccounts.loading' | translate }}</span>
            </div>
          } @else if (userAccounts().length === 0) {
            <p-message 
              severity="info" 
              [text]="'project.external.noAccounts' | translate">
            </p-message>
            <p-button 
              [label]="'project.external.manageAccounts' | translate"
              icon="pi pi-external-link"
              routerLink="/user/external-accounts"
              styleClass="p-button-outlined">
            </p-button>
          } @else {
            <form [formGroup]="connectionForm" (ngSubmit)="connect()">
              <div class="form-field">
                <label for="userAccountId">{{ 'project.external.account' | translate }}</label>
                <p-dropdown
                  id="userAccountId"
                  formControlName="userAccountId"
                  [options]="getAccountOptions()"
                  optionLabel="label"
                  optionValue="value"
                  [placeholder]="'project.external.selectAccount' | translate"
                  styleClass="w-full">
                  <ng-template let-account pTemplate="item">
                    <div class="account-option">
                      <i [class]="getProviderIcon(account.provider)"></i>
                      <span>{{ account.accountName }} ({{ getProviderDisplayName(account.provider) }})</span>
                    </div>
                  </ng-template>
                </p-dropdown>
              </div>
              
              <div class="form-field">
                <label for="projectRef">{{ 'project.external.projectRef' | translate }}</label>
                <input
                  pInputText
                  id="projectRef"
                  formControlName="projectRef"
                  [placeholder]="getProjectRefPlaceholder()"
                  class="w-full">
                <small class="form-help">{{ getProjectRefHelp() }}</small>
              </div>
              
              <div class="form-actions">
                <p-button
                  type="submit"
                  [label]="'project.external.connect' | translate"
                  icon="pi pi-link"
                  [disabled]="!connectionForm.valid"
                  [loading]="isProcessing()">
                </p-button>
              </div>
            </form>
          }
        }
      </div>
    </p-card>
  `,
  styles: [`
    .external-integration-form {
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }
    
    .form-field {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }
    
    .form-field label {
      font-weight: 600;
      font-size: 0.875rem;
    }
    
    .form-help {
      color: #6b7280;
      font-size: 0.75rem;
    }
    
    .form-actions {
      display: flex;
      justify-content: flex-end;
      margin-top: 1rem;
    }
    
    .connection-actions {
      display: flex;
      gap: 0.5rem;
      margin-top: 1rem;
    }

    .loading-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 1rem;
      padding: 2rem;
    }

    .account-option {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }
  `]
})
export class ExternalIntegrationComponent implements OnInit {
  @Input() projectId!: string;
  
  private readonly formBuilder = inject(FormBuilder);
  private readonly http = inject(HttpClient);
  private readonly externalService = inject(ExternalIntegrationService);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);
  private readonly userAccountService = inject(UserExternalAccountService);

  connectionForm!: FormGroup;
  isConnected = signal(false);
  currentProvider = signal<string>('');
  currentProjectRef = signal<string>('');
  isProcessing = signal(false);
  isSyncing = signal(false);
  userAccounts = signal<UserExternalAccount[]>([]);
  isLoadingAccounts = signal(false);

  ngOnInit() {
    this.initializeForm();
    this.checkConnectionStatus();
  }

  private initializeForm() {
    this.connectionForm = this.formBuilder.group({
      userAccountId: ['', [Validators.required]],
      projectRef: ['', [Validators.required]]
    });
  }

  private checkConnectionStatus() {
    // Load user accounts and check connection status
    this.loadUserAccounts();
    this.checkExistingConnection();
  }

  private checkExistingConnection() {
    if (!this.projectId) return;

    // Make direct HTTP call to get project details with external connection data
    const projectUrl = `${environment.apiUrl}/projects/${this.projectId}`;
    this.http.get<Project>(projectUrl).subscribe({
      next: (project) => {
        this.updateConnectionStatus(project);
      },
      error: (error) => {
        console.warn('Could not load project details for external connection check:', error);
        this.isConnected.set(false);
      }
    });
  }

  private updateConnectionStatus(projectData: Project) {
    if (projectData?.externalConnection) {
      this.isConnected.set(true);
      this.currentProvider.set(projectData.externalConnection.provider);
      this.currentProjectRef.set(projectData.externalConnection.projectRef);
      
      // Pre-populate form with existing connection data
      this.connectionForm.patchValue({
        userAccountId: projectData.externalConnection.userAccountId,
        projectRef: projectData.externalConnection.projectRef
      });
    } else {
      this.isConnected.set(false);
      this.currentProvider.set('');
      this.currentProjectRef.set('');
    }
  }

  private loadUserAccounts() {
    this.isLoadingAccounts.set(true);
    this.userAccountService.getAccounts().subscribe({
      next: (response) => {
        this.userAccounts.set(response.accounts);
        this.isLoadingAccounts.set(false);
      },
      error: (error) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('error'),
          detail: this.translateService.instant('user.externalAccounts.loadError')
        });
        this.isLoadingAccounts.set(false);
      }
    });
  }

  getAccountOptions() {
    return this.userAccounts().map(account => ({
      label: `${account.accountName} (${this.getProviderDisplayName(account.provider)})`,
      value: account.id,
      provider: account.provider
    }));
  }

  getSelectedAccountProvider(): string {
    const accountId = this.connectionForm.get('userAccountId')?.value;
    if (!accountId) return '';
    const account = this.userAccounts().find(a => a.id === accountId);
    return account?.provider || '';
  }

  getProjectRefPlaceholder(): string {
    const provider = this.getSelectedAccountProvider();
    switch (provider) {
      case 'github':
      case 'gitlab':
        return 'owner/repository-name';
      case 'jira':
        return 'PROJECT-KEY';
      default:
        return this.translateService.instant('project.external.projectRefPlaceholder');
    }
  }

  getProjectRefHelp(): string {
    const provider = this.getSelectedAccountProvider();
    switch (provider) {
      case 'github':
        return this.translateService.instant('project.external.githubHelp');
      case 'gitlab':
        return this.translateService.instant('project.external.gitlabHelp');
      case 'jira':
        return this.translateService.instant('project.external.jiraHelp');
      default:
        return '';
    }
  }

  getProviderIcon(provider: string): string {
    switch (provider) {
      case 'github':
        return 'pi pi-github';
      case 'gitlab':
        return 'pi pi-code';
      case 'jira':
        return 'pi pi-ticket';
      default:
        return 'pi pi-link';
    }
  }

  getProviderDisplayName(provider: string): string {
    switch (provider) {
      case 'github':
        return 'GitHub';
      case 'gitlab':
        return 'GitLab';
      case 'jira':
        return 'Jira';
      default:
        return provider;
    }
  }

  getConnectionMessage(): string {
    return this.translateService.instant('project.external.connectedTo', {
      provider: this.currentProvider(),
      projectRef: this.currentProjectRef()
    });
  }

  connect() {
    if (!this.connectionForm.valid || !this.projectId) return;

    this.isProcessing.set(true);

    const request: ConnectProjectToAccountRequest = this.connectionForm.value;
    const selectedAccount = this.userAccounts().find(a => a.id === request.userAccountId);

    this.userAccountService.connectProjectToAccount(this.projectId, request).subscribe({
      next: (response) => {
        this.messageService.add({
          severity: 'success',
          summary: this.translateService.instant('success'),
          detail: this.translateService.instant('project.external.connectSuccess')
        });
        
        this.isConnected.set(true);
        this.currentProvider.set(selectedAccount?.provider || '');
        this.currentProjectRef.set(request.projectRef);
        this.connectionForm.reset();
        this.isProcessing.set(false);
      },
      error: (error) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('error'),
          detail: error.error?.error || this.translateService.instant('project.external.connectError')
        });
        this.isProcessing.set(false);
      }
    });
  }

  disconnect() {
    if (!this.projectId) return;

    this.isProcessing.set(true);

    this.userAccountService.disconnectProject(this.projectId).subscribe({
      next: (response) => {
        this.messageService.add({
          severity: 'success',
          summary: this.translateService.instant('success'),
          detail: this.translateService.instant('project.external.disconnectSuccess')
        });
        
        this.isConnected.set(false);
        this.currentProvider.set('');
        this.currentProjectRef.set('');
        this.isProcessing.set(false);
      },
      error: (error) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('error'),
          detail: error.error?.error || this.translateService.instant('project.external.disconnectError')
        });
        this.isProcessing.set(false);
      }
    });
  }

  syncIssues() {
    if (!this.projectId) return;

    this.isSyncing.set(true);

    this.externalService.syncProjectIssues(this.projectId).subscribe({
      next: (response) => {
        this.messageService.add({
          severity: 'success',
          summary: this.translateService.instant('success'),
          detail: this.translateService.instant('project.external.syncSuccess')
        });
        this.isSyncing.set(false);
      },
      error: (error) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('error'),
          detail: error.error?.error || this.translateService.instant('project.external.syncError')
        });
        this.isSyncing.set(false);
      }
    });
  }
}