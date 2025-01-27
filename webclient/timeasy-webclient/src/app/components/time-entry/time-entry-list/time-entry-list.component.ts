import {Component, effect, inject, OnInit} from '@angular/core';
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
  ],
  providers: [ConfirmationService, MessageService],
  templateUrl: './time-entry-list.component.html',
  styleUrl: './time-entry-list.component.css'
})

export class TimeEntryListComponent implements OnInit {

  private readonly router = inject(Router);
  private readonly timeEntryService = inject(TimeEntryService);
  private readonly confirmationService = inject(ConfirmationService);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);

  timeEntries = this.timeEntryService.timeEntries();
  deleted = this.timeEntryService.deleted();
  error = this.timeEntryService.error();

  constructor() {
    effect(() => {
      if (this.deleted()) {
        this.timeEntryService.loadTimeEntries();
      }
      const error = this.error();
      if (error) {
        this.messageService.add({severity: 'error', summary: this.translateService.instant('globals.error'), detail: error});
      }
    });
  }

  ngOnInit(): void {
    this.timeEntryService.loadTimeEntries()
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
      message: `Are you sure that you want to delete the timeEntry "${timeEntry.description}"?`,
      header: 'Confirmation',
      closable: true,
      closeOnEscape: true,
      icon: 'pi pi-exclamation-triangle',
      rejectButtonProps: {
        label: 'No',
        severity: 'secondary',
        outlined: true,
      },
      acceptButtonProps: {
        label: 'Yes',
      },
      accept: () => {
        this.deleteTimeEntry(timeEntry)
      },
    });
  }

  deleteTimeEntry(timeEntry: TimeEntry): void {
    this.timeEntryService.deleteTimeEntry(timeEntry);
  }
}
