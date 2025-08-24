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
import { ToastModule } from 'primeng/toast';

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
    ProgressSpinner,
    ToastModule
  ],
  providers: [MessageService],
  templateUrl: './external-integration.component.html',
  styleUrl: './external-integration.component.css'
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
        console.error('Connect project error:', error);
        let errorMessage = this.translateService.instant('project.external.connectError');
        
        // Provide more specific error messages based on error type
        if (error.status === 401) {
          errorMessage = this.translateService.instant('project.external.authError');
        } else if (error.status === 403) {
          errorMessage = this.translateService.instant('project.external.accessError');
        } else if (error.status === 503) {
          errorMessage = this.translateService.instant('project.external.providerError');
        } else if (error.error?.error) {
          errorMessage = error.error.error;
        }
        
        console.log('Showing error message:', errorMessage);
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('error'),
          detail: errorMessage,
          life: 8000, // Show for 8 seconds for better readability
          sticky: false
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
        let errorMessage = this.translateService.instant('project.external.syncError');
        
        // Provide more specific error messages based on error type
        if (error.status === 401) {
          errorMessage = this.translateService.instant('project.external.authError');
        } else if (error.status === 404) {
          errorMessage = this.translateService.instant('project.external.noConnectionError');
        } else if (error.status === 503) {
          errorMessage = this.translateService.instant('project.external.providerError');
        } else if (error.error?.error) {
          errorMessage = error.error.error;
        }
        
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('error'),
          detail: errorMessage,
          life: 6000 // Show for 6 seconds for better readability
        });
        this.isSyncing.set(false);
      }
    });
  }
}