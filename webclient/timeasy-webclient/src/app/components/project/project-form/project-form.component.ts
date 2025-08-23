import {Component, effect, inject, Input, OnInit, signal} from '@angular/core';
import {FormBuilder, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute, Router} from '@angular/router';
import {ProjectService} from '../../../services/project.service';
import {Project} from '../../../models/project.model';
import {InputText} from 'primeng/inputtext';
import {Button} from 'primeng/button';
import {Toast} from 'primeng/toast';
import {MessageService} from 'primeng/api';
import {TranslatePipe, TranslateService} from '@ngx-translate/core';
import {DropdownModule} from 'primeng/dropdown';
import {DatePicker} from 'primeng/datepicker';
import {DateOnly} from '../../../models/date_only';
import {InputNumber} from 'primeng/inputnumber';
import {ToggleSwitch} from 'primeng/toggleswitch';
import {ExternalIntegrationComponent} from '../external-integration/external-integration.component';

@Component({
  selector: 'app-project-form',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    InputText,
    Button,
    Toast,
    TranslatePipe,
    DropdownModule,
    DatePicker,
    InputNumber,
    ToggleSwitch,
    ExternalIntegrationComponent
  ],
  providers: [MessageService],
  templateUrl: './project-form.component.html',
  styleUrl: './project-form.component.css'
})
export class ProjectFormComponent implements OnInit {
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly formBuilder = inject(FormBuilder);
  private readonly projectService = inject(ProjectService);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);

  projectId : string = '';
  projectForm!: FormGroup;

  isNew = signal<boolean>(true);
  project = this.projectService.project();
  error = this.projectService.error();
  updateSuccessful = this.projectService.updateSuccessful();

  colors = [
    { name: 'Blue', hex: '#1E90FF' },
    { name: 'Green', hex: '#2ECC71' },
    { name: 'Red', hex: '#E74C3C' },
    { name: 'Orange', hex: '#E67E22' },
    { name: 'Purple', hex: '#9B59B6' },
    { name: 'Cyan', hex: '#1ABC9C' },
    { name: 'Yellow', hex: '#F1C40F' },
    { name: 'Pink', hex: '#E91E63' },
    { name: 'Gray', hex: '#95A5A6' },
    { name: 'Brown', hex: '#A0522D' },
  ];

  constructor() {
    this.projectService.resetState();
    effect(() => {
      const projectData = this.project();
      if (projectData && !this.isNew()) {
        this.projectForm.patchValue({
          name: projectData.name,
          color: projectData.color,
          deadline: projectData.deadline ? projectData.deadline.toDate() : null,
          hourlyRate: projectData.hourlyRate ? projectData.hourlyRate : 0,
          timeBudget: projectData.timeBudget ? projectData.timeBudget : 0,
          isActive: projectData.isActive,
        });
      }
      const updateSuccessful = this.updateSuccessful();
      if (updateSuccessful) {
        this.navigateToProjectList();
      }
      const error = this.error();
      if (error) {
        this.messageService.add({severity: 'error', summary: this.translateService.instant('globals.error'), detail: error});
      }
    });
  }

  ngOnInit() {
    this.projectForm = this.formBuilder.group({
      name: ['', [Validators.required]],
      color: [this.colors[0].hex, [Validators.required]],
      deadline: [null],
      hourlyRate: [0 as number],
      timeBudget: [0 as number],
      isActive: [true as boolean]
    });

    this.projectId = this.route.snapshot.paramMap.get('projectId') || '';
    this.isNew.set(this.projectId==='');

    if (!this.isNew()) {
      this.projectService.loadProject(this.projectId);
    }
  }

  onSubmit() {
    if (this.projectForm.valid) {
      const deadlineDate: Date | null = this.projectForm.value.deadline;
      const project: Project = {
        id: this.projectId,
        name: this.projectForm.value.name,
        color: this.projectForm.value.color,
        deadline: deadlineDate ? DateOnly.fromDate(deadlineDate) : new DateOnly(0,0,0),
        hourlyRate: this.projectForm.value.hourlyRate ? Number(this.projectForm.value.hourlyRate) : 0,
        timeBudget: this.projectForm.value.timeBudget ? Number(this.projectForm.value.timeBudget) : 0,
        isActive: this.projectForm.value.isActive
      }
      if (this.isNew()) {
        this.projectService.addProject(project);
      } else {
        this.projectService.updateProject(project);
      }
    } else {
      this.showFormValidationErrors();
    }
  }

  private showFormValidationErrors() {
    if (this.projectForm.get('name')!.invalid) {
      this.messageService.add({severity: 'error', summary: this.translateService.instant('globals.error'), detail: this.translateService.instant('projects.specifyName')});
    }
  }

  navigateToProjectList() {
    this.router.navigate([`/projects`]);
  }
}
