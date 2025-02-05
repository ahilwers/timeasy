import {WeeklyStatisticsState} from './weekly-statistics.state';
import {inject} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {environment} from '../../environments/environment';
import {TranslateService} from '@ngx-translate/core';
import {WeekNumber} from '../models/week-number.model';

export class WeeklyStatisticsService {

  private readonly apiUrl = `${environment.apiUrl}/weeklystatistics`;
  private readonly state = new WeeklyStatisticsState();

  private readonly http = inject(HttpClient);
  private readonly translateService = inject(TranslateService);

  load(weekNumber: number, year: number): void {
    this.state.error.set(null);
    this.http.get<WeeklyStatistics>(`${this.apiUrl}/${weekNumber}/${year}`).subscribe({
      next: (statistics) => {
        this.state.weeklyStatistics.set(statistics);
      },
      error: (err) => {
        console.error(`Failed to loading weekly statistics for week ${weekNumber}/${year}:`, err);
        this.state.error.set(this.translateService.instant('weeklyStatistics.errorLoadingWeeklyStatistics'));
      }
    });
  }

  loadCurrentWeekNumber(): void {
    this.state.error.set(null);
    this.http.get<WeekNumber>(`${this.apiUrl}/currentweeknumber`).subscribe({
      next: (weekNumber) => {
        this.state.currentWeekNumber.set(weekNumber.weekNumber);
      },
      error: (err) => {
        console.error('Failed to load current week number:', err);
        this.state.error.set(this.translateService.instant('weeklyStatistics.errorLoadingWeekNumber'));
      }
    });
  }
}
