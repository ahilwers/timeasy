import {Component, OnInit, inject, signal} from '@angular/core';
import {CommonModule} from '@angular/common';
import {FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {CardModule} from 'primeng/card';
import {InputTextModule} from 'primeng/inputtext';
import {ButtonModule} from 'primeng/button';
import {DropdownModule} from 'primeng/dropdown';
import {ToastModule} from 'primeng/toast';
import {PasswordModule} from 'primeng/password';
import {ProgressSpinnerModule} from 'primeng/progressspinner';
import {MessageService} from 'primeng/api';
import {TranslateModule, TranslateService} from '@ngx-translate/core';

import {User} from '../../models/user.model';
import {UserService} from '../../services/user.service';
import {UserProfile, PasswordChangeRequest} from '../../interfaces/auth-provider.interface';

@Component({
  selector: 'app-user-profile',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    CardModule,
    InputTextModule,
    ButtonModule,
    DropdownModule,
    ToastModule,
    PasswordModule,
    ProgressSpinnerModule,
    TranslateModule
  ],
  templateUrl: './user-profile.component.html',
  styleUrl: './user-profile.component.css'
})
export class UserProfileComponent implements OnInit {
  private readonly userService = inject(UserService);
  private readonly formBuilder = inject(FormBuilder);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);

  user = signal<UserProfile | null>(null);
  profileForm: FormGroup;
  passwordForm: FormGroup;
  isEditMode = signal(false);
  isLoading = signal(false);

  availableLanguages = [
    { code: 'en', name: 'English' },
    { code: 'de', name: 'Deutsch' },
    { code: 'es', name: 'Español' },
    { code: 'fr', name: 'Français' }
  ];

  get currentUserLanguageName(): string {
    const currentUser = this.user();
    if (!currentUser?.language) return 'English';
    const lang = this.availableLanguages.find(l => l.code === currentUser.language);
    return lang?.name || 'English';
  }

  constructor() {
    this.profileForm = this.formBuilder.group({
      firstName: ['', [Validators.required]],
      lastName: ['', [Validators.required]],
      email: ['', [Validators.required, Validators.email]],
      language: ['']
    });

    this.passwordForm = this.formBuilder.group({
      currentPassword: ['', [Validators.required]],
      newPassword: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', [Validators.required]]
    }, { validators: this.passwordMatchValidator });
  }

  async ngOnInit() {
    await this.loadUserProfile();
  }

  private async loadUserProfile() {
    this.isLoading.set(true);
    try {
      this.userService.getCurrentUser().subscribe({
        next: (user) => {
          this.user.set(user);
          this.profileForm.patchValue({
            firstName: user.firstName,
            lastName: user.lastName,
            email: user.email,
            language: user.language || 'en'
          });
          this.isLoading.set(false);
        },
        error: (error) => {
          console.error('Failed to load user profile:', error);
          this.showError('Failed to load user profile');
          this.isLoading.set(false);
        }
      });
    } catch (error) {
      console.error('Failed to load user profile:', error);
      this.showError('Failed to load user profile');
      this.isLoading.set(false);
    }
  }

  toggleEditMode() {
    this.isEditMode.set(!this.isEditMode());
    if (!this.isEditMode()) {
      // Reset form when canceling
      const currentUser = this.user();
      if (currentUser) {
        this.profileForm.patchValue({
          firstName: currentUser.firstName,
          lastName: currentUser.lastName,
          email: currentUser.email,
          language: currentUser.language || 'en'
        });
      }
    }
  }

  async saveProfile() {
    if (this.profileForm.valid) {
      this.isLoading.set(true);
      const formValue = this.profileForm.value;
      
      try {
        this.userService.updateUserProfile(formValue).subscribe({
          next: (updatedUser) => {
            this.user.set(updatedUser);
            this.isEditMode.set(false);
            this.isLoading.set(false);
            this.showSuccess('Profile updated successfully');
            
            // Update application language if changed
            if (formValue.language) {
              this.translateService.use(formValue.language);
            }
          },
          error: (error) => {
            console.error('Failed to update profile:', error);
            this.showError('Failed to update profile');
            this.isLoading.set(false);
          }
        });
      } catch (error) {
        console.error('Failed to update profile:', error);
        this.showError('Failed to update profile');
        this.isLoading.set(false);
      }
    }
  }

  async changePassword() {
    if (this.passwordForm.valid) {
      this.isLoading.set(true);
      const { currentPassword, newPassword } = this.passwordForm.value;
      
      const passwordRequest: PasswordChangeRequest = {
        currentPassword,
        newPassword
      };

      try {
        this.userService.changePassword(passwordRequest).subscribe({
          next: () => {
            this.passwordForm.reset();
            this.isLoading.set(false);
            this.showSuccess('Password changed successfully');
          },
          error: (error) => {
            console.error('Failed to change password:', error);
            this.showError('Failed to change password');
            this.isLoading.set(false);
          }
        });
      } catch (error) {
        console.error('Failed to change password:', error);
        this.showError('Failed to change password');
        this.isLoading.set(false);
      }
    }
  }

  private passwordMatchValidator(form: any) {
    const newPassword = form.get('newPassword');
    const confirmPassword = form.get('confirmPassword');
    
    if (newPassword && confirmPassword && newPassword.value !== confirmPassword.value) {
      confirmPassword.setErrors({ passwordMismatch: true });
      return { passwordMismatch: true };
    }
    return null;
  }

  private showSuccess(message: string) {
    this.messageService.add({ 
      severity: 'success', 
      summary: 'Success', 
      detail: message,
      life: 3000
    });
  }

  private showError(message: string) {
    this.messageService.add({ 
      severity: 'error', 
      summary: 'Error', 
      detail: message,
      life: 5000
    });
  }
}
