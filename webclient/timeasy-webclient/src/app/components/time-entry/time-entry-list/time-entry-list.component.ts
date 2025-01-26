import {Component, OnInit} from '@angular/core';
import {TimeEntry} from '../../../models/timeentry.model';
import {TimeEntryService} from '../../../services/time-entry.service';
import {UtcToLocalDatePipe} from '../../../pipes/utc-to-local-date.pipe';
import {UtcToLocalTimePipe} from '../../../pipes/utc-to-local-time.pipe';
import {Button} from 'primeng/button';
import {TableModule} from 'primeng/table';
import {Project} from '../../../models/project.model';
import {TranslatePipe} from '@ngx-translate/core';

@Component({
  selector: 'app-time-entry-list',
  standalone: true,
  imports: [
    UtcToLocalDatePipe,
    UtcToLocalTimePipe,
    Button,
    TableModule,
    TranslatePipe,
  ],
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


  editTImeEntry(timeEntry: TimeEntry) {
    console.log("editTimeEntry");
  }

  deleteTimeEntry(timeEntry: TimeEntry) {
    console.log("deleteTimeEntry");
  }
}
