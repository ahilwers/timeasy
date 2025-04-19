import {Component} from '@angular/core';
import {TranslatePipe} from '@ngx-translate/core';
import {WeeklyStatisticsChartComponent} from '../charts/weekly-statistics-chart.component';
import {WeeklyPieChartComponent} from '../charts/weekly-pie-chart.component';
import {CardComponent} from '../card/card.component';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  templateUrl: './dashboard.component.html',
  imports: [
    TranslatePipe,
    WeeklyStatisticsChartComponent,
    WeeklyPieChartComponent,
    CardComponent
  ],
  styleUrl: './dashboard.component.css'
})

export class DashboardComponent {
}
