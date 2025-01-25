import {Injectable, signal, WritableSignal} from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {catchError, map, of, tap} from 'rxjs';
import {Project} from '../models/project.model';

@Injectable({
  providedIn: 'root'
})
export class ProjectService {
  private apiUrl = 'http://localhost:8080/api/v1/projects';

  private project = signal<Project | null>(null);
  private projects = signal<Project[]>([]);
  private projectDeleted = signal<boolean>(false);
  private error = signal<string | null>(null);
  private updateSuccessful = signal<boolean>(false);
  private lastUpdatedProject = signal<Project | null>(null);

  constructor(private http: HttpClient) {
    console.log("ProjectService constructor");
  }

  getProjectSignal(): WritableSignal<Project | null> {
    return this.project
  }

  getProjectsSignal() {
    return this.projects;
  }

  getDeletedSignal() {
    return this.projectDeleted;
  }

  getErrorSignal(): WritableSignal<string | null> {
    return this.error
  }

  getUpdateSuccessfulSignal(): WritableSignal<boolean> {
    return this.updateSuccessful;
  }

  loadProject(id: string): void {
    this.reset();
    this.http.get<any>(`${this.apiUrl}/${id}`).pipe(
      map((project) => ({
        Id: project.ID, // Mapping der API-Daten auf das Interface
        name: project.Name
      }))
    ).subscribe({
      next: (transformedProject) => {
        this.project.set(transformedProject);
      },
      error: (err) => {
        console.error('Failed to load project:', err);
        this.error.set('Failed to load project. Please try again later.');
      }
    });
  }

  loadProjects(): void {
    this.reset();
    this.http.get<any[]>(this.apiUrl).pipe(
      map((projects) =>
        projects.map((project) => ({
          Id: project.ID,
          name: project.Name
        }))
      ),
      catchError((err) => {
        console.error('Failed to load projects:', err);
        this.error.set('Failed to load projects. Please try again later.');
        return of([]);
      })
    ).subscribe((transformedProjects) => {
      this.projects.set(transformedProjects);
    });
  }

  updateProject(data: Project): void {
    this.reset();
    this.http.put<Project>(`${this.apiUrl}/${data.Id}`, data).pipe(
      tap((updatedProject) => {
        this.lastUpdatedProject.set(updatedProject);
        this.updateSuccessful.set(true);
        this.error.set(null);
      }),
      catchError((err) => {
        const errorMessage = err.error?.message || 'Failed to update project.';
        this.error.set(errorMessage);
        this.updateSuccessful.set(false);
        return of(null as unknown as Project);
      }),
    ).subscribe(() => {
    })
  }

  addProject(data: Project) {
    this.reset();
    this.http.post<Project>(`${this.apiUrl}`, data).pipe(
      tap((addedProject) => {
        this.lastUpdatedProject.set(addedProject);
        this.error.set(null);
        this.updateSuccessful.set(true);
      }),
      catchError((err) => {
        const errorMessage = err.error?.message || 'Failed to add project.';
        this.error.set(errorMessage);
        this.updateSuccessful.set(false);
        return of(null as unknown as Project);
      }),
    ).subscribe(() => {
    })
  }

  deleteProject(data: Project) {
    this.reset();
    this.http.delete<Project>(`${this.apiUrl}/${data.Id}`).pipe(
      tap(() => {
        this.projectDeleted.set(true);
        this.error.set(null);
        this.updateSuccessful.set(true);
      }),
      catchError((err) => {
        const errorMessage = err.error?.message || 'Failed to delete project.';
        this.error.set(errorMessage);
        this.updateSuccessful.set(false);
        return of(null as unknown as Project);
      }),
    ).subscribe(() => {
    })
  }

  reset() {
    this.project.set(null);
    this.projects.set([]);
    this.projectDeleted.set(false);
    this.error.set(null);
    this.updateSuccessful.set(false);
    this.lastUpdatedProject.set(null);
  }
}
