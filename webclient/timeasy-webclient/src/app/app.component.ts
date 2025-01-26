import {Component, inject} from '@angular/core';
import { RouterOutlet } from '@angular/router';
import {MenuComponent} from './components/menu/menu.component';
import {HeaderComponent} from './components/header/header.component';
import {FooterComponent} from './components/footer/footer.component';
import {TranslateService} from '@ngx-translate/core';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [MenuComponent, RouterOutlet, HeaderComponent, FooterComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css'
})

export class AppComponent {
  title = 'timeasy-webclient';

  private readonly translateService = inject(TranslateService)
  private readonly supportedLanguages = ['en', 'de'];

  constructor() {
    const userLocale = this.getLAnguageCode(navigator.language);
    this.translateService.addLangs(this.supportedLanguages);
    this.translateService.setDefaultLang('en');
    this.translateService.use(this.supportedLanguages.includes(userLocale) ? userLocale : 'en');
  }

  /**
   * Returns the language code of the given locale
   * @param locale The locale
   * @returns The language code
   */
  private getLAnguageCode(locale: string): string {
    return locale.split('-')[0];
  }
}
