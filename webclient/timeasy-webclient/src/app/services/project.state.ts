import {signal} from '@angular/core';
import {Project} from '../models/project.model';

export class ProjectState {
  readonly project = signal<Project | null>(null);
  readonly projects = signal<Project[]>([]);
  readonly projectDeleted = signal<boolean>(false);
  readonly error = signal<string | null>(null);
  readonly updateSuccessful = signal<boolean>(false);
  readonly lastUpdatedProject = signal<Project | null>(null);

  reset() {
    this.project.set(null);
    this.projects.set([]);
    this.projectDeleted.set(false);
    this.error.set(null);
    this.updateSuccessful.set(false);
    this.lastUpdatedProject.set(null);
  }

}
