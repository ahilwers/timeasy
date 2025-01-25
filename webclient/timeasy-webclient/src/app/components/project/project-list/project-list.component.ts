import {Component, effect, inject, OnInit, signal} from '@angular/core';
import {Project} from '../../../models/project.model';
import {ProjectService} from '../../../services/project.service';
import {Button} from 'primeng/button';
import {TableModule} from 'primeng/table';
import {Router} from '@angular/router';
import {ConfirmationService, MessageService} from 'primeng/api';
import {ConfirmDialog} from 'primeng/confirmdialog';
import {Toast} from 'primeng/toast';

@Component({
  selector: 'app-project-list',
  standalone: true,
  imports: [Button, TableModule, ConfirmDialog, Toast],
  providers: [ConfirmationService, MessageService],
  templateUrl: './project-list.component.html',
  styleUrl: './project-list.component.css'
})
export class ProjectListComponent implements OnInit {

  private readonly router = inject(Router);
  private readonly projectService = inject(ProjectService);
  private readonly confirmationService = inject(ConfirmationService);
  private readonly messageService = inject(MessageService);

  projects = this.projectService.getProjectsSignal();
  projectDeleted = this.projectService.getDeletedSignal();
  error = this.projectService.getErrorSignal()  ;

  constructor() {
    effect(() => {
      const projectDeleted = this.projectDeleted();
      if (projectDeleted) {
        this.projectService.loadProjects();
      }
      const error = this.error();
      if (error) {
        this.messageService.add({severity: 'error', summary: 'Error', detail: error});
      }
    });
  }

  ngOnInit(): void {
    this.projectService.loadProjects()
  }

  editProject(project: Project): void {
    this.router.navigate([`/projects/edit/${project.Id}`]);
  }

  addProject() {
    this.router.navigate(['/projects/add']);
  }

  deleteProjectRequest(project: Project) {
    this.confirmationService.confirm({
      target: event?.target as EventTarget,
      message: `Are you sure that you want to delete the project "${project.name}"?`,
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
        this.deleteProject(project)
      },
    });
  }

  deleteProject(project: Project): void {
    this.projectService.deleteProject(project);
  }


}
