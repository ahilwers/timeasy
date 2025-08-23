import { Component, computed, effect, inject, OnInit, signal, ViewChild } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MessageService } from 'primeng/api';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { TimeEntry } from '../../../models/timeentry.model';
import { TimeEntryService } from '../../../services/time-entry.service';
import { ExternalIntegrationService } from '../../../services/external-integration.service';
import { Button } from 'primeng/button';
import { InputText } from 'primeng/inputtext';
import { Toast } from 'primeng/toast';
import { DatePicker } from 'primeng/datepicker';
import { ProjectService } from '../../../services/project.service';
import { Select } from 'primeng/select';
import { Project } from '../../../models/project.model';

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
  private readonly externalService = inject(ExternalIntegrationService);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);

  timeEntryId: string = '';
  timeEntryForm!: FormGroup;

  isNew = signal<boolean>(true);
  timeEntry = this.timeEntryService.timeEntry();
  error = this.timeEntryService.error();
  updateSuccessful = this.timeEntryService.updateSuccessful();
  projects = this.projectService.projects();
  project = this.projectService.project();
  selectedProject = this.projectService.selectedProject();
  currentProjectData: Project | null = null;
  
  // Autocomplete for descriptions
  descriptionSuggestions: string[] = [];
  filteredSuggestions: string[] = [];

  constructor() {
    this.timeEntryService.resetState();
    effect(() => {
      const timeEntryData = this.timeEntry();
      if (timeEntryData) {
        this.projectService.loadProject(timeEntryData.projectId);
        this.timeEntryForm.patchValue({
          startTime: timeEntryData.startTime,
          startDate: timeEntryData.startTime,
          endTime: timeEntryData.endTime,
          endDate: timeEntryData.endTime,
          description: timeEntryData.description
        });
      }
    });
    effect(() => {
      const projectData = this.project();
      if (projectData && projectData != this.currentProjectData) {
        this.currentProjectData = projectData;
        this.timeEntryForm.patchValue({
          project: projectData,
        });
      }
    });

    effect(() => {
      const error = this.error();
      if (error) {
        this.messageService.add({ severity: 'error', summary: this.translateService.instant('globals.error'), detail: error });
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
    this.isNew.set(this.timeEntryId === '');

    if (!this.isNew()) {
      this.timeEntryService.loadTimeEntry(this.timeEntryId);
    }
  }

  onSubmit() {
    if (this.timeEntryForm.valid) {
      const startTimeStamp = this.combineDateAndTime(this.timeEntryForm.value.startDate, this.timeEntryForm.value.startTime);
      const endTimeStamp = this.timeEntryForm.value.endTime ? this.combineDateAndTime(this.timeEntryForm.value.endDate, this.timeEntryForm.value.endTime) : undefined
      const timeEntry: TimeEntry = {
        id: this.timeEntryId,
        projectId: this.timeEntryForm.value.project.id,
        startTime: startTimeStamp,
        endTime: endTimeStamp,
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
      this.messageService.add({ severity: 'error', summary: this.translateService.instant('globals.error'), detail: this.translateService.instant('timeEntries.selectProject') });
    }
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

  // Manual autocomplete for description field
  onDescriptionInput(event: Event) {
    const input = event.target as HTMLInputElement;
    const query = input.value;
    
    console.log('DEBUG: onDescriptionInput called with query:', query);
    
    if (query.length < 2) {
      console.log('DEBUG: Query too short, clearing suggestions');
      this.filteredSuggestions = [];
      return;
    }

    const selectedProject = this.timeEntryForm.get('project')?.value;
    console.log('DEBUG: Selected project:', selectedProject);
    
    if (!selectedProject || !selectedProject.id) {
      console.log('DEBUG: No project selected, clearing suggestions');
      this.filteredSuggestions = [];
      return;
    }

    console.log('DEBUG: Making API call for suggestions...');
    // Get suggestions from external integration service
    this.externalService.getDescriptionSuggestions(selectedProject.id, query, 10).subscribe({
      next: (response) => {
        console.log('DEBUG: API response received:', response);
        this.filteredSuggestions = response.suggestions || [];
        console.log('DEBUG: filteredSuggestions set to:', this.filteredSuggestions);
        console.log('DEBUG: filteredSuggestions.length:', this.filteredSuggestions.length);
      },
      error: (error) => {
        console.error('DEBUG: API error:', error);
        this.filteredSuggestions = [];
      }
    });
  }

  selectSuggestion(suggestion: string) {
    this.timeEntryForm.get('description')?.setValue(suggestion);
    this.filteredSuggestions = [];
  }

  navigateToTimeEntryList() {
    this.router.navigate([`/timeentries`]);
  }
}
