import {ApplicationConfig, LOCALE_ID} from '@angular/core';
import {HttpClient, provideHttpClient, withInterceptors} from '@angular/common/http';
import {appRoutes} from './app.routes';
import {
  AutoRefreshTokenService,
  createInterceptorCondition,
  INCLUDE_BEARER_TOKEN_INTERCEPTOR_CONFIG,
  IncludeBearerTokenCondition,
  includeBearerTokenInterceptor,
  provideKeycloak,
  UserActivityService,
  withAutoRefreshToken
} from 'keycloak-angular';
import {provideAnimationsAsync} from '@angular/platform-browser/animations/async';
import {providePrimeNG} from 'primeng/config';
import {ColorPreset} from './color.preset';
import localeDe from '@angular/common/locales/de';
import {registerLocaleData} from '@angular/common';
import {provideTranslateService, TranslateLoader} from '@ngx-translate/core';
import {TranslateHttpLoader} from '@ngx-translate/http-loader';
import { environment } from '../environments/environment';

export const provideKeycloakAngular = () =>
  provideKeycloak({
    config: {
      url: environment.authUrl,
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

const apiUrl = environment.apiUrl.replace(/\/$/, '');
const urlCondition = createInterceptorCondition<IncludeBearerTokenCondition>({
  urlPattern: new RegExp(`^${apiUrl.replace(/[-/\\^$*+?.()|[\]{}]/g, '\\$&')}(\\/.*)?$`, 'i'),
  bearerPrefix: 'Bearer'
});

registerLocaleData(localeDe);
const supportedLocales = ['en-US', 'de-DE'];
const userLocale = navigator.language;
const locale = supportedLocales.includes(userLocale) ? userLocale : 'en-US';

const httpLoaderFactory: (http: HttpClient) => TranslateHttpLoader = (http: HttpClient) =>
  new TranslateHttpLoader(http, './i18n/', '.json');


export const appConfig: ApplicationConfig = {
  providers: [
    appRoutes,
    {
      provide: INCLUDE_BEARER_TOKEN_INTERCEPTOR_CONFIG,
      useValue: [urlCondition]
    },
    provideHttpClient(withInterceptors([includeBearerTokenInterceptor])),
    provideKeycloakAngular(),
    provideAnimationsAsync(),
    providePrimeNG({
      theme: {
        preset: ColorPreset
      },
      ripple: true
    }),
    { provide: LOCALE_ID, useValue: locale },
    provideHttpClient(),
    provideTranslateService({
      loader: {
        provide: TranslateLoader,
        useFactory: httpLoaderFactory,
        deps: [HttpClient],
      },
    })
  ],
};
