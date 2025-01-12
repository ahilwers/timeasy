import {Component, effect, inject} from '@angular/core';
import Keycloak, {KeycloakProfile} from 'keycloak-js';
import {KEYCLOAK_EVENT_SIGNAL, KeycloakEventType, ReadyArgs, typeEventArgs} from 'keycloak-angular';

@Component({
  selector: 'app-footer',
  standalone: true,
  imports: [],
  templateUrl: './footer.component.html',
  styleUrl: './footer.component.css'
})

export class FooterComponent {
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

}
