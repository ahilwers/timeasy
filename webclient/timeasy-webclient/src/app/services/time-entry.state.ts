import {signal} from '@angular/core';
import {TimeEntry} from '../models/timeentry.model';

export class TimeEntryState {
  readonly timeEntry = signal<TimeEntry | null>(null);
  readonly timeEntries = signal<TimeEntry[]>([]);
  readonly timeEntryDeleted = signal<boolean>(false);
  readonly error = signal<string | null>(null);
  readonly updateSuccessful = signal<boolean>(false);
  readonly lastUpdatedTimeEntry = signal<TimeEntry | null>(null);
  readonly selectedDateRange = signal<[Date, Date]>(this.getDefaultDateRange());
  readonly selectedProjectId = signal<string | undefined>(undefined);

  reset() {
    this.timeEntry.set(null);
    this.timeEntries.set([]);
    this.timeEntryDeleted.set(false);
    this.error.set(null);
    this.updateSuccessful.set(false);
    this.lastUpdatedTimeEntry.set(null);
  }

  getDefaultDateRange(): [Date, Date] {
    const end = new Date();

    const start = new Date();
    start.setMonth(start.getMonth() - 1);

    return [start, end];
  }

}
