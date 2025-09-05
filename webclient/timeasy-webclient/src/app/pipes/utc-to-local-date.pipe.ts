import { Pipe, PipeTransform, inject } from '@angular/core';
import { DynamicLocaleService } from '../services/dynamic-locale.service';

@Pipe({
  name: 'utcToLocalDate',
  standalone: true,
  pure: false
})
export class UtcToLocalDatePipe implements PipeTransform {
  private readonly dynamicLocaleService = inject(DynamicLocaleService);

  transform(value: Date | undefined): string {
    if (!value) {
      return '';
    }
    const locale = this.dynamicLocaleService.getCurrentLocale();
    return value.toLocaleDateString(locale, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    });
  }
}
