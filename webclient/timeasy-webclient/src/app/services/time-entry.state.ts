import {signal} from '@angular/core';
import {TimeEntry} from '../models/timeentry.model';

export class TimeEntryState {
  readonly timeEntry = signal<TimeEntry | null>(null);
  readonly timeEntries = signal<TimeEntry[]>([]);
  readonly timeEntryDeleted = signal<boolean>(false);
  readonly error = signal<string | null>(null);
  readonly updateSuccessful = signal<boolean>(false);
  readonly lastUpdatedTimeEntry = signal<TimeEntry | null>(null);

  reset() {
    this.timeEntry.set(null);
    this.timeEntries.set([]);
    this.timeEntryDeleted.set(false);
    this.error.set(null);
    this.updateSuccessful.set(false);
    this.lastUpdatedTimeEntry.set(null);
  }

}
