import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {map, Observable} from 'rxjs';
import {Project} from '../models/project.model';

@Injectable({
  providedIn: 'root'
})
export class ProjectService {
  private apiUrl = 'http://localhost:8080/api/v1/projects';

  constructor(private http: HttpClient) {}

  getProjects(): Observable<Project[]> {
    return this.http.get<any[]>(this.apiUrl).pipe(
      map((projects) =>
        projects.map((project) => ({
          Id: project.ID,
          name: project.Name
        }))
      )
    );
  }

  createProject(data: Project): Observable<Project> {
    return this.http.post<Project>(this.apiUrl, data);
  }
}
