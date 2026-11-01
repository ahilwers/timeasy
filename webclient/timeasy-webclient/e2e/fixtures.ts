import { test as base, expect } from '@playwright/test';
import { installApiMocks } from './support/mock-api';
import { installAuthMocks } from './support/mock-auth';
import { createStore, MockStore } from './support/mock-store';

/**
 * E2E modes:
 *  - mock (default): API and Keycloak are intercepted in the browser context.
 *    No backend, database or Docker required. Used by CI.
 *  - live: real Go API + real Keycloak. Opt in with `E2E_MODE=live` for
 *    local integration checks (see auth.setup.ts).
 */
export const isMockMode = (process.env.E2E_MODE ?? 'mock') !== 'live';

export const test = base.extend<{ mockStore: MockStore }>({
  mockStore: [
    async ({ context }, use) => {
      if (isMockMode) {
        const store = createStore();
        await installAuthMocks(context);
        await installApiMocks(context, store);
        await use(store);
        return;
      }
      await use(createStore());
    },
    { auto: true },
  ],
});

export { expect };
