import {signal} from '@angular/core';

export class WeeklyStatisticsState {
  readonly weeklyStatistics = signal<WeeklyStatistics | null>(null);
  readonly currentWeekNumber = signal<number | null>(null);
  readonly error = signal<string | null>(null);
}
