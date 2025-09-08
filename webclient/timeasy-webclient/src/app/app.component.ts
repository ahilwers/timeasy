import {Component, inject} from '@angular/core';
import { RouterOutlet } from '@angular/router';
import {MenuComponent} from './components/menu/menu.component';
import {HeaderComponent} from './components/header/header.component';
import {FooterComponent} from './components/footer/footer.component';
import { LanguageService } from './services/language.service';
import { DynamicLocaleService } from './services/dynamic-locale.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [MenuComponent, RouterOutlet, HeaderComponent, FooterComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css'
})

export class AppComponent {
  title = 'timeasy-webclient';

  private readonly languageService = inject(LanguageService);
  private readonly dynamicLocaleService = inject(DynamicLocaleService);

  constructor() {
    this.initializeServices();
  }

  private async initializeServices(): Promise<void> {
    await this.languageService.initializeLanguage();
    await this.dynamicLocaleService.initializeLocale();
  }
}
