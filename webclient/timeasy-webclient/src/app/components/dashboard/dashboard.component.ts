import {Component} from '@angular/core';
import {TranslatePipe} from '@ngx-translate/core';
import {Card} from 'primeng/card';
import {WeeklyStatisticsChartComponent} from '../charts/weekly-statistics-chart.component';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  templateUrl: './dashboard.component.html',
  imports: [
    TranslatePipe,
    Card,
    WeeklyStatisticsChartComponent
  ],
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent {


}
