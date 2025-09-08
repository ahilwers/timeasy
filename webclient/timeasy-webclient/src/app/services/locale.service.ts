import { Injectable, inject } from '@angular/core';
import Keycloak from 'keycloak-js';

@Injectable({
  providedIn: 'root'
})
export class LocaleService {
  private readonly keycloak = inject(Keycloak);
  private readonly supportedLocales = ['en-US', 'de-DE'];
  private cachedLocale: string | null = null;

  async getUserLocale(): Promise<string> {
    if (this.cachedLocale) {
      return this.cachedLocale;
    }

    // First try to get locale from Keycloak user profile
    const keycloakLocale = await this.getKeycloakUserLocale();
    if (keycloakLocale && this.supportedLocales.includes(keycloakLocale)) {
      this.cachedLocale = keycloakLocale;
      return keycloakLocale;
    }

    // Fallback to browser locale
    const browserLocale = this.getBrowserLocale();
    const supportedBrowserLocale = this.supportedLocales.includes(browserLocale) ? browserLocale : 'en-US';
    this.cachedLocale = supportedBrowserLocale;
    return supportedBrowserLocale;
  }

  private async getKeycloakUserLocale(): Promise<string | null> {
    try {
      if (!this.keycloak.authenticated) {
        return null;
      }

      const profile = await this.keycloak.loadUserProfile();
      return this.extractLocaleFromAttributes(profile.attributes);
    } catch (error) {
      console.warn('Failed to load Keycloak user profile for locale detection:', error);
      return null;
    }
  }

  private extractLocaleFromAttributes(attributes: Record<string, unknown> | undefined): string | null {
    if (!attributes || !attributes['locale']) {
      return null;
    }

    const locale = attributes['locale'];
    let localeCode: string;

    if (Array.isArray(locale) && locale.length > 0) {
      localeCode = String(locale[0]);
    } else {
      localeCode = String(locale);
    }

    // Map language codes to full locales if needed
    return this.mapToSupportedLocale(localeCode);
  }

  private getBrowserLocale(): string {
    return navigator.language;
  }

  private mapToSupportedLocale(locale: string): string {
    // If it's already a full locale, return as is
    if (this.supportedLocales.includes(locale)) {
      return locale;
    }

    // Extract language code and map to default locale
    const languageCode = locale.split('-')[0];
    switch (languageCode) {
      case 'de':
        return 'de-DE';
      case 'en':
      default:
        return 'en-US';
    }
  }

  clearCache(): void {
    this.cachedLocale = null;
  }
}