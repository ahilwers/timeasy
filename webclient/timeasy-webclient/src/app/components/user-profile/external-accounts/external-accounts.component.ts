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
  templateUrl: './external-accounts.component.html',
  styleUrl: './external-accounts.component.css'
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
    
    this.accountService.getAccounts().subscribe({
      next: (response) => {
        this.accounts.set(response.accounts || []);
        this.isLoading.set(false);
      },
      error: (error) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translateService.instant('globals.error'),
          detail: this.translateService.instant('user.externalAccounts.loadError')
        });
        this.isLoading.set(false);
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