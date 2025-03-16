import {Component, effect, inject, OnInit, signal} from '@angular/core';
import {UtcToLocalDatePipe} from '../../../pipes/utc-to-local-date.pipe';
import {UtcToLocalTimePipe} from '../../../pipes/utc-to-local-time.pipe';
import {Button} from 'primeng/button';
import {TableModule} from 'primeng/table';
import {TimeEntry} from '../../../models/timeentry.model';
import {TranslatePipe, TranslateService} from '@ngx-translate/core';
import {Router} from '@angular/router';
import {ConfirmationService, MessageService} from 'primeng/api';
import {TimeEntryService} from '../../../services/time-entry.service';
import {ConfirmDialog} from 'primeng/confirmdialog';
import {Toast} from 'primeng/toast';
import {ProjectService} from '../../../services/project.service';
import {Project} from '../../../models/project.model';
import {Select} from 'primeng/select';
import {FormsModule} from '@angular/forms';

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
  ],
  providers: [ConfirmationService, MessageService],
  templateUrl: './time-entry-list.component.html',
  styleUrl: './time-entry-list.component.css'
})

export class TimeEntryListComponent implements OnInit {

  private readonly router = inject(Router);
  private readonly timeEntryService = inject(TimeEntryService);
  private readonly projectService = inject(ProjectService);
  private readonly confirmationService = inject(ConfirmationService);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);

  timeEntries = this.timeEntryService.timeEntries();
  deleted = this.timeEntryService.deleted();
  error = this.timeEntryService.error();
  projects = this.projectService.projects();

  projectMap = new Map<string, Project>();
  selectedProjectId = signal<string | undefined>(undefined);

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
      this.timeEntryService.loadTimeEntries(this.selectedProjectId());
    });
  }

  ngOnInit(): void {
    this.projectService.loadProjects();
    this.timeEntryService.loadTimeEntries(this.selectedProjectId());
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
      this.selectedProjectId.set(event.value.id);
    } else {
      this.selectedProjectId.set(undefined);
    }
  }

  projectSelectionCleared($event: Event) {
    this.selectedProjectId.set(undefined);
  }
}
