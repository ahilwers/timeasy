import {Component, effect, inject, OnInit, signal} from '@angular/core';
import {UtcToLocalDatePipe} from '../../../pipes/utc-to-local-date.pipe';
import {UtcToLocalTimePipe} from '../../../pipes/utc-to-local-time.pipe';
import {Button} from 'primeng/button';
import {TableModule} from 'primeng/table';
import {TimeEntry} from '../../../models/timeentry.model';
import {TranslatePipe, TranslateService} from '@ngx-translate/core';
import {Router} from '@angular/router';
import {ConfirmationService, MenuItem, MessageService} from 'primeng/api';
import {TimeEntryExportFomat, TimeEntryService} from '../../../services/time-entry.service';
import {ConfirmDialog} from 'primeng/confirmdialog';
import {Toast} from 'primeng/toast';
import {ProjectService} from '../../../services/project.service';
import {Project} from '../../../models/project.model';
import {ExternalIntegrationService} from '../../../services/external-integration.service';
import {ExternalIssue} from '../../../models/external-issue.model';
import {Select} from 'primeng/select';
import {FormsModule} from '@angular/forms';
import {DatePicker} from 'primeng/datepicker';
import {SplitButton} from 'primeng/splitbutton';

@Component({
  selector: 'app-time-entry-list',
  standalone: true,
  imports: [
    UtcToLocalDatePipe,
    UtcToLocalTimePipe,
    Button,
    TableModule,
    TranslatePipe,
    ConfirmDialog,
    Toast,
    Select,
    FormsModule,
    DatePicker,
    SplitButton,
  ],
  providers: [ConfirmationService, MessageService],
  templateUrl: './time-entry-list.component.html',
  styleUrl: './time-entry-list.component.css'
})

export class TimeEntryListComponent implements OnInit {

