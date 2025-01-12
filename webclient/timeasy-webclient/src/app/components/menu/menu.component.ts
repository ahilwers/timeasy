import {Component, effect, inject, OnInit} from '@angular/core';
import Keycloak, {KeycloakProfile} from 'keycloak-js';
import {KEYCLOAK_EVENT_SIGNAL, KeycloakEventType, ReadyArgs, typeEventArgs} from 'keycloak-angular';
import {Button} from 'primeng/button';
import {Menu} from 'primeng/menu';
import {MenuItem} from 'primeng/api';

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
        label: 'Home',
        icon: 'pi pi-home',
        routerLink: '/'
      },
      {
        label: 'Dashboard',
        icon: 'pi pi-gauge',
        routerLink: '/dashboard'
      },
      {
        label: 'Project',
        icon: 'pi pi-clipboard',
        routerLink: '/projects'
      },
      {
        label: 'Time Entries',
        icon: 'pi pi-calendar-clock',
        routerLink: '/timeentries'
      }
    ];

    if (this.authenticated && this.isAdmin) {
      this.menuItems.push({
        label: 'Admin',
        icon: 'pi pi-shield',
        routerLink: '/admin'
      });
    }

  }
}
