import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { provideTranslateService, TranslateLoader } from '@ngx-translate/core';
import { of } from 'rxjs';
import { TimeEntryService } from '../../../services/time-entry.service';
import { WeeklyStatisticsService } from '../../../services/weekly-statistcs.service';
import { ProjectService } from '../../../services/project.service';
import { DynamicLocaleService } from '../../../services/dynamic-locale.service';
import { WeeklyStatistics } from '../../../models/weekly-statistics.model';
import { WeeklyOverviewComponent } from './weekly-overview.component';

describe('WeeklyOverviewComponent', () => {
  // The first day of the week is sent by the server as UTC midnight
  // (2026-11-02 is a Monday, so 2026-11-08 is the Sunday of that week):
  const statistics = signal<WeeklyStatistics>({
    weekNumber: 45,
    year: 2026,
    firstDay: new Date('2026-11-02T00:00:00Z'),
    lastDay: new Date('2026-11-08T00:00:00Z'),
    days: [
        { weekday: 'monday', timeInSeconds: 0 },
        { weekday: 'tuesday', timeInSeconds: 0 },
        { weekday: 'wednesday', timeInSeconds: 0 },
        { weekday: 'thursday', timeInSeconds: 0 },
        { weekday: 'friday', timeInSeconds: 0 },
        { weekday: 'saturday', timeInSeconds: 0 },
        { weekday: 'sunday', timeInSeconds: 0 },
    ],
    sumInSeconds: 0,
  });

  function setup(selectedProjectId?: string) {
    const router = { navigate: vi.fn().mockName('navigate') };
    const timeEntryService = {
      selectDateRange: vi.fn().mockName('selectDateRange'),
      selectProject: vi.fn().mockName('selectProject'),
    };
    const weeklyStatisticsService = {
      // Mimic the real service, whose exposed signals return the state
      // signals themselves:
      weeklyStatistics: () => statistics,
      currentWeekNumber: () => signal(45),
      error: () => signal(null),
      load: vi.fn().mockName('load'),
      loadCurrentWeekNumber: vi.fn().mockName('loadCurrentWeekNumber'),
      reset: vi.fn().mockName('reset'),
    };
    const projectService = {
      projects: () => signal([]),
      loadProjects: vi.fn().mockName('loadProjects'),
    };

    TestBed.configureTestingModule({
        imports: [WeeklyOverviewComponent],
        providers: [
            provideNoopAnimations(),
            provideTranslateService({ loader: { provide: TranslateLoader, useValue: { getTranslation: () => of({}) } } }),
            { provide: Router, useValue: router },
            { provide: TimeEntryService, useValue: timeEntryService },
            { provide: WeeklyStatisticsService, useValue: weeklyStatisticsService },
            { provide: ProjectService, useValue: projectService },
            // The pipes used by the template resolve the locale lazily via
            // Keycloak - not needed for these tests:
            { provide: DynamicLocaleService, useValue: { getCurrentLocale: () => 'en-US' } },
        ]
    });

    const fixture = TestBed.createComponent(WeeklyOverviewComponent);
    fixture.componentInstance.selectedProjectId.set(selectedProjectId);
    return { component: fixture.componentInstance, router, timeEntryService };
  }

  it('derives the day dates from the first day of the week, sunday being the last day', () => {
    const { component } = setup();

    expect(component.getDateForDayIndex(0)).toEqual(new Date(2026, 10, 2));
    expect(component.getDateForDayIndex(5)).toEqual(new Date(2026, 10, 7));
    expect(component.getDateForDayIndex(6)).toEqual(new Date(2026, 10, 8));
  });

  it('opens the time entry list filtered to the clicked day', () => {
    const { component, router, timeEntryService } = setup();

    component.openDay(statistics().days[0]);

    const expectedDay = new Date(2026, 10, 2);
    expect(timeEntryService.selectDateRange).toHaveBeenCalledWith([expectedDay, expectedDay]);
    expect(router.navigate).toHaveBeenCalledWith(['/timeentries']);
  });

  it('passes the project selected in the weekly view to the time entry list', () => {
    const { component, timeEntryService } = setup('project-1');

    component.openDay(statistics().days[2]);

    const expectedDay = new Date(2026, 10, 4);
    expect(timeEntryService.selectDateRange).toHaveBeenCalledWith([expectedDay, expectedDay]);
    expect(timeEntryService.selectProject).toHaveBeenCalledWith('project-1');
  });

  it('clears the project filter of the time entry list when the weekly view shows all projects', () => {
    const { component, timeEntryService } = setup(undefined);

    component.openDay(statistics().days[6]);

    expect(timeEntryService.selectProject).toHaveBeenCalledWith(undefined);
  });
});