  private readonly router = inject(Router);
  private readonly timeEntryService = inject(TimeEntryService);
  private readonly projectService = inject(ProjectService);
  private readonly externalService = inject(ExternalIntegrationService);
  private readonly confirmationService = inject(ConfirmationService);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);
  private selectedDates : Date[] = [];

  timeEntries = this.timeEntryService.timeEntries();
  deleted = this.timeEntryService.deleted();
  error = this.timeEntryService.error();
  projects = this.projectService.projects();
  selectedDateRange = this.timeEntryService.selectedDateRange();
  selectedProjectId = this.timeEntryService.selectedProjectId();

  projectMap = new Map<string, Project>();
  externalIssuesMap = signal(new Map<string, ExternalIssue>());
  exportMenuItems: MenuItem[];

  constructor() {
    effect(() => {
      if (this.deleted()) {
        this.timeEntryService.loadTimeEntries(this.selectedProjectId());
      }
      const error = this.error();
      if (error) {
        this.messageService.add({severity: 'error', summary: this.translateService.instant('globals.error'), detail: error});
      }
    });
    effect(() => {
      if (this.projects()) {
        this.projects().forEach(project => {
          this.projectMap.set(project.id, project);
        });
      }
    });
    effect(() => {
      this.timeEntryService.loadTimeEntries(this.selectedProjectId(), this.selectedDateRange());
    });
    // Load external issues when time entries change, but use a computed to avoid infinite loops
    effect(() => {
      const entries = this.timeEntries();
      const deleted = this.deleted();
      
      // Only load external issues when time entries are loaded (not when deleted)
      if (entries && entries.length > 0 && !deleted) {
        // Use a micro task to avoid triggering during the current change detection cycle
        Promise.resolve().then(() => {
          this.loadExternalIssuesIfNeeded(entries);
        });
      }
    });
    this.exportMenuItems = [
      {
        label: 'CSV',
        command: () => {
          this.exportTimeEntries(TimeEntryExportFomat.CSV);
        }
      },
      {
        label: 'XLSX',
        command: () => {
          this.exportTimeEntries(TimeEntryExportFomat.XLSX);
        }
      },
      {
        label: 'XLSX (eine Zeile pro Tag)' ,
        command: () => {
          this.exportTimeEntries(TimeEntryExportFomat.XLSX_ONELINEPERDAY);
        }
      },
    ];
  }

  ngOnInit(): void {
    this.projectService.loadProjects();
  }

  editTimeEntry(timeEntry: TimeEntry): void {
    this.router.navigate([`/timeentries/edit/${timeEntry.id}`]);
  }

  addTimeEntry() {
    this.router.navigate(['/timeentries/add']);
  }

  deleteTimeEntryRequest(timeEntry: TimeEntry) {
    this.confirmationService.confirm({
      target: event?.target as EventTarget,
      message: this.translateService.instant('timeEntries.confirmDelete', {name: timeEntry.description}),
      header: this.translateService.instant('timeEntries.deleteTimeEntry'),
      closable: true,
      closeOnEscape: true,
      icon: 'pi pi-exclamation-triangle',
      rejectButtonProps: {
        label: this.translateService.instant('globals.no'),
        severity: 'secondary',
        outlined: true,
      },
      acceptButtonProps: {
        label: this.translateService.instant('globals.yes'),
      },
      accept: () => {
        this.deleteTimeEntry(timeEntry)
      },
    });
  }

  deleteTimeEntry(timeEntry: TimeEntry): void {
    this.timeEntryService.deleteTimeEntry(timeEntry);
  }

  getProjectName(projectId: string): string {
    return this.projectMap.get(projectId)?.name || '';
  }

  projectSelected(event: any) {
    if (event.value) {
      this.timeEntryService.selectProject(event.value);
    } else {
      this.timeEntryService.selectProject(undefined);
    }
  }

  projectSelectionCleared($event: Event) {
    this.timeEntryService.selectProject(undefined);
  }

  updateSelectedDateRange(value: Date) {
    this.selectedDates.push(value);
    if (this.selectedDates.length == 2) {
      this.timeEntryService.selectDateRange([this.selectedDates[0], this.selectedDates[1]])
      this.selectedDates = [];
    }
  }

  getTotalTimeString(entries: TimeEntry[]): string {
    const { hours, minutes } = this.getTotalTime(entries);
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}`;
  }

  getTotalTime(entries: TimeEntry[]): { hours: number; minutes: number } {
    const totalMilliseconds = entries.reduce((sum, entry) => {
      const start = new Date(entry.startTime).getTime();
      const end = entry.endTime ? new Date(entry.endTime).getTime() : new Date().getTime();
      return sum + (end - start);
    }, 0);

    const totalMinutes = Math.floor(totalMilliseconds / (1000 * 60));
    const hours = Math.floor(totalMinutes / 60);
    const minutes = totalMinutes % 60;

    return { hours, minutes };
  }

  exportTimeEntries(exportFormat: TimeEntryExportFomat) {
    this.timeEntryService.exportTimeEntries(this.selectedProjectId(), this.selectedDateRange(), exportFormat);
  }

  private loadingExternalIssues = new Set<string>(); // Track loading state

  loadExternalIssuesIfNeeded(entries: TimeEntry[]): void {
    console.log('Loading external issues for entries:', entries.length);
    
    entries.forEach(entry => {
      if (entry.externalIssueId && 
          !this.externalIssuesMap().has(entry.externalIssueId) && 
          !this.loadingExternalIssues.has(entry.externalIssueId)) {
        
        console.log('Fetching external issue for ID:', entry.externalIssueId);
        this.loadingExternalIssues.add(entry.externalIssueId);
        
        this.externalService.getExternalIssue(entry.externalIssueId).subscribe({
          next: (externalIssue) => {
            console.log('Received external issue:', externalIssue);
            this.loadingExternalIssues.delete(entry.externalIssueId!);
            
            // Update the signal by modifying the current map
            this.externalIssuesMap.update(currentMap => {
              const newMap = new Map(currentMap);
              newMap.set(entry.externalIssueId!, externalIssue);
              console.log('Updated external issues map, now has:', newMap.size, 'issues');
              return newMap;
            });
          },
          error: (err) => {
            console.error('Failed to load external issue:', entry.externalIssueId, err);
            this.loadingExternalIssues.delete(entry.externalIssueId!);
          }
        });
      }
    });
  }


  getExternalIssueDisplay(timeEntry: TimeEntry): { key: string; url?: string } | undefined {
    // Check if we have loaded external issue data
    if (timeEntry.externalIssueId) {
      const externalIssue = this.externalIssuesMap().get(timeEntry.externalIssueId);
      console.log('Looking for external issue:', timeEntry.externalIssueId, 'found:', externalIssue);
      if (externalIssue) {
        return {
          key: externalIssue.key,
          url: externalIssue.url
        };
      }
    }
    
    // Fallback to pendingExternalRef (shows issue key without link)
    if (timeEntry.pendingExternalRef) {
      console.log('Using pendingExternalRef:', timeEntry.pendingExternalRef);
      return {
        key: timeEntry.pendingExternalRef
      };
    }
    
    return undefined;
  }

  protected readonly TimeEntryExportFomat = TimeEntryExportFomat;
}
