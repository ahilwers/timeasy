import {Component, effect, inject} from '@angular/core';
import {RouterLink} from '@angular/router';
import Keycloak, {KeycloakProfile} from 'keycloak-js';
import {KEYCLOAK_EVENT_SIGNAL, KeycloakEventType, ReadyArgs, typeEventArgs} from 'keycloak-angular';

@Component({
  selector: 'app-menu',
  standalone: true,
  imports: [RouterLink],
  templateUrl: './menu.component.html',
  styleUrl: './menu.component.css'
})
export class MenuComponent {

  authenticated : boolean = false;
  userProfile : KeycloakProfile = {};
  isAdmin : boolean = false;
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
            this.isAdmin = this.keyCloak.hasRealmRole('admin')
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


}
