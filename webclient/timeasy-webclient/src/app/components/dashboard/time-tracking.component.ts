import { Component, effect, inject, OnInit, signal, ViewChild } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';
import { Button } from 'primeng/button';
import { InputText } from 'primeng/inputtext';
import { Select } from 'primeng/select';
import { TimeEntry } from '../../models/timeentry.model';
import { ProjectService } from '../../services/project.service';
import { TimeEntryService } from '../../services/time-entry.service';
import { TimerComponent } from './timer.component';
import { DescriptionAutocompleteComponent } from '../shared/description-autocomplete/description-autocomplete.component';

@Component({
    selector: 'app-time-tracking',
    standalone: true,
    templateUrl: './time-tracking.component.html',
    styleUrls: ['./time-tracking.component.css'],
    imports: [
        Button,
        Select,
        TranslatePipe,
        TimerComponent,
        FormsModule,
        DescriptionAutocompleteComponent
    ]
})

export class TimeTrackingComponent implements OnInit {
    private readonly projectService = inject(ProjectService);
    private readonly timeEntryService = inject(TimeEntryService);
    private debounceTimer: any;

    @ViewChild('timerRef') timerComponent!: TimerComponent;

    projects = this.projectService.activeProjects
    project = this.projectService.project()
    lastOpenTimeEntry = this.timeEntryService.lastOpenTimeEntry()
    addedTimeEntry = this.timeEntryService.lastUpdatedTimeEntry()

    selectedProject: any = null;
    timeEntry: TimeEntry | null = null;
    description = signal<string>('')
    descriptionManuallyChanged = signal<boolean>(false)
    running: boolean = false;

    constructor() {
        effect(() => {
            if (this.lastOpenTimeEntry()) {
                this.timeEntry = this.lastOpenTimeEntry();
                if (this.timeEntry) {
                    this.descriptionManuallyChanged.set(false);
                    this.description.set(this.timeEntry.description);
                    this.projectService.loadProject(this.timeEntry.projectId);
                    this.timerComponent?.start(new Date(this.timeEntry.startTime));
                }
                this.running = true;
            }
        });
        effect(() => {
            const projectData = this.project();
            if (projectData && !this.selectedProject) {
                this.selectedProject = projectData;
            }
        });
        effect(() => {
            if (this.addedTimeEntry()) {
                this.timeEntryService.loadLastOpenTimeEntry()
            }
        });
        effect(() => {
            if (this.descriptionManuallyChanged()) {
                const value = this.description();
                clearTimeout(this.debounceTimer);
                this.debounceTimer = setTimeout(() => {
                    this.saveDescription(value);
                }, 500);
            }
        });
    }

    ngOnInit(): void {
        this.projectService.loadProjects();
        this.timeEntryService.loadLastOpenTimeEntry();
    }

    onPlay() {
        if (this.selectedProject) {
            const timeEntry: TimeEntry = {
                id: '',
                projectId: this.selectedProject.id,
                startTime: new Date(),
                endTime: undefined,
                description: this.description()
            }
            this.timeEntryService.addTimeEntry(timeEntry)
            this.running = true;
            this.timerComponent?.start();
        }
    }

    onPause() {
        if (this.timeEntry) {
            this.timeEntry.endTime = new Date();
            this.timeEntryService.updateTimeEntry(this.timeEntry);
        }
        this.running = false;
        this.timerComponent.stop();
        this.timerComponent.reset();
    }

    projectSelected(event: any) {
        this.selectedProject = event.value;
        if (this.running && this.timeEntry) {
            this.timeEntry.projectId = this.selectedProject.id;
            this.timeEntryService.updateTimeEntry(this.timeEntry);
        }
    }

    saveDescription(description: string) {
        if (this.timeEntry) {
            this.timeEntry.description = description;
            this.timeEntryService.updateTimeEntry(this.timeEntry);
        }
    }

    onDescriptionChanged(newValue: string) {
        this.description.set(newValue);
        this.descriptionManuallyChanged.set(true);
    }
}
