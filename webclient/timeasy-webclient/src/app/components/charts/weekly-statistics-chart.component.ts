import {ChangeDetectorRef, Component, inject, OnInit, PLATFORM_ID} from '@angular/core';
import {isPlatformBrowser} from '@angular/common';
import {UIChart} from 'primeng/chart';

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
  data: any;

  options: any;

  platformId = inject(PLATFORM_ID);

  constructor(private cd: ChangeDetectorRef) {
  }


  ngOnInit() {
    this.initChart();
  }

  initChart() {
    if (isPlatformBrowser(this.platformId)) {
      const documentStyle = getComputedStyle(document.documentElement);
      const textColor = documentStyle.getPropertyValue('--p-text-color');
      const textColorSecondary = documentStyle.getPropertyValue('--p-text-muted-color');
      const surfaceBorder = documentStyle.getPropertyValue('--p-content-border-color');

      this.data = {
        labels: ['Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Freitag', 'Samstag', 'Sonntag'],
        datasets: [
          {
            type: 'bar',
            label: 'Projekt 1',
            backgroundColor: documentStyle.getPropertyValue('--p-cyan-500'),
            data: [50, 25, 12, 48, 90, 76, 42]
          },
          {
            type: 'bar',
            label: 'Projekt 2',
            backgroundColor: documentStyle.getPropertyValue('--p-gray-500'),
            data: [21, 84, 24, 75, 37, 65, 34]
          },
          {
            type: 'bar',
            label: 'Projekt 3',
            backgroundColor: documentStyle.getPropertyValue('--p-orange-500'),
            data: [41, 52, 24, 74, 23, 21, 32]
          }
        ]
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

}
