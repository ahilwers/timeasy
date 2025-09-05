import { Injectable, inject } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import Keycloak from 'keycloak-js';

@Injectable({
  providedIn: 'root'
})
export class LanguageService {
  private readonly translateService = inject(TranslateService);
  private readonly keycloak = inject(Keycloak);
  private readonly supportedLanguages = ['en', 'de'];

  async initializeLanguage(): Promise<void> {
    this.translateService.addLangs(this.supportedLanguages);
    this.translateService.setDefaultLang('en');

    const language = await this.getPreferredLanguage();
    this.translateService.use(language);
  }

  private async getPreferredLanguage(): Promise<string> {
    // First try to get language from Keycloak user profile
    const keycloakLanguage = await this.getKeycloakUserLanguage();
    if (keycloakLanguage && this.supportedLanguages.includes(keycloakLanguage)) {
      return keycloakLanguage;
    }

    // Fallback to browser language
    const browserLanguage = this.getBrowserLanguage();
    return this.supportedLanguages.includes(browserLanguage) ? browserLanguage : 'en';
  }

  private async getKeycloakUserLanguage(): Promise<string | null> {
    try {
      if (!this.keycloak.authenticated) {
        return null;
      }

      const profile = await this.keycloak.loadUserProfile();
      return this.extractLanguageFromAttributes(profile.attributes);
    } catch (error) {
      console.warn('Failed to load Keycloak user profile for language detection:', error);
      return null;
    }
  }

  private extractLanguageFromAttributes(attributes: Record<string, unknown> | undefined): string | null {
    if (!attributes || !attributes['locale']) {
      return null;
    }

    const locale = attributes['locale'];
    let languageCode: string;

    if (Array.isArray(locale) && locale.length > 0) {
      languageCode = String(locale[0]);
    } else {
      languageCode = String(locale);
    }

    // Extract language code from locale (e.g., 'en-US' -> 'en')
    return this.getLanguageCode(languageCode);
  }

  private getBrowserLanguage(): string {
    return this.getLanguageCode(navigator.language);
  }

  private getLanguageCode(locale: string): string {
    return locale.split('-')[0];
  }
}