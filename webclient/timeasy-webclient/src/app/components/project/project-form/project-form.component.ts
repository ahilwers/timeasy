import {Component, effect, inject, Input, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute} from '@angular/router';
import {ProjectService} from '../../../services/project.service';
import {Project} from '../../../models/project.model';

@Component({
  selector: 'app-project-form',
  standalone: true,
  imports: [
    ReactiveFormsModule
  ],
  templateUrl: './project-form.component.html',
  styleUrl: './project-form.component.css'
})
export class ProjectFormComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly formBuilder = inject(FormBuilder);
  private readonly projectService = inject(ProjectService);

  private isNew : boolean = true;

  projectId : string = '';
  projectForm!: FormGroup;

  project = this.projectService.getProjectSignal();
  error = this.projectService.getErrorSignal();

  constructor() {
    effect(() => {
      const projectData = this.project();
      if (projectData) {
        this.projectForm.patchValue({
          name: projectData.name
        });
      }
    });
  }

  ngOnInit() {
    this.projectForm = this.formBuilder.group({
      name: ['', [Validators.required]]
    });

    this.projectId = this.route.snapshot.paramMap.get('projectId') || '';
    this.isNew = this.projectId==='';

    if (!this.isNew) {
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
      if (this.isNew) {
        this.projectService.addProject(project);
      } else {
        this.projectService.updateProject(project);
      }
    }
  }
}
