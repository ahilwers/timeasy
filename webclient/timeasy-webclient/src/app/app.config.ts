import {ApplicationConfig} from '@angular/core';
import { provideHttpClient } from '@angular/common/http';
import {appRoutes} from './app.routes';
import {AutoRefreshTokenService, provideKeycloak, UserActivityService, withAutoRefreshToken} from 'keycloak-angular';


export const provideKeycloakAngular = () =>
  provideKeycloak({
    config: {
      url: 'http://localhost:8180',
      realm: 'timeasy',
      clientId: 'timeasy-webclient'
    },
    initOptions: {
      onLoad: 'check-sso',
    },
    features: [
      withAutoRefreshToken({
        onInactivityTimeout: 'logout',
        sessionTimeout: 60000
      })
    ],
    providers: [AutoRefreshTokenService, UserActivityService]
  });


export const appConfig: ApplicationConfig = {
  providers: [
    appRoutes,
    provideHttpClient(),
    provideKeycloakAngular()
  ],
};
