import {Component, OnInit} from '@angular/core';
import {Project} from '../../../models/project.model';
import {ProjectService} from '../../../services/project.service';
import {DataView} from 'primeng/dataview';
import {JsonPipe, NgClass} from '@angular/common';
import {Tag} from 'primeng/tag';
import {Button} from 'primeng/button';
import {TableModule} from 'primeng/table';

@Component({
  selector: 'app-project-list',
  standalone: true,
  imports: [DataView, NgClass, Tag, Button, JsonPipe, TableModule],
  templateUrl: './project-list.component.html',
  styleUrl: './project-list.component.css'
})
export class ProjectListComponent implements OnInit {

  projects : Project[] = [];

  constructor(private readonly projectService: ProjectService) {}

  ngOnInit(): void {
    this.loadProjects();
  }

  loadProjects(): void {
    this.projectService.getProjects().subscribe({
      next: (data) => {
        this.projects = data;
        console.log(this.projects)
      },
      error: (err) => {
        console.error('Error fetching projects:', err);
      }
    });
  }

  editProject(project: Project): void {
    console.log('Edit project:', project);
  }

  deleteProject(project: Project): void {
    console.log('Delete project:', project);
  }


}
