import { Pipe, PipeTransform, inject } from '@angular/core';
import { DateOnly } from '../models/date_only';
import { DynamicLocaleService } from '../services/dynamic-locale.service';

@Pipe({
  name: 'dateOnly',
  standalone: true,
  pure: false
})
export class DateOnlyPipe implements PipeTransform {
  private readonly dynamicLocaleService = inject(DynamicLocaleService);

  transform(value: DateOnly | null | undefined): string {
    if (!value || value.year === 0 || value.month === 0 || value.day === 0) {
      return '';
    }

    const date = value.toDate();
    const locale = this.dynamicLocaleService.getCurrentLocale();
    return new Intl.DateTimeFormat(locale, {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    }).format(date);
  }
}
