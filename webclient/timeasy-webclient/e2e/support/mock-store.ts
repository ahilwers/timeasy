/**
 * In-memory fixtures for the mocked E2E backend.
 *
 * The store mimics the JSON contract of the Go API closely enough that the
 * existing components work unchanged. It is intentionally small: only the
 * endpoints exercised by the E2E suite are implemented (see mock-api.ts).
 */

export interface MockProject {
  id: string;
  name: string;
  userId: string;
  teamId: string | null;
  color: string;
  deadline: string | null;
  hourlyRate: string;
  timeBudget: number;
  isActive: boolean;
}

export interface MockTimeEntry {
  id: string;
  projectId: string;
  userId: string;
  description: string;
  startTime: string;
  endTime: string | null;
}

export interface MockStore {
  projects: MockProject[];
  timeEntries: MockTimeEntry[];
}

const USER_ID = '00000000-0000-0000-0000-00000000e2e1';

export function getIsoWeek(date: Date): number {
  const d = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
  const dayNum = d.getUTCDay() || 7;
  d.setUTCDate(d.getUTCDate() + 4 - dayNum);
  const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
  return Math.ceil(((d.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
}

function startOfWeek(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay() || 7; // Monday = 1 ... Sunday = 7
  d.setDate(d.getDate() - (day - 1));
  d.setHours(9, 0, 0, 0);
  return d;
}

function addHours(date: Date, hours: number): Date {
  return new Date(date.getTime() + hours * 60 * 60 * 1000);
}

function iso(date: Date): string {
  return date.toISOString();
}

export function createStore(): MockStore {
  const now = new Date();
  const monday = startOfWeek(now);
  const today = new Date(now);
  today.setHours(8, 30, 0, 0);

  return {
    projects: [
      {
        id: 'e2e-project-alpha',
        name: 'Alpha Website',
        userId: USER_ID,
        teamId: null,
        color: '#4F46E5',
        deadline: null,
        hourlyRate: '95.00',
        timeBudget: 40,
        isActive: true,
      },
      {
        id: 'e2e-project-beta',
        name: 'Beta App',
        userId: USER_ID,
        teamId: null,
        color: '#10B981',
        deadline: null,
        hourlyRate: '110.00',
        timeBudget: 0,
        isActive: true,
      },
      {
        id: 'e2e-project-inactive',
        name: 'Gamma Legacy',
        userId: USER_ID,
        teamId: null,
        color: '#F59E0B',
        deadline: null,
        hourlyRate: '80.00',
        timeBudget: 0,
        isActive: false,
      },
    ],
    timeEntries: [
      {
        id: 'e2e-entry-monday',
        projectId: 'e2e-project-alpha',
        userId: USER_ID,
        description: 'Kickoff and planning',
        startTime: iso(addHours(monday, 0)),
        endTime: iso(addHours(monday, 1.5)),
      },
      {
        id: 'e2e-entry-tuesday',
        projectId: 'e2e-project-beta',
        userId: USER_ID,
        description: 'Implement login flow',
        startTime: iso(addHours(monday, 24)),
        endTime: iso(addHours(monday, 26)),
      },
      {
        id: 'e2e-entry-today',
        projectId: 'e2e-project-alpha',
        userId: USER_ID,
        description: 'Review pull requests',
        startTime: iso(today),
        endTime: iso(addHours(today, 1)),
      },
    ],
  };
}

export function newId(prefix: string): string {
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}`;
}

export function weekStatistics(store: MockStore, weekNumber: number, year: number) {
  const weekdays = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];

  // Determine the Monday of the requested ISO week.
  const jan4 = new Date(Date.UTC(year, 0, 4));
  const jan4Day = jan4.getUTCDay() || 7;
  const firstMonday = new Date(jan4);
  firstMonday.setUTCDate(jan4.getUTCDate() - (jan4Day - 1));
  const monday = new Date(firstMonday);
  monday.setUTCDate(firstMonday.getUTCDate() + (weekNumber - 1) * 7);

  const days = weekdays.map((weekday, index) => {
    const dayStart = new Date(monday);
    dayStart.setUTCDate(monday.getUTCDate() + index);
    const dayEnd = new Date(dayStart);
    dayEnd.setUTCDate(dayStart.getUTCDate() + 1);

    const entries = store.timeEntries.filter((entry) => {
      const start = new Date(entry.startTime);
      return start >= dayStart && start < dayEnd;
    });

    const perProject = new Map<string, number>();
    for (const entry of entries) {
      if (!entry.endTime) continue;
      const seconds = Math.round((new Date(entry.endTime).getTime() - new Date(entry.startTime).getTime()) / 1000);
      perProject.set(entry.projectId, (perProject.get(entry.projectId) ?? 0) + seconds);
    }

    return {
      weekday,
      timeInSeconds: [...perProject.values()].reduce((sum, value) => sum + value, 0),
      timesPerProject: [...perProject.entries()].map(([projectId, timeInSeconds]) => {
        const project = store.projects.find((p) => p.id === projectId);
        return {
          projectId,
          projectName: project?.name ?? 'Unknown',
          projectColor: project?.color ?? '#999999',
          timeInSeconds,
        };
      }),
    };
  });

  return {
    weekNumber,
    year,
    firstDay: iso(monday),
    lastDay: iso(new Date(monday.getTime() + 6 * 86400000 + 86399000)),
    days,
    sumInSeconds: days.reduce((sum, day) => sum + day.timeInSeconds, 0),
  };
}
