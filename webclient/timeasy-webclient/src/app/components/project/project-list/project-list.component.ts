import {Component, effect, inject, OnInit, signal} from '@angular/core';
import {Project} from '../../../models/project.model';
import {ProjectService} from '../../../services/project.service';
import {Button} from 'primeng/button';
import {TableModule} from 'primeng/table';
import {Router} from '@angular/router';
import {ConfirmationService, MessageService} from 'primeng/api';
import {ConfirmDialog} from 'primeng/confirmdialog';
import {Toast} from 'primeng/toast';
import {TranslatePipe, TranslateService} from '@ngx-translate/core';
import {DateOnlyPipe} from '../../../pipes/dateonly.pipe';
import {MoneyPipe} from '../../../pipes/money.pipe';
import {NgClass} from '@angular/common';
import {DateOnly} from '../../../models/date_only';

@Component({
  selector: 'app-project-list',
  standalone: true,
  imports: [Button, TableModule, ConfirmDialog, Toast, TranslatePipe, DateOnlyPipe, MoneyPipe, NgClass],
  providers: [ConfirmationService, MessageService],
  templateUrl: './project-list.component.html',
  styleUrl: './project-list.component.css'
})
export class ProjectListComponent implements OnInit {

  private readonly router = inject(Router);
  private readonly projectService = inject(ProjectService);
  private readonly confirmationService = inject(ConfirmationService);
  private readonly messageService = inject(MessageService);
  private readonly translateService = inject(TranslateService);

  projects = this.projectService.projects();
  deleted = this.projectService.deleted();
  error = this.projectService.error();
  today = DateOnly.fromDate(new Date());

  constructor() {
    effect(() => {
      if (this.deleted()) {
        this.projectService.loadProjects();
      }
      const error = this.error();
      if (error) {
        this.messageService.add({severity: 'error', summary: this.translateService.instant('globals.error'), detail: error});
      }
    });
  }

  ngOnInit(): void {
    this.projectService.loadProjects()
  }

  editProject(project: Project): void {
    this.router.navigate([`/projects/edit/${project.id}`]);
  }

  addProject() {
    this.router.navigate(['/projects/add']);
  }

  deleteProjectRequest(project: Project) {
    this.confirmationService.confirm({
      target: event?.target as EventTarget,
      message: this.translateService.instant('projects.confirmDelete'),
      header: this.translateService.instant('projects.deleteProject'),
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
        this.deleteProject(project)
      },
    });
  }

  deleteProject(project: Project): void {
    this.projectService.deleteProject(project);
  }
}
