import {computed, inject, Injectable, signal, WritableSignal} from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {catchError, map, of, tap} from 'rxjs';
import {Project} from '../models/project.model';
import {ProjectState} from './project.state';
import {environment} from '../../environments/environment';
import {TranslateService} from '@ngx-translate/core';

@Injectable({
  providedIn: 'root'
})

export class ProjectService {
  private readonly apiUrl = `${environment.apiUrl}/projects`;
  private readonly state = new ProjectState();

  private readonly translateService = inject(TranslateService);
  private readonly http = inject(HttpClient);

  projects = computed(() => this.state.projects);
  project = computed(() => this.state.project);
  deleted = computed(() => this.state.projectDeleted);
  error = computed(() => this.state.error);
  updateSuccessful = computed(() => this.state.updateSuccessful);
  lastUpdatedProject = computed(() => this.state.lastUpdatedProject);
  selectedProject = computed(() => this.state.selectedProject);

  loadProject(id: string): void {
    this.resetState();
    this.http.get<Project>(`${this.apiUrl}/${id}`).subscribe({
      next: (project) => {
        this.state.project.set(project);
      },
      error: (err) => {
        console.error('Failed to load project:', err);
        this.state.error.set(this.translateService.instant('projects.errorLoadingProject'));
      }
    });
  }

  loadProjects(): void {
    this.resetState();
    this.http.get<Project[]>(this.apiUrl).subscribe({
      next: (projects) => {
        this.state.projects.set(projects);
      },
      error: (err) => {
        console.error('Failed to load projects:', err);
        this.state.error.set(this.translateService.instant('projects.errorLoadingProjects'));
      }
    });
  }

  updateProject(data: Project): void {
    this.resetState();
    this.http.put<Project>(`${this.apiUrl}/${data.id}`, data).pipe(
      tap((updatedProject) => {
        this.state.lastUpdatedProject.set(updatedProject);
        this.state.updateSuccessful.set(true);
        this.state.error.set(null);
      }),
      catchError((err) => {
        const errorMessage = err.error?.message || this.translateService.instant('projects.errorUpdatingProject');
        this.state.error.set(errorMessage);
        this.state.updateSuccessful.set(false);
        return of(null as unknown as Project);
      }),
    ).subscribe(() => {
    })
  }

  addProject(data: Project) {
    this.resetState();
    this.http.post<Project>(`${this.apiUrl}`, data).pipe(
      tap((addedProject) => {
        this.state.lastUpdatedProject.set(addedProject);
        this.state.error.set(null);
        this.state.updateSuccessful.set(true);
      }),
      catchError((err) => {
        const errorMessage = err.error?.message || this.translateService.instant('projects.errorAddingProject');
        this.state.error.set(errorMessage);
        this.state.updateSuccessful.set(false);
        return of(null as unknown as Project);
      }),
    ).subscribe(() => {
    })
  }

  deleteProject(data: Project) {
    this.resetState();
    this.http.delete<Project>(`${this.apiUrl}/${data.id}`).pipe(
      tap(() => {
        this.state.projectDeleted.set(true);
        this.state.error.set(null);
        this.state.updateSuccessful.set(true);
      }),
      catchError((err) => {
        const errorMessage = err.error?.message || this.translateService.instant('projects.errorDeletingProject');
        this.state.error.set(errorMessage);
        this.state.updateSuccessful.set(false);
        return of(null as unknown as Project);
      }),
    ).subscribe(() => {
    })
  }

  selectProject(project: Project) {
    this.state.selectedProject.set(project);
  }

  resetState() {
    this.state.reset();
  }
}
