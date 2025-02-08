import {inject, Pipe, PipeTransform} from '@angular/core';
import {TranslateService} from '@ngx-translate/core';

@Pipe({
  name: 'translateWeekday',
  standalone: true

})
export class TranslateWeekdayPipe implements PipeTransform {
    private readonly translateService = inject(TranslateService);

    transform(weekday: string) {
      return this.translateService.instant(`globals.weekdays.${weekday.toLowerCase()}`)
    }

}

