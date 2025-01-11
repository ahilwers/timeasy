import {provideRouter, Routes} from '@angular/router';
import {canActivateAuth} from './guards/auth.guard';

const routes: Routes = [
  { path: '', loadComponent: () => import('./components/home/home.component').then(m => m.HomeComponent) },
  { path: 'forbidden', loadComponent: () => import('./components/forbidden/forbidden.component').then(m => m.ForbiddenComponent) },
  {
    path: 'dashboard',
    loadComponent: () => import('./components/dashboard/dashboard.component').then(m => m.DashboardComponent),
    canActivate: [canActivateAuth],
  },
  {
    path: 'admin',
    loadComponent: () => import('./components/admin/admin.component').then(m => m.AdminComponent),
    canActivate: [canActivateAuth],
    data: { role: 'admin' }
  },
  {
    path: 'user',
    loadComponent: () => import('./components/user-profile/user-profile.component').then(m => m.UserProfileComponent),
    canActivate: [canActivateAuth],
  },
  { path: '**', redirectTo: '' },
];

export const appRoutes = provideRouter(routes);


