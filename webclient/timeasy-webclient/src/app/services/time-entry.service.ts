import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {map, Observable} from 'rxjs';
import {TimeEntry} from '../models/timeentry.model';

@Injectable({
  providedIn: 'root'
})
export class TimeEntryService {
  private apiUrl = 'http://localhost:8080/api/v1/timeentries';

  constructor(private http: HttpClient) {}

  getTimeEntries(): Observable<TimeEntry[]> {
    return this.http.get<any[]>(this.apiUrl).pipe(
      map((entries) =>
        entries.map((entry) => ({
          id: entry.Id,
          projectId: entry.projectId,
          description: entry.description,
          startTimeUTCUnix: entry.startTimeUTCUnix,
          endTimeUTCUnix: entry.EndTimeUTCUnix,
        }))
      )
    );
  }

  createTimeEntry(data: TimeEntry): Observable<TimeEntry> {
    return this.http.post<TimeEntry>(this.apiUrl, data);
  }
}
