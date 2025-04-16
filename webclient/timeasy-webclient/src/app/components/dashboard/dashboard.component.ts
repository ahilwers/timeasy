import {Component} from '@angular/core';
import {TranslatePipe} from '@ngx-translate/core';
import {Card} from 'primeng/card';
import {WeeklyStatisticsChartComponent} from '../charts/weekly-statistics-chart.component';
import {WeeklyPieChartComponent} from '../charts/weekly-pie-chart.component';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  templateUrl: './dashboard.component.html',
  imports: [
    TranslatePipe,
    Card,
    WeeklyStatisticsChartComponent,
    WeeklyPieChartComponent
  ],
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent {


}
