import {Component, effect, inject, OnInit, signal, ChangeDetectionStrategy} from '@angular/core';
import {Router} from '@angular/router';
import {WeeklyStatisticsService} from '../../../services/weekly-statistcs.service';
import {TimeEntryService} from '../../../services/time-entry.service';
import {TranslatePipe} from '@ngx-translate/core';
import {Toast} from '@openng/optimus-ui/toast';
import {TableModule} from '@openng/optimus-ui/table';
import {MessageService} from '@openng/optimus-ui/api';
import {SecondsToTimePipe} from '../../../pipes/seconds-to-time.pipe';
import {Button} from '@openng/optimus-ui/button';
import {TranslateWeekdayPipe} from '../../../pipes/translate-weekday.pipe';
import {UtcToLocalDatePipe} from '../../../pipes/utc-to-local-date.pipe';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {Select} from '@openng/optimus-ui/select';
import {ProjectService} from '../../../services/project.service';
import {DailyStatistics} from '../../../models/weekly-statistics.model';

@Component({
  selector: 'app-weekly-overview',
  standalone: true,
  imports: [
    TranslatePipe,
    Toast,
    TableModule,
    SecondsToTimePipe,
    Button,
    TranslateWeekdayPipe,
    UtcToLocalDatePipe,
    FormsModule,
    ReactiveFormsModule,
    Select
  ],
  providers: [MessageService],
  templateUrl: './weekly-overview.component.html',
  changeDetection: ChangeDetectionStrategy.Eager,
  styleUrl: './weekly-overview.component.css'
})

export class WeeklyOverviewComponent implements OnInit
{
  private readonly weeklyStatisticsService = inject(WeeklyStatisticsService);
  private readonly timeEntryService = inject(TimeEntryService);
  private readonly router = inject(Router);
  private readonly projectService = inject(ProjectService);

  currentWeekNumber = this.weeklyStatisticsService.currentWeekNumber();
  weeklyStatistics = this.weeklyStatisticsService.weeklyStatistics();
  error = this.weeklyStatisticsService.error();
  projects = this.projectService.projects();

  selectedWeekNumber = signal<number>(0);
  selectedYear = signal<number>(new Date().getFullYear());
  selectedProjectId = signal<string | undefined>(undefined);

  constructor() {
    effect(() => {
      if (this.currentWeekNumber()>0) {
        this.selectedWeekNumber.set(this.currentWeekNumber());
      }
    });
    effect(() => {
      if (this.selectedWeekNumber() > 0) {
        this.weeklyStatisticsService.load(this.selectedWeekNumber(), this.selectedYear(), this.selectedProjectId());
      }
    });
    effect(() => {
      if (this.selectedProjectId()) {
        this.weeklyStatisticsService.load(this.selectedWeekNumber(), this.selectedYear(), this.selectedProjectId());
      }
    });
  }

  ngOnInit(): void {
    this.projectService.loadProjects();
    this.weeklyStatisticsService.loadCurrentWeekNumber();
  }

  previousWeek() {
    let week = this.selectedWeekNumber()-1;
    if (week==0) {
      week = 52;
      this.selectedYear.set(this.selectedYear()-1);
    }
    this.selectedWeekNumber.set(week)
  }

  nextWeek() {
    let week = this.selectedWeekNumber()+1;
    if (week==53) {
      week = 1;
      this.selectedYear.set(this.selectedYear()+1);
    }
    this.selectedWeekNumber.set(week)
  }

  projectSelected(event: any) {
    if (event.value) {
      this.selectedProjectId.set(event.value.id);
    } else {
      this.selectedProjectId.set(undefined);
    }

  }

  openDay(day: DailyStatistics) {
    const index = this.weeklyStatistics()!.days.indexOf(day);
    if (index < 0) {
      return;
    }
    const date = this.getDateForDayIndex(index);
    // Reuse the time entry list with a single day filter. The project filter
    // is synced as well so the list shows the same data as the weekly view:
    this.timeEntryService.selectDateRange([date, date]);
    this.timeEntryService.selectProject(this.selectedProjectId());
    this.router.navigate(['/timeentries']);
  }

  getDateForDayIndex(index: number): Date {
    // The server sends the days starting with Monday and the first day of the
    // week as UTC midnight. Derive the calendar day in UTC so it cannot be
    // shifted by the browser's timezone, then build a local date from it that
    // matches the local date formatting used for the time entry filter:
    const day = new Date(this.weeklyStatistics()!.firstDay);
    day.setUTCDate(day.getUTCDate() + index);
    return new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate());
  }
}
