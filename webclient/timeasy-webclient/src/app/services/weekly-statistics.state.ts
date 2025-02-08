import {signal} from '@angular/core';
import {WeeklyStatistics} from '../models/weekly-statistics.model';

export class WeeklyStatisticsState {
  readonly weeklyStatistics = signal<WeeklyStatistics | null>(null);
  readonly currentWeekNumber = signal<number>(0);
  readonly error = signal<string | null>(null);
}
