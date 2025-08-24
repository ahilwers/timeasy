import { computed, inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { catchError, map, Observable, of, switchMap, tap } from 'rxjs';
import { TimeEntry } from '../models/timeentry.model';
import { environment } from '../../environments/environment';
import { TranslateService } from '@ngx-translate/core';
import { TimeEntryState } from './time-entry.state';
import { ExternalIntegrationService } from './external-integration.service';

export enum TimeEntryExportFomat {
  CSV = 'CSV',
  XLSX = 'XLSX',
  XLSX_ONELINEPERDAY = 'XLSX_ONELINEPERDAY'
}

@Injectable({
  providedIn: 'root'
})
export class TimeEntryService {
  private readonly apiUrl = `${environment.apiUrl}/timeentries`;
  private readonly state = new TimeEntryState();

  private readonly translateService = inject(TranslateService);
  private readonly http = inject(HttpClient);
  private readonly externalService = inject(ExternalIntegrationService);

  timeEntries = computed(() => this.state.timeEntries);
  timeEntry = computed(() => this.state.timeEntry);
  lastOpenTimeEntry = computed(() => this.state.lastOpenTimeEntry);
  deleted = computed(() => this.state.timeEntryDeleted);
  error = computed(() => this.state.error);
  updateSuccessful = computed(() => this.state.updateSuccessful);
  lastUpdatedTimeEntry = computed(() => this.state.lastUpdatedTimeEntry);
  selectedDateRange = computed(() => this.state.selectedDateRange);
  selectedProjectId = computed(() => this.state.selectedProjectId);

  loadTimeEntry(id: string): void {
    this.resetState();
    this.http.get<TimeEntry>(`${this.apiUrl}/${id}`).subscribe({
      next: (timeEntry) => {
        const transformedTimeEntry = {
          ...timeEntry,
          startTime: new Date(timeEntry.startTime),
          endTime: timeEntry.endTime ? new Date(timeEntry.endTime) : undefined
        };
        this.state.timeEntry.set(transformedTimeEntry);
      },
      error: (err) => {
        console.error('Failed to load time entry:', err);
        this.state.error.set(this.translateService.instant('timeEntries.errorLoadingTimeEntry'));
      }
    });
  }

  loadLastOpenTimeEntry(): void {
    this.resetState();
    this.http.get<TimeEntry>(`${this.apiUrl}/lastopen`).subscribe({
      next: (timeEntry) => {
        const transformedTimeEntry = {
          ...timeEntry,
          startTime: new Date(timeEntry.startTime),
          endTime: timeEntry.endTime ? new Date(timeEntry.endTime) : undefined
        };
        this.state.lastOpenTimeEntry.set(transformedTimeEntry);
      },
    });
  }

  loadTimeEntries(projectId: string | undefined, dateRange?: [Date, Date]): void {
    console.log("loading time entries")
    this.resetState();
    var apiUrl = this.apiUrl;
    if (projectId) {
      apiUrl += `?projectId=${projectId}`;
    }
    if (dateRange) {
      if (projectId) {
        apiUrl +="&";
      } else {
        apiUrl += "?";
      }
      apiUrl += `startDate=${this.formatDateToRFC3339(dateRange[0])}&endDate=${this.formatDateToRFC3339(dateRange[1])}`;
    }
    this.http.get<TimeEntry[]>(apiUrl).subscribe({
      next: (timeEntries) => {
        const transformedTimeEntries = timeEntries.map((entry) => ({
          ...entry,
          startTime: new Date(entry.startTime),
          endTime: entry.endTime ? new Date(entry.endTime) : undefined
        }));
        console.log(transformedTimeEntries);
        this.state.timeEntries.set(transformedTimeEntries);
      },
      error: (err) => {
        console.error('Failed to load time entries:', err);
        this.state.error.set(this.translateService.instant('timeEntries.errorLoadingTimeEntries'));
      }
    });
  }

  formatDateToRFC3339(date: Date): string {
    return new Intl.DateTimeFormat('sv-SE').format(date);
  }

  updateTimeEntry(data: TimeEntry): void {
    this.resetState();
    
    // Check if description contains issue patterns
    const issuePattern = this.externalService.detectIssuePattern(data.description);
    
    if (issuePattern && data.projectId) {
      // If issue pattern detected, resolve it first
      this.externalService.resolveIssue(data.projectId, data.description).pipe(
        switchMap((resolveResult) => {
          // Update time entry with resolved issue info
          const updatedData = { ...data };
          if (resolveResult.status === 'resolved' && resolveResult.issueId) {
            updatedData.externalIssueId = resolveResult.issueId;
          } else if (resolveResult.status === 'pending') {
            updatedData.pendingExternalRef = resolveResult.key;
          }
          
          return this.http.put<TimeEntry>(`${this.apiUrl}/${updatedData.id}`, updatedData);
        }),
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
        })
      ).subscribe();
    } else {
      // No issue pattern detected, update normally
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
        })
      ).subscribe();
    }
  }

  addTimeEntry(data: TimeEntry) {
    this.resetState();
    
    // Check if description contains issue patterns
    const issuePattern = this.externalService.detectIssuePattern(data.description);
    console.log('Issue pattern detection:', { description: data.description, pattern: issuePattern });
    
    if (issuePattern && data.projectId) {
      // If issue pattern detected, resolve it first
      this.externalService.resolveIssue(data.projectId, data.description).pipe(
        switchMap((resolveResult) => {
          console.log('Issue resolve result:', resolveResult);
          // Update time entry with resolved issue info
          const updatedData = { ...data };
          if (resolveResult.status === 'resolved' && resolveResult.issueId) {
            updatedData.externalIssueId = resolveResult.issueId;
            console.log('Setting externalIssueId:', resolveResult.issueId);
          } else if (resolveResult.status === 'pending') {
            updatedData.pendingExternalRef = resolveResult.key;
            console.log('Setting pendingExternalRef:', resolveResult.key);
          }
          console.log('Saving time entry with data:', updatedData);
          
          return this.http.post<TimeEntry>(`${this.apiUrl}`, updatedData);
        }),
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
        })
      ).subscribe();
    } else {
      // No issue pattern detected, save normally
      console.log('No issue pattern detected, saving normally');
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
        })
      ).subscribe();
    }
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

  selectDateRange(dateRange: [Date, Date]) {
    this.state.selectedDateRange.set(dateRange);
  }

  selectProject(projectId: string | undefined) {
    this.state.selectedProjectId.set(projectId);
  }

  resetState() {
    this.state.reset();
  }


  exportTimeEntries(projectId: string | undefined, dateRange?: [Date, Date], exportFormat?: TimeEntryExportFomat) {
    let apiUrl = this.apiUrl;
    switch (exportFormat) {
      case TimeEntryExportFomat.XLSX:
        apiUrl += "/asxlsx";
        break;
      case TimeEntryExportFomat.XLSX_ONELINEPERDAY:
        apiUrl += "/asxlsxonelineperday";
        break;
      default:
        apiUrl += "/ascsv";
    }
    if (projectId) {
      apiUrl += `?projectId=${projectId}`;
    }
    if (dateRange) {
      if (projectId) {
        apiUrl += "&";
      } else {
        apiUrl += "?";
      }
      apiUrl += `startDate=${this.formatDateToRFC3339(dateRange[0])}&endDate=${this.formatDateToRFC3339(dateRange[1])}`;
    }

    // API-Request mit responseType 'blob' (da eine Datei zurückgegeben wird)
    this.http.get(apiUrl, { responseType: 'blob' }).subscribe({
      next: (blob) => {
        const contentType = blob.type;

        // Prüfen, ob die Datei ein XLSX oder CSV ist
        let fileExtension = ".csv";
        if (contentType === "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet") {
          fileExtension = ".xlsx";
        }

        // Datei-Download starten
        const fileName = `export${fileExtension}`;
        const link = document.createElement('a');
        link.href = window.URL.createObjectURL(blob);
        link.download = fileName;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
      },
      error: (err) => {
        const errorMessage = err.error?.message || this.translateService.instant('timeEntries.errorExportingTimeEntries');
        this.state.error.set(errorMessage);
      }
    });
  }
}
