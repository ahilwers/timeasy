import {Component, effect, inject, OnInit, signal} from '@angular/core';
import {WeeklyStatisticsService} from '../../../services/weekly-statistcs.service';
import {TranslatePipe} from '@ngx-translate/core';
import {Toast} from 'primeng/toast';
import {TableModule} from 'primeng/table';
import {MessageService} from 'primeng/api';
import {SecondsToTimePipe} from '../../../pipes/seconds-to-time.pipe';
import {Button} from 'primeng/button';
import {TranslateWeekdayPipe} from '../../../pipes/translate-weekday.pipe';
import {UtcToLocalDatePipe} from '../../../pipes/utc-to-local-date.pipe';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {Select} from 'primeng/select';
import {ProjectService} from '../../../services/project.service';

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
  styleUrl: './weekly-overview.component.css'
})

export class WeeklyOverviewComponent implements OnInit
{
  private readonly weeklyStatisticsService = inject(WeeklyStatisticsService);
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
}
