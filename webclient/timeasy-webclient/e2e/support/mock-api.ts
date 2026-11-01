import type { BrowserContext, Route } from '@playwright/test';
import {
  getIsoWeek,
  MockProject,
  MockStore,
  MockTimeEntry,
  newId,
  weekStatistics,
} from './mock-store';

/**
 * Network mocks for the Go API. Routes are handled in the browser context, so
 * no backend, database or Docker is required. Responses carry permissive CORS
 * headers because the Angular dev server (localhost:4200) calls the API on a
 * different origin (localhost:8080).
 */

function corsHeaders(route: Route): Record<string, string> {
  const origin = route.request().headers()['origin'] ?? '*';
  return {
    'access-control-allow-origin': origin,
    'access-control-allow-credentials': 'true',
    'access-control-allow-methods': 'GET,POST,PUT,DELETE,OPTIONS',
    'access-control-allow-headers': 'authorization,content-type,x-requested-with',
    vary: 'Origin',
  };
}

function json(route: Route, body: unknown, status = 200): Promise<void> {
  return route.fulfill({
    status,
    contentType: 'application/json',
    headers: corsHeaders(route),
    body: JSON.stringify(body),
  });
}

function empty(route: Route, status = 204): Promise<void> {
  return route.fulfill({ status, headers: corsHeaders(route) });
}

function readBody(route: Route): Record<string, unknown> {
  try {
    return (route.request().postDataJSON() ?? {}) as Record<string, unknown>;
  } catch {
    return {};
  }
}

function localDateString(isoValue: string): string {
  const date = new Date(isoValue);
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${date.getFullYear()}-${month}-${day}`;
}

function normalizeProject(id: string, body: Record<string, unknown>): MockProject {
  const deadline = body['deadline'];
  return {
    id,
    name: String(body['name'] ?? ''),
    userId: '00000000-0000-0000-0000-00000000e2e1',
    teamId: null,
    color: String(body['color'] ?? '#4F46E5'),
    deadline: typeof deadline === 'string' && deadline.length > 0 ? deadline : null,
    hourlyRate: String(body['hourlyRate'] ?? '0.00'),
    timeBudget: Number(body['timeBudget'] ?? 0),
    isActive: body['isActive'] !== false,
  };
}

function normalizeTimeEntry(id: string, body: Record<string, unknown>): MockTimeEntry {
  return {
    id,
    projectId: String(body['projectId'] ?? ''),
    userId: '00000000-0000-0000-0000-00000000e2e1',
    description: String(body['description'] ?? ''),
    startTime: String(body['startTime'] ?? new Date().toISOString()),
    endTime: body['endTime'] ? String(body['endTime']) : null,
  };
}

export async function installApiMocks(context: BrowserContext, store: MockStore): Promise<void> {
  await context.route('**/api/v1/**', async (route) => {
    const request = route.request();
    if (request.method() === 'OPTIONS') {
      return empty(route);
    }

    const url = new URL(request.url());
    const path = url.pathname.replace(/^\/api\/v1/, '');
    const method = request.method();

    // --- projects ---
    if (path === '/projects' && method === 'GET') {
      return json(route, store.projects);
    }
    if (path === '/projects' && method === 'POST') {
      const project = normalizeProject(newId('e2e-project'), readBody(route));
      store.projects.push(project);
      return json(route, project, 201);
    }

    const projectMatch = path.match(/^\/projects\/([^/]+)$/);
    if (projectMatch) {
      const id = projectMatch[1];
      const index = store.projects.findIndex((project) => project.id === id);
      if (method === 'GET') {
        return index >= 0 ? json(route, store.projects[index]) : json(route, { message: 'not found' }, 404);
      }
      if (method === 'PUT' && index >= 0) {
        store.projects[index] = normalizeProject(id, readBody(route));
        return json(route, store.projects[index]);
      }
      if (method === 'DELETE') {
        if (index >= 0) store.projects.splice(index, 1);
        return empty(route);
      }
    }

    // --- description suggestions / issue resolution ---
    if (/^\/projects\/[^/]+\/descriptions\/suggest$/.test(path)) {
      return json(route, { suggestions: [] });
    }
    if (/^\/projects\/[^/]+\/issues\/resolve$/.test(path)) {
      return json(route, { status: 'pending', key: '' });
    }
    if (/^\/projects\/[^/]+\/external\/.*$/.test(path)) {
      return json(route, { message: 'not configured' }, 404);
    }

    // --- time entries ---
    if (path === '/timeentries' && method === 'GET') {
      const projectId = url.searchParams.get('projectId');
      const startDate = url.searchParams.get('startDate');
      const endDate = url.searchParams.get('endDate');

      const entries = store.timeEntries.filter((entry) => {
        if (projectId && entry.projectId !== projectId) return false;
        const entryDate = localDateString(entry.startTime);
        if (startDate && entryDate < startDate) return false;
        if (endDate && entryDate > endDate) return false;
        return true;
      });
      return json(route, entries.slice().sort((a, b) => b.startTime.localeCompare(a.startTime)));
    }
    if (path === '/timeentries' && method === 'POST') {
      const entry = normalizeTimeEntry(newId('e2e-entry'), readBody(route));
      store.timeEntries.push(entry);
      return json(route, entry, 201);
    }
    if (path === '/timeentries/lastopen' && method === 'GET') {
      const open = store.timeEntries.find((entry) => entry.endTime === null);
      return open ? json(route, open) : json(route, { message: 'not found' }, 404);
    }

    const entryMatch = path.match(/^\/timeentries\/([^/]+)$/);
    if (entryMatch) {
      const id = entryMatch[1];
      const index = store.timeEntries.findIndex((entry) => entry.id === id);
      if (method === 'GET') {
        return index >= 0 ? json(route, store.timeEntries[index]) : json(route, { message: 'not found' }, 404);
      }
      if (method === 'PUT' && index >= 0) {
        store.timeEntries[index] = normalizeTimeEntry(id, readBody(route));
        return json(route, store.timeEntries[index]);
      }
      if (method === 'DELETE') {
        if (index >= 0) store.timeEntries.splice(index, 1);
        return empty(route);
      }
    }

    // --- statistics ---
    if (path === '/currentweeknumber' && method === 'GET') {
      const now = new Date();
      return json(route, { weekNumber: getIsoWeek(now), year: now.getFullYear() });
    }
    const statsMatch = path.match(/^\/weeklystatistics\/(\d+)\/(\d+)$/);
    if (statsMatch) {
      const weekNumber = Number(statsMatch[1]);
      const year = Number(statsMatch[2]);
      return json(route, weekStatistics(store, weekNumber, year));
    }

    // --- billing / external accounts (not used by the suite, but page loads) ---
    if (path === '/billing/status') {
      return json(route, { enabled: false, webApiKey: null });
    }
    if (path === '/subscription/status') {
      return json(route, { active: false });
    }
    if (path === '/user/external-accounts') {
      return json(route, []);
    }

    // --- generic fallbacks so unrelated page loads never fail hard ---
    if (method === 'GET') {
      return json(route, path.split('/').filter(Boolean).length > 1 ? {} : []);
    }
    if (method === 'DELETE') {
      return empty(route);
    }
    return json(route, readBody(route));
  });
}
