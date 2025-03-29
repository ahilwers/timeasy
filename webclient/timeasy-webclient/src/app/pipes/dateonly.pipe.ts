import { Pipe, PipeTransform } from '@angular/core';
import { DateOnly } from '../models/date_only';

@Pipe({
  name: 'dateOnly',
  standalone: true
})
export class DateOnlyPipe implements PipeTransform {
  transform(value: DateOnly | null | undefined): string {
    if (!value || value.year === 0 || value.month === 0 || value.day === 0) {
      return '';
    }

    const date = value.toDate();
    return new Intl.DateTimeFormat(navigator.language, {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    }).format(date);
  }
}
