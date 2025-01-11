import {Component, inject} from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [RouterLink],
  template: `
    <h1>Welcome to the Home Page</h1>

    <p>Welcome :)</p>
  `,
})

export class HomeComponent {
}

