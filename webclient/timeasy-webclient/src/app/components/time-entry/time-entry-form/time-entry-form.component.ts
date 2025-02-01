import {Component, computed, effect, inject, OnInit, signal, ViewChild} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {FormBuilder, FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {MessageService} from 'primeng/api';
import {TranslatePipe, TranslateService} from '@ngx-translate/core';
import {TimeEntry} from '../../../models/timeentry.model';
import {TimeEntryService} from '../../../services/time-entry.service';
import {Button} from 'primeng/button';
import {InputText} from 'primeng/inputtext';
import {Toast} from 'primeng/toast';
import {DatePicker} from 'primeng/datepicker';
import {ProjectService} from '../../../services/project.service';
import {Select} from 'primeng/select';
import {Project} from '../../../models/project.model';

@Component({
  selector: 'app-time-entry-form',
  standalone: true,
  imports: [
    Button,
    InputText,
    ReactiveFormsModule,
    Toast,
    TranslatePipe,
    DatePicker,
    Select
  ],
  providers: [MessageService],
  templateUrl: './time-entry-form.component.html',
  styleUrl: './time-entry-form.component.css'
})

export class TimeEntryFormComponent implements OnInit {
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly formBuilder = inject(FormBuilder);
  private readonly timeEntryService = inject(TimeEntryService);
  private readonly projectService = inject(ProjectService);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);

  timeEntryId : string = '';
  timeEntryForm!: FormGroup;

  isNew = signal<boolean>(true);
  timeEntry = this.timeEntryService.timeEntry();
  error = this.timeEntryService.error();
  updateSuccessful = this.timeEntryService.updateSuccessful();
  projects = this.projectService.projects();
  project = this.projectService.project();
  selectedProject = this.projectService.selectedProject();
  currentProjectData: Project | null = null;

  constructor() {
    this.timeEntryService.resetState();
    effect(() => {
      const timeEntryData = this.timeEntry();
      if (timeEntryData) {
        this.projectService.loadProject(timeEntryData.projectId);
        const startTime = new Date(timeEntryData.startTimeUTCUnix*1000);
        const endTime = timeEntryData.endTimeUTCUnix>0 ? new Date(timeEntryData.startTimeUTCUnix*1000) : undefined;
        console.log("loading");
        console.log(timeEntryData.projectId);
        this.timeEntryForm.patchValue({
          startTime: startTime,
          startDate: startTime,
          endTime: endTime,
          endDate: endTime,
          description: timeEntryData.description
        });
      }
    });
    effect(() => {
      console.log("project")
      const projectData = this.project();
      if (projectData && projectData!=this.currentProjectData) {
        this.currentProjectData = projectData;
        this.timeEntryForm.patchValue({
          project: projectData,
        });
      }
    });

    effect(() => {
      const error = this.error();
      if (error) {
        this.messageService.add({severity: 'error', summary: this.translateService.instant('globals.error'), detail: error});
      }
    });
    effect(() => {
      const updateSuccessful = this.updateSuccessful();
      if (updateSuccessful) {
        this.navigateToTimeEntryList();
      }
    });
  }

  ngOnInit() {
    this.projectService.loadProjects();
    this.timeEntryForm = this.formBuilder.group({
      project: new FormControl<Project | null>(this.selectedProject(), [Validators.required]),
      startTime: [new Date(), [Validators.required]],
      startDate: [new Date(), [Validators.required]],
      endDate: ['', []],
      endTime: ['', []],
      description: ['', []]
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
      const startTimeStamp = this.generateTimeStamp(this.timeEntryForm.value.startDate, this.timeEntryForm.value.startTime);
      const endTimeStamp = this.timeEntryForm.value.endTime ? this.generateTimeStamp(this.timeEntryForm.value.endDate, this.timeEntryForm.value.endTime) : 0;
      const timeEntry: TimeEntry = {
        id: this.timeEntryId,
        projectId: this.timeEntryForm.value.project.Id,
        startTimeUTCUnix: startTimeStamp,
        endTimeUTCUnix: endTimeStamp,
        description: this.timeEntryForm.value.description
      }
      this.projectService.selectProject(this.timeEntryForm.value.project);
      if (this.isNew()) {
        this.timeEntryService.addTimeEntry(timeEntry);
      } else {
        this.timeEntryService.updateTimeEntry(timeEntry);
      }
    } else {
      this.showFormValidationErrors();
    }
  }

  private showFormValidationErrors() {
    if (this.timeEntryForm.get('project')!.invalid) {
      this.messageService.add({severity: 'error', summary: this.translateService.instant('globals.error'), detail: this.translateService.instant('timeEntries.selectProject')});
    }
  }

  private generateTimeStamp(date: Date, time: Date): number {
    const combinedDateAndTime = this.combineDateAndTime(date, time);
    return Math.floor(combinedDateAndTime.getTime()/1000);
  }

  private combineDateAndTime(date: Date, time: Date): Date {
    return new Date(
      date.getFullYear(),
      date.getMonth(),
      date.getDate(),
      time.getHours(),
      time.getMinutes(),
      time.getSeconds(),
      time.getMilliseconds()
    )
  }

  navigateToTimeEntryList() {
    this.router.navigate([`/timeentries`]);
  }
}
