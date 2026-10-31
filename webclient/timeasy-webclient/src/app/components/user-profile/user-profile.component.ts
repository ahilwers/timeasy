
import { Component, OnInit, inject, signal, ChangeDetectionStrategy } from '@angular/core';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import Keycloak from 'keycloak-js';
import { MessageService } from 'primeng/api';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { ProgressSpinnerModule } from 'primeng/progressspinner';
import { ToastModule } from 'primeng/toast';

import { Router } from '@angular/router';
import { UserProfile } from '../../interfaces/auth-provider.interface';
import { BillingService } from '../../services/billing.service';

@Component({
    selector: 'app-user-profile',
    standalone: true,
    imports: [
    CardModule,
    ButtonModule,
    ToastModule,
    ProgressSpinnerModule,
    TranslatePipe
],
    templateUrl: './user-profile.component.html',
    changeDetection: ChangeDetectionStrategy.Eager,
    styleUrl: './user-profile.component.css'
})
export class UserProfileComponent implements OnInit {
    private readonly messageService = inject(MessageService);
    private readonly keycloak = inject(Keycloak);
    private readonly translateService = inject(TranslateService);
    private readonly billingService = inject(BillingService);

    user = signal<UserProfile | null>(null);
    isLoading = signal(false);
    isPortalLoading = signal(false);
    /** Exposed for the template: billing UI (manage subscription) only shows when enabled. */
    readonly billingEnabled = this.billingService.billingEnabled;

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

    constructor(private router: Router) { }

    async ngOnInit() {
        await this.loadUserProfile();
        this.billingService.loadBillingStatus().subscribe();
    }

    openSubscriptionPortal() {
        if (!this.keycloak.authenticated) {
            this.showError('user.profile.errors.authenticationRequired');
            return;
        }
        this.isPortalLoading.set(true);
        this.billingService.getManagementURL().subscribe({
            next: (url) => {
                this.isPortalLoading.set(false);
                if (url) {
                    // Paddle customer portal for web purchases, App Store /
                    // Play Store management for mobile purchases.
                    window.location.href = url;
                } else {
                    // No active subscription yet — offer the upgrade flow.
                    this.router.navigate(['/upgrade']);
                }
            },
            error: () => {
                this.isPortalLoading.set(false);
                this.showError('license.portalError');
            }
        });
    }

    private async loadUserProfile() {
        this.isLoading.set(true);
        try {
            if (!this.keycloak.authenticated) {
                this.showError('user.profile.errors.authenticationRequired');
                this.isLoading.set(false);
                return;
            }

            const keycloakProfile = await this.keycloak.loadUserProfile();
            const userProfile: UserProfile = {
                id: this.keycloak.subject || '',
                username: keycloakProfile.username || '',
                email: keycloakProfile.email || '',
                firstName: keycloakProfile.firstName || '',
                lastName: keycloakProfile.lastName || '',
                displayName: `${keycloakProfile.firstName} ${keycloakProfile.lastName}`.trim() || keycloakProfile.username || '',
                language: this.extractLanguageFromAttributes(keycloakProfile.attributes) || 'en',
                attributes: this.convertAttributes(keycloakProfile.attributes)
            };

            this.user.set(userProfile);
            this.isLoading.set(false);
        } catch (error) {
            console.error('Failed to load user profile:', error);
            this.showError('user.profile.errors.failedToLoadProfile');
            this.isLoading.set(false);
        }
    }

    openAccountManagement() {
        try {
            if (!this.keycloak.authenticated) {
                this.showError('user.profile.errors.authenticationRequired');
                return;
            }
            this.keycloak.createLoginUrl().then(url => {
                const loginUrl = `${url}&kc_action=UPDATE_PROFILE`
                window.location.href = loginUrl;
            });
        }
        catch (error) {
            console.error('All methods failed:', error);
            this.showError('user.profile.errors.accountManagementUnavailable');
        }
    }

    openPasswordManagement() {
        try {
            if (!this.keycloak.authenticated) {
                this.showError('user.profile.errors.passwordAuthenticationRequired');
                return;
            }
            this.keycloak.createLoginUrl().then(url => {
                const changePasswordUrl = `${url}&kc_action=UPDATE_PASSWORD`
                window.location.href = changePasswordUrl;
            });
        }
        catch (error) {
            this.showError('user.profile.errors.passwordChangeUnavailable');
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

    private showError(messageKey: string) {
        const translatedMessage = this.translateService.instant(messageKey);
        const errorTitle = this.translateService.instant('globals.error');
        this.messageService.add({
            severity: 'error',
            summary: errorTitle,
            detail: translatedMessage,
            life: 5000
        });
    }

    private extractLanguageFromAttributes(attributes: Record<string, unknown> | undefined): string | undefined {
        if (!attributes || !attributes['locale']) {
            return undefined;
        }

        const locale = attributes['locale'];
        if (Array.isArray(locale) && locale.length > 0) {
            return String(locale[0]);
        }

        return String(locale);
    }

    private convertAttributes(attributes: Record<string, unknown> | undefined): { [key: string]: string[] } | undefined {
        if (!attributes) {
            return undefined;
        }

        const converted: { [key: string]: string[] } = {};

        Object.entries(attributes).forEach(([key, value]) => {
            if (Array.isArray(value)) {
                converted[key] = value.map(v => String(v));
            } else if (value !== undefined && value !== null) {
                converted[key] = [String(value)];
            }
        });

        return Object.keys(converted).length > 0 ? converted : undefined;
    }

}
