import {computed, inject, Injectable} from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {catchError, map, Observable, of, tap} from 'rxjs';
import {TimeEntry} from '../models/timeentry.model';
import {environment} from '../../environments/environment';
import {TranslateService} from '@ngx-translate/core';
import {TimeEntryState} from './time-entry.state';
import {Project} from '../models/project.model';

@Injectable({
  providedIn: 'root'
})
export class TimeEntryService {
  private readonly apiUrl = `${environment.apiUrl}/timeentries`;
  private readonly state = new TimeEntryState();

  private readonly translateService = inject(TranslateService);
  private readonly http = inject(HttpClient);

  timeEntries = computed(() => this.state.timeEntries);
  timeEntry = computed(() => this.state.timeEntry);
  deleted = computed(() => this.state.timeEntryDeleted);
  error = computed(() => this.state.error);
  updateSuccessful = computed(() => this.state.updateSuccessful);
  lastUpdatedTimeEntry = computed(() => this.state.lastUpdatedTimeEntry);

  loadTimeEntry(id: string): void {
    this.resetState();
    this.http.get<any>(`${this.apiUrl}/${id}`).pipe(
      map((timeEntry) => ({
        id: timeEntry.Id,
        projectId: timeEntry.projectId,
        timeEntryId: timeEntry.timeEntryId,
        description: timeEntry.description,
        startTimeUTCUnix: timeEntry.startTimeUTCUnix,
        endTimeUTCUnix: timeEntry.EndTimeUTCUnix,
      }))
    ).subscribe({
      next: (transformedTimeEntry) => {
        this.state.timeEntry.set(transformedTimeEntry);
      },
      error: (err) => {
        console.error('Failed to load time entry:', err);
        this.state.error.set(this.translateService.instant('timeEntries.errorLoadingTimeEntry'));
      }
    });
  }

  loadTimeEntries(): void {
    this.resetState();
    this.http.get<any[]>(this.apiUrl).pipe(
      map((timeEntries) =>
        timeEntries.map((timeEntry) => ({
          id: timeEntry.Id,
          projectId: timeEntry.projectId,
          timeEntryId: timeEntry.timeEntryId,
          description: timeEntry.description,
          startTimeUTCUnix: timeEntry.startTimeUTCUnix,
          endTimeUTCUnix: timeEntry.EndTimeUTCUnix,
        }))
      ),
      catchError((err) => {
        console.error('Failed to load time entries:', err);
        this.state.error.set(this.translateService.instant('timeEntries.errorLoadingTimeEntries'));
        return of([]);
      })
    ).subscribe((transformedTimeEntries) => {
      this.state.timeEntries.set(transformedTimeEntries);
    });
  }

  updateTimeEntry(data: TimeEntry): void {
    this.resetState();
    this.http.put<TimeEntry>(`${this.apiUrl}/${data.id}`, data).pipe(
      tap((updatedTimeEntry) => {
        this.state.lastUpdatedTimeEntry.set(updatedTimeEntry);
        this.state.updateSuccessful.set(true);
        this.state.error.set(null);
      }),
      catchError((err) => {
        const errorMessage = err.error?.message || this.translateService.instant('timeEntries.errorUpdatingTimeEntry');
        this.state.error.set(errorMessage);
        this.state.updateSuccessful.set(false);
        return of(null as unknown as TimeEntry);
      }),
    ).subscribe(() => {
    })
  }

  addTimeEntry(data: TimeEntry) {
    this.resetState();
    this.http.post<TimeEntry>(`${this.apiUrl}`, data).pipe(
      tap((addedTimeEntry) => {
        this.state.lastUpdatedTimeEntry.set(addedTimeEntry);
        this.state.error.set(null);
        this.state.updateSuccessful.set(true);
      }),
      catchError((err) => {
        const errorMessage = err.error?.message || this.translateService.instant('timeEntries.errorAddingTimeEntry');
        this.state.error.set(errorMessage);
        this.state.updateSuccessful.set(false);
        return of(null as unknown as TimeEntry);
      }),
    ).subscribe(() => {
    })
  }

  deleteTimeEntry(data: TimeEntry) {
    this.resetState();
    this.http.delete<TimeEntry>(`${this.apiUrl}/${data.id}`).pipe(
      tap(() => {
        this.state.timeEntryDeleted.set(true);
        this.state.error.set(null);
        this.state.updateSuccessful.set(true);
      }),
      catchError((err) => {
        const errorMessage = err.error?.message || this.translateService.instant('timeEntries.errorDeletingTimeEntry');
        this.state.error.set(errorMessage);
        this.state.updateSuccessful.set(false);
        return of(null as unknown as TimeEntry);
      }),
    ).subscribe(() => {
    })
  }

  resetState() {
    this.state.reset();
  }


}
