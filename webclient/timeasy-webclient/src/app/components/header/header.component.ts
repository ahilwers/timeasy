import { Component, effect, inject, ViewChild } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { KEYCLOAK_EVENT_SIGNAL, KeycloakEventType, ReadyArgs, typeEventArgs } from 'keycloak-angular';
import Keycloak, { KeycloakProfile } from 'keycloak-js';
import { MenuItem } from 'primeng/api';
import { Button } from 'primeng/button';
import { Menu } from 'primeng/menu';
import { Popover } from 'primeng/popover';

@Component({
    selector: 'app-header',
    standalone: true,
    imports: [
        RouterLink,
        Button,
        Popover,
        Menu,
        TranslatePipe
    ],
    templateUrl: './header.component.html',
    styleUrl: './header.component.css'
})
export class HeaderComponent {
    @ViewChild('op') op!: Popover;

    authenticated: boolean = false;
    userProfile: KeycloakProfile = {};
    mobileMenuOpen: boolean = false;
    isAdmin: boolean = false;
    menuItems: MenuItem[] = [];

    private readonly keyCloak = inject(Keycloak);
    private readonly keyCloakSignal = inject(KEYCLOAK_EVENT_SIGNAL);
    private readonly translateService = inject(TranslateService);

    constructor() {
        effect(() => {
            const keycloakEvent = this.keyCloakSignal();
            if (keycloakEvent.type === KeycloakEventType.Ready) {
                this.authenticated = typeEventArgs<ReadyArgs>(keycloakEvent.args)
                if (this.authenticated) {
                    this.keyCloak.loadUserProfile().then(profile => {
                        console.log("logged in as user " + profile.username);
                        this.userProfile = profile;
                        this.isAdmin = this.keyCloak.hasRealmRole('admin');
                        this.updateMenu();
                    })
                }
            }
            if (keycloakEvent.type === KeycloakEventType.AuthLogout) {
                this.authenticated = false;
                this.updateMenu();
            }
        });

        // Initialize menu items
        this.updateMenu();
    }

    login() {
        this.keyCloak.login()
    }

    logout() {
        this.keyCloak.logout()
    }

    toggle({ event }: { event: any }) {
        this.op.toggle(event);
    }

    toggleMobileMenu() {
        this.mobileMenuOpen = !this.mobileMenuOpen;
        // Prevent scrolling when menu is open
        document.body.style.overflow = this.mobileMenuOpen ? 'hidden' : '';
    }

    closeMobileMenu() {
        this.mobileMenuOpen = false;
        document.body.style.overflow = '';
    }

    updateMenu() {
        this.menuItems = [
            {
                label: this.translateService.instant('menu.dashboard'),
                icon: 'pi pi-gauge',
                routerLink: '/dashboard'
            },
            {
                label: this.translateService.instant('menu.projects'),
                icon: 'pi pi-clipboard',
                routerLink: '/projects'
            },
            {
                label: this.translateService.instant('menu.weeklyOverview'),
                icon: 'pi pi-calendar',
                routerLink: '/statistics/weeklyoverview'
            },
            {
                label: this.translateService.instant('menu.timeEntries'),
                icon: 'pi pi-calendar-clock',
                routerLink: '/timeentries'
            }
        ];

        if (this.authenticated && this.isAdmin) {
            this.menuItems.push({
                label: this.translateService.instant('menu.admin'),
                icon: 'pi pi-shield',
                routerLink: '/admin'
            });
        }
    }
}
