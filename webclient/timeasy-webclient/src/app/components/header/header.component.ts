import {Component, effect, inject, ViewChild} from '@angular/core';
import {Button} from 'primeng/button';
import {Popover} from 'primeng/popover';
import Keycloak, {KeycloakProfile} from 'keycloak-js';
import {KEYCLOAK_EVENT_SIGNAL, KeycloakEventType, ReadyArgs, typeEventArgs} from 'keycloak-angular';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [
    Button,
    Popover
  ],
  templateUrl: './header.component.html',
  styleUrl: './header.component.css'
})
export class HeaderComponent {
  @ViewChild('op') op!: Popover;

  authenticated : boolean = false;
  userProfile : KeycloakProfile = {};

  private readonly keyCloak = inject(Keycloak);
  private readonly keyCloakSignal = inject(KEYCLOAK_EVENT_SIGNAL);

  constructor() {
    effect(() => {
      const keycloakEvent = this.keyCloakSignal();
      if (keycloakEvent.type === KeycloakEventType.Ready) {
        this.authenticated = typeEventArgs<ReadyArgs>(keycloakEvent.args)
        if (this.authenticated) {
          this.keyCloak.loadUserProfile().then(profile => {
            console.log("logged in as user "+profile.username);
            this.userProfile = profile
          })
        }
      }
      if (keycloakEvent.type === KeycloakEventType.AuthLogout) {
        this.authenticated = false
      }
    });
  }

  login() {
    this.keyCloak.login()
  }

  logout() {
    this.keyCloak.logout()
  }

  toggle({event}: { event: any }) {
    this.op.toggle(event);
  }
}
