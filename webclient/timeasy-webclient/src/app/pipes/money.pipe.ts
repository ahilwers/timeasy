import { Pipe, PipeTransform, inject } from '@angular/core';
import { DynamicLocaleService } from '../services/dynamic-locale.service';

@Pipe({
  name: 'money',
  standalone: true,
  pure: false
})
export class MoneyPipe implements PipeTransform {
  private readonly dynamicLocaleService = inject(DynamicLocaleService);

  transform(value: number | string | null | undefined): string {
    if (value === null || value === undefined || value === '') {
      return '';
    }

    const amount = typeof value === 'string' ? parseFloat(value) : value;
    const locale = this.dynamicLocaleService.getCurrentLocale();

    return new Intl.NumberFormat(locale, {
      style: 'decimal',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(amount);
  }
}
