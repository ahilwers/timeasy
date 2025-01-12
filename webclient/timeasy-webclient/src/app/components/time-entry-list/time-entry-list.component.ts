import {Component, OnInit} from '@angular/core';
import {TimeEntry} from '../../models/timeentry.model';
import {TimeEntryService} from '../../services/time-entry.service';

@Component({
  selector: 'app-time-entry-list',
  standalone: true,
  imports: [],
  templateUrl: './time-entry-list.component.html',
  styleUrl: './time-entry-list.component.css'
})

export class TimeEntryListComponent implements OnInit {

  timeEntries : TimeEntry[] = [];

  constructor(private timeEntryService: TimeEntryService) {}

  ngOnInit(): void {
    this.loadTimeEntries();
  }

  loadTimeEntries(): void {
    this.timeEntryService.getTimeEntries().subscribe({
      next: (data) => {
        this.timeEntries = data;
        console.log(this.timeEntries)
      },
      error: (err) => {
        console.error('Error fetching time entries:', err);
      }
    });
  }


}
