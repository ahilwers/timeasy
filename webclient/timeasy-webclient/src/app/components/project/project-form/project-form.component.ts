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

@Component({
  selector: 'app-project-form',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    InputText,
    Button,
    Toast,
    TranslatePipe
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

  constructor() {
    this.projectService.resetState();
    effect(() => {
      const projectData = this.project();
      if (projectData) {
        this.projectForm.patchValue({
          name: projectData.name
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
      name: ['', [Validators.required]]
    });

    this.projectId = this.route.snapshot.paramMap.get('projectId') || '';
    this.isNew.set(this.projectId==='');

    if (!this.isNew()) {
      this.projectService.loadProject(this.projectId);
    }
  }

  onSubmit() {
    console.log(this.projectForm.value);
    if (this.projectForm.valid) {
      const project: Project = {
        Id: this.projectId,
        name: this.projectForm.value.name
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
