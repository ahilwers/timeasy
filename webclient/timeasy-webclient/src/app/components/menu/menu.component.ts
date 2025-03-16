import {Component, effect, inject, OnInit} from '@angular/core';
import Keycloak, {KeycloakProfile} from 'keycloak-js';
import {KEYCLOAK_EVENT_SIGNAL, KeycloakEventType, ReadyArgs, typeEventArgs} from 'keycloak-angular';
import {Button} from 'primeng/button';
import {Menu} from 'primeng/menu';
import {MenuItem} from 'primeng/api';
import {TranslateService} from '@ngx-translate/core';

@Component({
  selector: 'app-menu',
  standalone: true,
  imports: [Button, Menu],
  templateUrl: './menu.component.html',
  styleUrl: './menu.component.css'
})

export class MenuComponent implements OnInit {
  authenticated : boolean = false;
  userProfile : KeycloakProfile = {};
  isAdmin : boolean = false;
  private readonly keyCloak = inject(Keycloak);
  private readonly keyCloakSignal = inject(KEYCLOAK_EVENT_SIGNAL);
  private readonly translateService = inject(TranslateService)

  menuItems : MenuItem[] = [];

  constructor() {
    effect(() => {
      const keycloakEvent = this.keyCloakSignal();
      if (keycloakEvent.type === KeycloakEventType.Ready) {
        this.authenticated = typeEventArgs<ReadyArgs>(keycloakEvent.args)
        if (this.authenticated) {
          this.keyCloak.loadUserProfile().then(profile => {
            console.log("logged in as user "+profile.username);
            this.userProfile = profile
            this.isAdmin = this.keyCloak.hasRealmRole('admin')
            this.updateMenu();
          })
        }
      }
      if (keycloakEvent.type === KeycloakEventType.AuthLogout) {
        this.authenticated = false
        this.updateMenu();
      }
    });
  }

  ngOnInit(): void {
    this.updateMenu();
  }

  login() {
    this.keyCloak.login()
  }

  logout() {
    this.keyCloak.logout()
  }

  updateMenu() {
    this.menuItems = [
      {
        label: this.translateService.instant('menu.home'),
        icon: 'pi pi-home',
        routerLink: '/'
      },
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
