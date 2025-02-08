import {WeeklyStatisticsState} from './weekly-statistics.state';
import {computed, inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {environment} from '../../environments/environment';
import {TranslateService} from '@ngx-translate/core';
import {WeekNumber} from '../models/week-number.model';
import {WeeklyStatistics} from '../models/weekly-statistics.model';

@Injectable({
  providedIn: 'root'
})

export class WeeklyStatisticsService {

  private readonly apiUrl = `${environment.apiUrl}`;
  private readonly state = new WeeklyStatisticsState();

  private readonly http = inject(HttpClient);
  private readonly translateService = inject(TranslateService);

  weeklyStatistics = computed(() => this.state.weeklyStatistics);
  currentWeekNumber = computed(() => this.state.currentWeekNumber);
  error = computed(() => this.state.error);

  load(weekNumber: number, year: number): void {
    console.log("loading statistics")
    this.state.error.set(null);
    this.http.get<WeeklyStatistics>(`${this.apiUrl}/weeklystatistics/${weekNumber}/${year}`).subscribe({
      next: (statistics) => {
        console.log('Received statistics:', statistics);
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
