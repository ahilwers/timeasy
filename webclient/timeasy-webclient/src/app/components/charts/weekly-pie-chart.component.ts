import {ChangeDetectorRef, Component, effect, Inject, inject, LOCALE_ID, PLATFORM_ID, signal} from '@angular/core';
import {ChartModule} from 'primeng/chart';
import {isPlatformBrowser} from '@angular/common';
import {WeeklyStatisticsService} from '../../services/weekly-statistcs.service';
import {TooltipItem} from 'chart.js';
import {formatSecondsToReadableTime} from '../../utils/time-utils';

@Component({
  selector: 'app-weekly-pie-chart',
  standalone: true,
  templateUrl: './weekly-pie-chart.component.html',
  imports: [
    ChartModule,
  ],
  styleUrls: ['./weekly-pie-chart.component.css']
})

export class WeeklyPieChartComponent {

  private readonly weeklyStatisticsService = inject(WeeklyStatisticsService);

  currentWeekNumber = this.weeklyStatisticsService.currentWeekNumber();
  weeklyStatistics = this.weeklyStatisticsService.weeklyStatistics();
  error = this.weeklyStatisticsService.error();

  selectedWeekNumber = signal<number>(0);
  selectedYear = signal<number>(new Date().getFullYear());
  data: any;
  options: any;
  platformId = inject(PLATFORM_ID);

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
    if (isPlatformBrowser(this.platformId)) {
      const documentStyle = getComputedStyle(document.documentElement);
      const textColor = documentStyle.getPropertyValue('--text-color');
      this.data = this.buildChartData();

      this.options = {
        plugins: {
          legend: {
            labels: {
              usePointStyle: true,
              color: textColor
            }
          },
          tooltip: {
            callbacks: {
              label: (context: TooltipItem<'bar'>) => {
                const seconds = context.raw as number;
                const formatted = formatSecondsToReadableTime(seconds, this.locale);
                return `${context.label}: ${formatted}`;
              }
            }
          }
        }
      };
      this.cd.markForCheck()
    }

  }

  buildChartData () {
    const projectMap = new Map<string, { name: string; color: string; data: number }>();
    if (this.weeklyStatistics()) {
      const stats = this.weeklyStatistics();
      for (const day of stats!.days!) {
        if (day.timesPerProject) {
          for (const project of day.timesPerProject) {
            let projectData = projectMap.get(project.projectId);
            if (!projectData) {
              projectData = {name: project.projectName, color: project.projectColor, data: 0};
              projectMap.set(project.projectId, projectData);
            }
            projectData.data += project.timeInSeconds;
          }
        }
      }
    }
    const sortedProjects = Array.from(projectMap.values()).sort((a, b) => a.name.localeCompare(b.name));

    const labels = sortedProjects.map(p => p.name);
    const data = sortedProjects.map(p => p.data);
    const backgroundColor = sortedProjects.map(p => p.color);
    const hoverBackgroundColor = backgroundColor.map(c => this.lightenColor(c, 20)); // 20% heller machen

    return {
      labels,
      datasets: [
        {
          data,
          backgroundColor,
          hoverBackgroundColor
        }
      ]
    };
  }

  lightenColor(color: string, percent: number): string {
    let r, g, b;

    if (color.startsWith('#')) {
      const bigint = parseInt(color.slice(1), 16);
      r = (bigint >> 16) & 255;
      g = (bigint >> 8) & 255;
      b = bigint & 255;
    } else if (color.startsWith('rgb')) {
      const result = color.match(/\d+/g);
      if (!result) return color;
      [r, g, b] = result.map(Number);
    } else {
      return color; // fallback
    }

    r = Math.min(255, Math.floor(r + (255 - r) * (percent / 100)));
    g = Math.min(255, Math.floor(g + (255 - g) * (percent / 100)));
    b = Math.min(255, Math.floor(b + (255 - b) * (percent / 100)));

    return `rgb(${r}, ${g}, ${b})`;
  }

}
