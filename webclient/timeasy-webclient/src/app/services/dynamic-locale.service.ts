import { Injectable, inject, LOCALE_ID } from '@angular/core';
import { LocaleService } from './locale.service';

@Injectable({
  providedIn: 'root'
})
export class DynamicLocaleService {
  private readonly localeService = inject(LocaleService);
  private currentLocale: string = 'en-US';

  async initializeLocale(): Promise<void> {
    this.currentLocale = await this.localeService.getUserLocale();
  }

  getCurrentLocale(): string {
    return this.currentLocale;
  }

  setLocale(locale: string): void {
    this.currentLocale = locale;
    this.localeService.clearCache(); // Clear cache when manually setting locale
  }
}