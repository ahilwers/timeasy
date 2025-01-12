import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class TimeEntryService {
  private apiUrl = 'http://localhost:8080/api/v1/timeentries';

  constructor(private http: HttpClient) {}

  getTimeEntries(): Observable<any> {
    return this.http.get(this.apiUrl);
  }

  createTimeEntry(data: any): Observable<any> {
    return this.http.post(this.apiUrl, data);
  }
}
