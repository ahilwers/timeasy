import { Pipe, PipeTransform, Inject, LOCALE_ID } from '@angular/core';

@Pipe({
  name: 'utcToLocalDate',
  standalone: true,
})
export class UtcToLocalDatePipe implements PipeTransform {
  constructor(@Inject(LOCALE_ID) private locale: string) {}

  transform(value: number | Date): string {
    let utcDate: Date;
    if (value instanceof Date) {
      utcDate = value;
    } else {
      utcDate = new Date(value*1000);
    }
    return utcDate.toLocaleDateString(this.locale, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    });
  }
}
