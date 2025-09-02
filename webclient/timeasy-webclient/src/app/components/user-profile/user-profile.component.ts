import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { TranslateModule } from '@ngx-translate/core';
import Keycloak from 'keycloak-js';
import { MessageService } from 'primeng/api';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { ProgressSpinnerModule } from 'primeng/progressspinner';
import { ToastModule } from 'primeng/toast';

import { UserProfile } from '../../interfaces/auth-provider.interface';
import { UserService } from '../../services/user.service';

@Component({
    selector: 'app-user-profile',
    standalone: true,
    imports: [
        CommonModule,
        CardModule,
        ButtonModule,
        ToastModule,
        ProgressSpinnerModule,
        TranslateModule
    ],
    templateUrl: './user-profile.component.html',
    styleUrl: './user-profile.component.css'
})
export class UserProfileComponent implements OnInit {
    private readonly userService = inject(UserService);
    private readonly messageService = inject(MessageService);
    private readonly keycloak = inject(Keycloak);

    user = signal<UserProfile | null>(null);
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

    constructor() {}

    async ngOnInit() {
        await this.loadUserProfile();
    }

    private async loadUserProfile() {
        this.isLoading.set(true);
        try {
            this.userService.getCurrentUser().subscribe({
                next: (user) => {
                    this.user.set(user);
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

    openAccountManagement() {
        try {
            if (!this.keycloak.authenticated) {
                this.showError('You must be logged in to access account management.');
                return;
            }
            this.keycloak.accountManagement();
        }
        catch (error) {
            console.error('All methods failed:', error);
            this.showError('Unable to access account management. Please contact support.');
        }
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
