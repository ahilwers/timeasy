import {Component, OnInit} from '@angular/core';
import Keycloak from 'keycloak-js';
import {User} from '../../models/user.model';

@Component({
  selector: 'app-user-profile',
  standalone: true,
  imports: [],
  templateUrl: './user-profile.component.html',
  styleUrl: './user-profile.component.css'
})
export class UserProfileComponent implements OnInit {

  user: User | undefined;

  constructor(private readonly keyCloak : Keycloak ) { }


  async ngOnInit() {
    if (this.keyCloak?.authenticated) {
      const profile = await this.keyCloak.loadUserProfile();
      this.user = {
        name: `${profile?.firstName} ${profile?.lastName}`,
        email: profile?.email ?? '',
        username: profile?.username ?? '',
      }
    }
  }

}
