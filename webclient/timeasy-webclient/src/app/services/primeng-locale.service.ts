import { Injectable, inject } from '@angular/core';
import { PrimeNG } from 'primeng/config';
import { DynamicLocaleService } from './dynamic-locale.service';

@Injectable({
  providedIn: 'root'
})
export class PrimeNGLocaleService {
  private readonly primeConfig = inject(PrimeNG);
  private readonly dynamicLocaleService = inject(DynamicLocaleService);

  setLocaleForPrimeNG(): void {
    const userLocale = this.dynamicLocaleService.getCurrentLocale();
    
    if (userLocale === 'de-DE') {
      this.primeConfig.setTranslation({
        dateFormat: 'dd.mm.yy',
        firstDayOfWeek: 1,
        dayNames: ['Sonntag', 'Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Freitag', 'Samstag'],
        dayNamesShort: ['Son', 'Mon', 'Die', 'Mit', 'Don', 'Fre', 'Sam'],
        dayNamesMin: ['So', 'Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa'],
        monthNames: ['Januar', 'Februar', 'März', 'April', 'Mai', 'Juni', 'Juli', 'August', 'September', 'Oktober', 'November', 'Dezember'],
        monthNamesShort: ['Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun', 'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez'],
        today: 'Heute',
        clear: 'Löschen'
      });
    } else {
      this.primeConfig.setTranslation({
        dateFormat: 'mm/dd/yy',
        firstDayOfWeek: 0,
        dayNames: ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'],
        dayNamesShort: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'],
        dayNamesMin: ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'],
        monthNames: ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'],
        monthNamesShort: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],
        today: 'Today',
        clear: 'Clear'
      });
    }
  }
}