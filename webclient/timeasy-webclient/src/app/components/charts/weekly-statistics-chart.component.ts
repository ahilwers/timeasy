import {
  ChangeDetectorRef,
  Component,
  effect,
  Inject,
  inject,
  LOCALE_ID,
  OnInit,
  signal
} from '@angular/core';
import {UIChart} from 'primeng/chart';
import {WeeklyStatisticsService} from '../../services/weekly-statistcs.service';
import {TranslatePipe, TranslateService} from '@ngx-translate/core';
import {TooltipItem} from 'chart.js';
import {formatSecondsToReadableTime} from '../../utils/time-utils';

@Component({
  selector: 'app-weekly-statistics-chart',
  standalone: true,
  templateUrl: './weekly-statistics-chart.component.html',
  imports: [
    UIChart,
    TranslatePipe
  ],
  styleUrls: ['./weekly-statistics-chart.component.css']
})

export class WeeklyStatisticsChartComponent implements OnInit {

  private readonly weeklyStatisticsService = inject(WeeklyStatisticsService);
  private readonly translateService = inject(TranslateService)

  currentWeekNumber = this.weeklyStatisticsService.currentWeekNumber();
  weeklyStatistics = this.weeklyStatisticsService.weeklyStatistics();
  error = this.weeklyStatisticsService.error();

  selectedWeekNumber = signal<number>(0);
  selectedYear = signal<number>(new Date().getFullYear());

  data: any;
  options: any;
  timeSum: number = 0;

  constructor(private cd: ChangeDetectorRef, @Inject(LOCALE_ID) private locale: string) {
    this.weeklyStatisticsService.reset();
    effect(() => {
      if (this.currentWeekNumber()>0) {
        this.selectedWeekNumber.set(this.currentWeekNumber());
      }
    });
    effect(() => {
      if (this.selectedWeekNumber() > 0) {
        this.weeklyStatisticsService.load(this.selectedWeekNumber(), this.selectedYear(), undefined);
      }
    });
    effect(() => {
      if (this.weeklyStatistics()) {
        this.initChart();
      }
    });
  }


  ngOnInit() {
    this.weeklyStatisticsService.loadCurrentWeekNumber();
  }

  initChart() {
    const documentStyle = getComputedStyle(document.documentElement);
    const textColor = documentStyle.getPropertyValue('--p-text-color');
    const textColorSecondary = documentStyle.getPropertyValue('--p-text-muted-color');
    const surfaceBorder = documentStyle.getPropertyValue('--p-content-border-color');

    this.data = {
      labels: [
        this.translateService.instant('globals.weekdays.monday'),
        this.translateService.instant('globals.weekdays.tuesday'),
        this.translateService.instant('globals.weekdays.wednesday'),
        this.translateService.instant('globals.weekdays.thursday'),
        this.translateService.instant('globals.weekdays.friday'),
        this.translateService.instant('globals.weekdays.saturday'),
        this.translateService.instant('globals.weekdays.sunday')
      ],
      datasets: this.buildBarData()
    };

    this.options = {
      maintainAspectRatio: false,
      aspectRatio: 0.8,
      plugins: {
        tooltip: {
          mode: 'index',
          intersect: false,
          callbacks: {
            label: (context: TooltipItem<'bar'>) => {
              const seconds = context.raw as number;
              return `${context.dataset.label}: ${formatSecondsToReadableTime(seconds, this.locale)}`;
            }
          }
        },
        legend: {
          labels: {
            color: textColor
          }
        }
      },
      scales: {
        x: {
          stacked: true,
          ticks: {
            color: textColorSecondary
          },
          grid: {
            color: surfaceBorder,
            drawBorder: false
          }
        },
        y: {
          stacked: true,
          ticks: {
            color: textColorSecondary,
            callback: (value: string | number) => {
              return formatSecondsToReadableTime(Number(value), this.locale);
            }
          },
          grid: {
            color: surfaceBorder,
            drawBorder: false
          }
        }
      }
    };
    this.cd.markForCheck()
  }

  buildBarData () {
    this.timeSum = 0;
    const projectMap = new Map<string, { name: string; color: string; data: number[] }>();
    if (this.weeklyStatistics()) {
      const stats = this.weeklyStatistics();
      var dayIndex = 0;
      for (const day of stats!.days!) {
        if (day.timesPerProject) {
          for (const project of day.timesPerProject) {
            let projectData = projectMap.get(project.projectId);
            if (!projectData) {
              projectData = {name: project.projectName, color: project.projectColor, data: Array(7).fill(0)};
              projectMap.set(project.projectId, projectData);
            }
            projectData.data[dayIndex] = project.timeInSeconds;
            this.timeSum += project.timeInSeconds;
          }
        }
        dayIndex++;
      }
    }
    const sortedProjects = Array.from(projectMap.values()).sort((a, b) => a.name.localeCompare(b.name));
    const datasets = [];
    for (const project of sortedProjects) {
      datasets.push({
        type: 'bar',
        label: project.name,
        backgroundColor: project.color,
        data: project.data
      });
    }
    return datasets;
  }
}
