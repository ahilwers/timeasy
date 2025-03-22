import {Component, effect, inject, OnInit} from '@angular/core';
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
import {Select} from 'primeng/select';
import {FormsModule} from '@angular/forms';
import {DatePicker} from 'primeng/datepicker';
import {FloatLabel} from 'primeng/floatlabel';
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
    FloatLabel,
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

  protected readonly TimeEntryExportFomat = TimeEntryExportFomat;
}
