import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { UserExternalAccountService } from '../../../services/user-external-account.service';
import { UserExternalAccount, UserExternalAccountRequest } from '../../../models/user-external-account.model';
import { MessageService } from 'primeng/api';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';

// PrimeNG imports
import { Card } from 'primeng/card';
import { Button } from 'primeng/button';
import { Dialog } from 'primeng/dialog';
import { InputText } from 'primeng/inputtext';
import { DropdownModule } from 'primeng/dropdown';
import { TableModule } from 'primeng/table';
import { ConfirmDialog } from 'primeng/confirmdialog';
import { Toast } from 'primeng/toast';
import { ProgressSpinner } from 'primeng/progressspinner';
import { ConfirmationService } from 'primeng/api';

@Component({
  selector: 'app-external-accounts',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    TranslatePipe,
    Card,
    Button,
    Dialog,
    InputText,
    DropdownModule,
    TableModule,
    ConfirmDialog,
    Toast,
    ProgressSpinner
  ],
  providers: [MessageService, ConfirmationService],
  template: `
    <div class="external-accounts-container">
      <p-card>
        <ng-template pTemplate="header">
          <div class="card-header">
            <h2>{{ 'user.externalAccounts.title' | translate }}</h2>
            <p-button
              [label]="'user.externalAccounts.addAccount' | translate"
              icon="pi pi-plus"
              (onClick)="openCreateDialog()"
              styleClass="p-button-success">
            </p-button>
          </div>
        </ng-template>

        <ng-template pTemplate="content">
          @if (isLoading()) {
            <div class="loading-container">
              <p-progressSpinner></p-progressSpinner>
              <span>{{ 'user.externalAccounts.loading' | translate }}</span>
            </div>
          } @else if (accounts().length === 0) {
            <div class="no-accounts">
              <i class="pi pi-info-circle"></i>
              <h3>{{ 'user.externalAccounts.noAccounts' | translate }}</h3>
              <p>{{ 'user.externalAccounts.noAccountsDescription' | translate }}</p>
            </div>
          } @else {
            <p-table [value]="accounts()" [responsiveLayout]="'scroll'">
              <ng-template pTemplate="header">
                <tr>
                  <th>{{ 'user.externalAccounts.provider' | translate }}</th>
                  <th>{{ 'user.externalAccounts.accountName' | translate }}</th>
                  <th>{{ 'user.externalAccounts.baseUrl' | translate }}</th>
                  <th>{{ 'user.externalAccounts.actions' | translate }}</th>
                </tr>
              </ng-template>
              <ng-template pTemplate="body" let-account>
                <tr>
                  <td>
                    <div class="provider-cell">
                      <i [class]="getProviderIcon(account.provider)"></i>
                      <span>{{ getProviderDisplayName(account.provider) }}</span>
                    </div>
                  </td>
                  <td>{{ account.accountName }}</td>
                  <td>{{ account.baseURL || '-' }}</td>
                  <td>
                    <div class="action-buttons">
                      <p-button
                        icon="pi pi-check-circle"
                        styleClass="p-button-rounded p-button-text p-button-success"
                        (onClick)="testAccount(account.id)"
                        [loading]="testingAccounts().has(account.id)">
                      </p-button>
                      <p-button
                        icon="pi pi-pencil"
                        styleClass="p-button-rounded p-button-text"
                        (onClick)="openEditDialog(account)">
                      </p-button>
                      <p-button
                        icon="pi pi-trash"
                        styleClass="p-button-rounded p-button-text p-button-danger"
                        (onClick)="confirmDelete(account)">
                      </p-button>
                    </div>
                  </td>
                </tr>
              </ng-template>
            </p-table>
          }
        </ng-template>
      </p-card>

      <!-- Create/Edit Dialog -->
      <p-dialog
        [(visible)]="showDialog"
        [header]="dialogTitle()"
        [modal]="true"
        [style]="{ width: '500px' }"
        [closable]="true">
        
        <form [formGroup]="accountForm" (ngSubmit)="saveAccount()">
          <div class="form-grid">
            <div class="form-field">
              <label for="provider">{{ 'user.externalAccounts.provider' | translate }}</label>
              <p-dropdown
                id="provider"
                formControlName="provider"
                [options]="providerOptions"
                optionLabel="label"
                optionValue="value"
                [placeholder]="'user.externalAccounts.selectProvider' | translate"
                styleClass="w-full">
              </p-dropdown>
            </div>

            <div class="form-field">
              <label for="accountName">{{ 'user.externalAccounts.accountName' | translate }}</label>
              <input
                pInputText
                id="accountName"
                formControlName="accountName"
                [placeholder]="'user.externalAccounts.accountNamePlaceholder' | translate"
                class="w-full">
            </div>

            <div class="form-field">
              <label for="oauthToken">{{ 'user.externalAccounts.oauthToken' | translate }}</label>
              <input
                pInputText
                id="oauthToken"
                type="password"
                formControlName="oauthToken"
                [placeholder]="'user.externalAccounts.tokenPlaceholder' | translate"
                class="w-full">
              <small class="form-help">{{ 'user.externalAccounts.tokenHelp' | translate }}</small>
            </div>

            @if (shouldShowBaseUrl()) {
              <div class="form-field">
                <label for="baseURL">{{ 'user.externalAccounts.baseUrl' | translate }}</label>
                <input
                  pInputText
                  id="baseURL"
                  formControlName="baseURL"
                  [placeholder]="getBaseUrlPlaceholder()"
                  class="w-full">
                <small class="form-help">{{ getBaseUrlHelp() }}</small>
              </div>
            }
          </div>

          <div class="form-actions">
            <p-button
              type="button"
              [label]="'globals.cancel' | translate"
              icon="pi pi-times"
              styleClass="p-button-outlined"
              (onClick)="closeDialog()">
            </p-button>
            <p-button
              type="submit"
              [label]="'globals.save' | translate"
              icon="pi pi-check"
              [disabled]="!accountForm.valid"
              [loading]="isSaving()">
            </p-button>
          </div>
        </form>
      </p-dialog>

      <p-confirmDialog></p-confirmDialog>
      <p-toast position="top-right"></p-toast>
    </div>
  `,
  styles: [`
    .external-accounts-container {
      max-width: 1200px;
      margin: 0 auto;
      padding: 1rem;
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      width: 100%;
    }

    .loading-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 1rem;
      padding: 2rem;
    }

    .no-accounts {
      text-align: center;
      padding: 2rem;
      color: #6b7280;
    }

    .no-accounts i {
      font-size: 3rem;
      margin-bottom: 1rem;
    }

    .provider-cell {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .action-buttons {
      display: flex;
      gap: 0.25rem;
    }

    .form-grid {
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
      gap: 0.5rem;
      margin-top: 1.5rem;
    }
  `]
})
export class ExternalAccountsComponent implements OnInit {
  private readonly formBuilder = inject(FormBuilder);
  private readonly accountService = inject(UserExternalAccountService);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);
  private readonly confirmationService = inject(ConfirmationService);

  accounts = signal<UserExternalAccount[]>([]);
  isLoading = signal(true);
  isSaving = signal(false);
  testingAccounts = signal(new Set<string>());
  
  showDialog = false;
  editingAccount: UserExternalAccount | null = null;
  accountForm!: FormGroup;

  providerOptions = [
    { label: 'GitHub', value: 'github' },
    { label: 'GitLab', value: 'gitlab' },
    { label: 'Jira', value: 'jira' }
  ];

  ngOnInit() {
    this.initializeForm();
    this.loadAccounts();
  }

  private initializeForm() {
    this.accountForm = this.formBuilder.group({
      provider: ['', [Validators.required]],
      accountName: ['', [Validators.required]],
      oauthToken: ['', [Validators.required]],
      baseURL: ['']
    });
  }

  private loadAccounts() {
    this.isLoading.set(true);
    console.log('Loading external accounts...');
    
    this.accountService.getAccounts().subscribe({
      next: (response) => {
        console.log('External accounts response:', response);
        console.log('Accounts array:', response.accounts);
        console.log('Accounts length:', response.accounts?.length);
        
        this.accounts.set(response.accounts || []);
        this.isLoading.set(false);
        
        console.log('Loading state set to false');
        console.log('Current isLoading signal:', this.isLoading());
      },
      error: (error) => {
        console.error('Error loading accounts:', error);
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('globals.error'),
          detail: this.translateService.instant('user.externalAccounts.loadError')
        });
        this.isLoading.set(false);
        console.log('Error - Loading state set to false');
      }
    });
  }

  dialogTitle() {
    return this.editingAccount
      ? this.translateService.instant('user.externalAccounts.editAccount')
      : this.translateService.instant('user.externalAccounts.addAccount');
  }

  shouldShowBaseUrl(): boolean {
    const provider = this.accountForm.get('provider')?.value;
    return provider === 'gitlab' || provider === 'jira';
  }

  getBaseUrlPlaceholder(): string {
    const provider = this.accountForm.get('provider')?.value;
    switch (provider) {
      case 'gitlab':
        return 'https://gitlab.example.com';
      case 'jira':
        return 'https://yourcompany.atlassian.net';
      default:
        return '';
    }
  }

  getBaseUrlHelp(): string {
    const provider = this.accountForm.get('provider')?.value;
    switch (provider) {
      case 'gitlab':
        return this.translateService.instant('user.externalAccounts.gitlabUrlHelp');
      case 'jira':
        return this.translateService.instant('user.externalAccounts.jiraUrlHelp');
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

  openCreateDialog() {
    this.editingAccount = null;
    this.accountForm.reset();
    this.showDialog = true;
  }

  openEditDialog(account: UserExternalAccount) {
    this.editingAccount = account;
    this.accountForm.patchValue({
      provider: account.provider,
      accountName: account.accountName,
      oauthToken: '', // Don't pre-fill the token for security
      baseURL: account.baseURL || ''
    });
    this.showDialog = true;
  }

  closeDialog() {
    this.showDialog = false;
    this.editingAccount = null;
    this.accountForm.reset();
  }

  saveAccount() {
    if (!this.accountForm.valid) return;

    this.isSaving.set(true);
    const formValue = this.accountForm.value as UserExternalAccountRequest;

    const operation = this.editingAccount
      ? this.accountService.updateAccount(this.editingAccount.id, formValue)
      : this.accountService.createAccount(formValue);

    operation.subscribe({
      next: (account) => {
        this.messageService.add({
          severity: 'success',
          summary: this.translateService.instant('globals.success'),
          detail: this.editingAccount
            ? this.translateService.instant('user.externalAccounts.updateSuccess')
            : this.translateService.instant('user.externalAccounts.createSuccess')
        });
        
        this.loadAccounts();
        this.closeDialog();
        this.isSaving.set(false);
      },
      error: (error) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('globals.error'),
          detail: error.error?.error || this.translateService.instant('user.externalAccounts.saveError')
        });
        this.isSaving.set(false);
      }
    });
  }

  testAccount(accountId: string) {
    const testing = new Set(this.testingAccounts());
    testing.add(accountId);
    this.testingAccounts.set(testing);

    this.accountService.testAccount(accountId).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: this.translateService.instant('globals.success'),
          detail: this.translateService.instant('user.externalAccounts.testSuccess')
        });
        
        const testing = new Set(this.testingAccounts());
        testing.delete(accountId);
        this.testingAccounts.set(testing);
      },
      error: (error) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('globals.error'),
          detail: error.error?.error || this.translateService.instant('user.externalAccounts.testError')
        });
        
        const testing = new Set(this.testingAccounts());
        testing.delete(accountId);
        this.testingAccounts.set(testing);
      }
    });
  }

  confirmDelete(account: UserExternalAccount) {
    this.confirmationService.confirm({
      message: this.translateService.instant('user.externalAccounts.deleteConfirmMessage', { name: account.accountName }),
      header: this.translateService.instant('user.externalAccounts.deleteConfirmTitle'),
      icon: 'pi pi-exclamation-triangle',
      acceptButtonStyleClass: 'p-button-danger',
      accept: () => {
        this.deleteAccount(account.id);
      }
    });
  }

  private deleteAccount(accountId: string) {
    this.accountService.deleteAccount(accountId).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: this.translateService.instant('globals.success'),
          detail: this.translateService.instant('user.externalAccounts.deleteSuccess')
        });
        this.loadAccounts();
      },
      error: (error) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('globals.error'),
          detail: error.error?.error || this.translateService.instant('user.externalAccounts.deleteError')
        });
      }
    });
  }
}