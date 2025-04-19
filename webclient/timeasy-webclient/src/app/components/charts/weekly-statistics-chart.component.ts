import {ChangeDetectorRef, Component, effect, inject, OnInit, PLATFORM_ID, signal} from '@angular/core';
import {isPlatformBrowser} from '@angular/common';
import {UIChart} from 'primeng/chart';
import {WeeklyStatisticsService} from '../../services/weekly-statistcs.service';

@Component({
  selector: 'app-weekly-statistics-chart',
  standalone: true,
  templateUrl: './weekly-statistics-chart.component.html',
  imports: [
    UIChart
  ],
  styleUrls: ['./weekly-statistics-chart.component.css']
})

export class WeeklyStatisticsChartComponent implements OnInit {

  private readonly weeklyStatisticsService = inject(WeeklyStatisticsService);

  currentWeekNumber = this.weeklyStatisticsService.currentWeekNumber();
  weeklyStatistics = this.weeklyStatisticsService.weeklyStatistics();
  error = this.weeklyStatisticsService.error();

  selectedWeekNumber = signal<number>(0);
  selectedYear = signal<number>(new Date().getFullYear());

  data: any;

  options: any;

  platformId = inject(PLATFORM_ID);

  constructor(private cd: ChangeDetectorRef) {
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
      const textColor = documentStyle.getPropertyValue('--p-text-color');
      const textColorSecondary = documentStyle.getPropertyValue('--p-text-muted-color');
      const surfaceBorder = documentStyle.getPropertyValue('--p-content-border-color');

      this.data = {
        labels: ['Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Freitag', 'Samstag', 'Sonntag'],
        datasets: this.buildBarData()
      };

      this.options = {
        maintainAspectRatio: false,
        aspectRatio: 0.8,
        plugins: {
          tooltip: {
            mode: 'index',
            intersect: false
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
              color: textColorSecondary
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
  }

  buildBarData () {
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
