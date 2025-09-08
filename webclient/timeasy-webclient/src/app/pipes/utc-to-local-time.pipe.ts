import { Pipe, PipeTransform, inject } from '@angular/core';
import { DynamicLocaleService } from '../services/dynamic-locale.service';

@Pipe({
  name: 'utcToLocalTime',
  standalone: true,
  pure: false
})
export class UtcToLocalTimePipe implements PipeTransform {
  private readonly dynamicLocaleService = inject(DynamicLocaleService);

  transform(value: Date | undefined): string {
    if (!value) {
      return '';
    }
    const locale = this.dynamicLocaleService.getCurrentLocale();
    return value.toLocaleTimeString(locale, {
      hour: '2-digit',
      minute: '2-digit',
    });
  }
}
