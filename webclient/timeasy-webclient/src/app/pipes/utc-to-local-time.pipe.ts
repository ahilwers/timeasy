import { Pipe, PipeTransform, Inject, LOCALE_ID } from '@angular/core';

@Pipe({
  name: 'utcToLocalTime',
  standalone: true,
})
export class UtcToLocalTimePipe implements PipeTransform {
  constructor(@Inject(LOCALE_ID) private locale: string) {}

  transform(value: Date | undefined): string {
    if (!value) {
      return '';
    }
    return value.toLocaleTimeString(this.locale, {
      hour: '2-digit',
      minute: '2-digit',
    });
  }
}
