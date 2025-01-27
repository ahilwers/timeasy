import {Component, effect, inject, OnInit, signal} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {MessageService} from 'primeng/api';
import {TranslateService} from '@ngx-translate/core';
import {TimeEntry} from '../../../models/timeentry.model';
import {TimeEntryService} from '../../../services/time-entry.service';

@Component({
  selector: 'app-time-entry-form',
  standalone: true,
  imports: [],
  providers: [MessageService],
  templateUrl: './time-entry-form.component.html',
  styleUrl: './time-entry-form.component.css'
})

export class TimeEntryFormComponent implements OnInit {
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly formBuilder = inject(FormBuilder);
  private readonly timeEntryService = inject(TimeEntryService);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);

  timeEntryId : string = '';
  timeEntryForm!: FormGroup;

  isNew = signal<boolean>(true);
  timeEntry = this.timeEntryService.timeEntry();
  error = this.timeEntryService.error();
  updateSuccessful = this.timeEntryService.updateSuccessful();

  constructor() {
    this.timeEntryService.resetState();
    effect(() => {
      const timeEntryData = this.timeEntry();
      if (timeEntryData) {
        this.timeEntryForm.patchValue({
          projectId: timeEntryData.projectId,
          startTime: timeEntryData.startTimeUTCUnix,
          endTime: timeEntryData.endTimeUTCUnix,
          description: timeEntryData.description
        });
      }
      const updateSuccessful = this.updateSuccessful();
      if (updateSuccessful) {
        this.navigateToTimeEntryList();
      }
      const error = this.error();
      if (error) {
        this.messageService.add({severity: 'error', summary: this.translateService.instant('globals.error'), detail: error});
      }
    });
  }

  ngOnInit() {
    this.timeEntryForm = this.formBuilder.group({
      projectId: ['', [Validators.required]],
      startTime: ['', [Validators.required]],
      endTime: ['', []],
      description: ['', [Validators.required]]
    });

    this.timeEntryId = this.route.snapshot.paramMap.get('timeEntryId') || '';
    this.isNew.set(this.timeEntryId==='');

    if (!this.isNew()) {
      this.timeEntryService.loadTimeEntry(this.timeEntryId);
    }
  }

  onSubmit() {
    console.log(this.timeEntryForm.value);
    if (this.timeEntryForm.valid) {
      const timeEntry: TimeEntry = {
        id: this.timeEntryId,
        projectId: this.timeEntryForm.value.projectId,
        startTimeUTCUnix: this.timeEntryForm.value.startTime,
        endTimeUTCUnix: this.timeEntryForm.value.endTime,
        description: this.timeEntryForm.value.description
      }
      if (this.isNew()) {
        this.timeEntryService.addTimeEntry(timeEntry);
      } else {
        this.timeEntryService.updateTimeEntry(timeEntry);
      }
    }
  }

  navigateToTimeEntryList() {
    this.router.navigate([`/timeentries`]);
  }

}
