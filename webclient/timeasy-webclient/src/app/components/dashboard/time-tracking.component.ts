import {Component, inject, OnInit} from '@angular/core';
import {FloatLabel} from 'primeng/floatlabel';
import {InputText} from 'primeng/inputtext';
import {Button} from 'primeng/button';
import {Select} from 'primeng/select';
import {TranslatePipe} from '@ngx-translate/core';
import {ProjectService} from '../../services/project.service';
import {TimerComponent} from './timer.component';

@Component({
  selector: 'app-time-tracking',
  standalone: true,
  templateUrl: './time-tracking.component.html',
  styleUrls: ['./time-tracking.component.css'],
  imports: [
    FloatLabel,
    InputText,
    Button,
    Select,
    TranslatePipe,
    TimerComponent
  ]
})

export class TimeTrackingComponent implements OnInit {
  private readonly projectService = inject(ProjectService);

  projects = this.projectService.projects();
  selectedProject: any = null;

  ngOnInit(): void {
    this.projectService.loadProjects();
  }

  onPlay() {

  }

  projectSelected(event: any) {
    this.selectedProject = event.value;
  }
}
