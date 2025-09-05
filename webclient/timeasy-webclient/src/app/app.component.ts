import {Component, inject} from '@angular/core';
import { RouterOutlet } from '@angular/router';
import {MenuComponent} from './components/menu/menu.component';
import {HeaderComponent} from './components/header/header.component';
import {FooterComponent} from './components/footer/footer.component';
import { LanguageService } from './services/language.service';

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

  constructor() {
    this.languageService.initializeLanguage();
  }
}
