import { Component, ChangeDetectionStrategy } from '@angular/core';

@Component({
  selector: 'app-admin',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.Eager,
  template: '<h1>Admin</h1><p>You are successfully logged in!</p>',
})
export class AdminComponent {}
