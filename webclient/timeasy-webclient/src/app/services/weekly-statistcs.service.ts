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

  load(weekNumber: number, year: number, projectId: string | undefined): void {
    console.log("loading statistics")
    this.state.error.set(null);
    let url = `${this.apiUrl}/weeklystatistics/${weekNumber}/${year}`;
    if (projectId) {
      url += `?project=${projectId}`;
    }
    this.http.get<WeeklyStatistics>(url).subscribe({
      next: (statistics) => {
        const transformedStatistics = {
          ...statistics,
          firstDay: new Date(statistics.firstDay),
          lastDay: new Date(statistics.lastDay)
        };
        console.log('Received statistics:', transformedStatistics);
        this.state.weeklyStatistics.set(transformedStatistics);
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

  reset(): void {
    this.state.weeklyStatistics.set(null);
    this.state.currentWeekNumber.set(0);
    this.state.error.set(null);
  }
}
