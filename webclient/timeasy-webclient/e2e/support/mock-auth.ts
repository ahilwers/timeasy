import type { BrowserContext, Route } from '@playwright/test';

/**
 * Mocks the parts of Keycloak that keycloak-js v26 talks to.
 *
 * keycloak-js no longer persists tokens in localStorage; on every full page
 * load it performs a `check-sso` flow. We intercept the authorization redirect
 * and answer it with a code, then serve fake JWTs from the token endpoint.
 * The nonce from the authorization request is echoed into the id_token because
 * keycloak-js validates it.
 */

const REALM_URL = 'http://localhost:8180/realms/timeasy';
const CLIENT_ID = 'timeasy-webclient';
const USER = {
  id: '00000000-0000-0000-0000-00000000e2e1',
  username: 'user1',
  firstName: 'E2E',
  lastName: 'User',
  email: 'user1@example.com',
};

function base64url(value: string): string {
  return Buffer.from(value, 'utf8')
    .toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

function jwt(payload: Record<string, unknown>): string {
  return `${base64url(JSON.stringify({ alg: 'none', typ: 'JWT' }))}.${base64url(JSON.stringify(payload))}.`;
}

function tokens(nonce: string) {
  const now = Math.floor(Date.now() / 1000);
  const base = {
    exp: now + 3600,
    iat: now,
    auth_time: now - 60,
    jti: `e2e-jti-${now}`,
    iss: REALM_URL,
    sub: USER.id,
    azp: CLIENT_ID,
    sid: 'e2e-session',
    session_state: 'e2e-session',
    acr: '1',
    email_verified: true,
    name: `${USER.firstName} ${USER.lastName}`,
    preferred_username: USER.username,
    given_name: USER.firstName,
    family_name: USER.lastName,
    email: USER.email,
    realm_access: { roles: [] as string[] },
    resource_access: {},
    scope: 'openid profile email',
  };

  return {
    access_token: jwt({ ...base, typ: 'Bearer', aud: 'account' }),
    refresh_token: jwt({ ...base, typ: 'Refresh' }),
    id_token: jwt({ ...base, typ: 'ID', aud: CLIENT_ID, nonce }),
    expires_in: 3600,
    refresh_expires_in: 86400,
    token_type: 'Bearer',
    'not-before-policy': 0,
    session_state: 'e2e-session',
    scope: 'openid profile email',
  };
}

function corsHeaders(route: Route): Record<string, string> {
  const origin = route.request().headers()['origin'] ?? '*';
  return {
    'access-control-allow-origin': origin,
    'access-control-allow-credentials': 'true',
    'access-control-allow-methods': 'GET,POST,OPTIONS',
    'access-control-allow-headers': 'authorization,content-type',
    vary: 'Origin',
  };
}

function json(route: Route, body: unknown): Promise<void> {
  return route.fulfill({
    status: 200,
    contentType: 'application/json',
    headers: corsHeaders(route),
    body: JSON.stringify(body),
  });
}

export async function installAuthMocks(context: BrowserContext): Promise<void> {
  // authorization code -> nonce (needed for the id_token)
  const codes = new Map<string, string>();

  await context.route('**/realms/timeasy/protocol/openid-connect/auth**', async (route) => {
    const url = new URL(route.request().url());
    const state = url.searchParams.get('state') ?? '';
    const nonce = url.searchParams.get('nonce') ?? '';
    const redirectUri = url.searchParams.get('redirect_uri') ?? 'http://localhost:4200/';

    const code = `e2e-code-${Math.random().toString(36).slice(2, 12)}`;
    codes.set(code, nonce);

    // keycloak-js defaults to response_mode=fragment, so the authorization
    // response must be returned in the URL hash (not as query params).
    const responseMode = url.searchParams.get('response_mode') ?? 'fragment';
    const target = new URL(redirectUri);
    const params = new URLSearchParams({ state, code });
    if (responseMode === 'fragment') {
      target.hash = params.toString();
    } else {
      target.searchParams.set('state', state);
      target.searchParams.set('code', code);
    }

    await route.fulfill({ status: 302, headers: { location: target.toString() } });
  });

  await context.route('**/realms/timeasy/protocol/openid-connect/token**', async (route) => {
    const body = new URLSearchParams(route.request().postData() ?? '');
    const code = body.get('code') ?? '';
    const nonce = codes.get(code) ?? '';
    codes.delete(code);
    return json(route, tokens(nonce));
  });

  // keycloak-js resolves the user profile via {realm}/account
  await context.route('**/realms/timeasy/account', async (route) =>
    json(route, {
      id: USER.id,
      username: USER.username,
      firstName: USER.firstName,
      lastName: USER.lastName,
      email: USER.email,
      emailVerified: true,
      attributes: { locale: ['en'] },
    }),
  );

  await context.route('**/realms/timeasy/protocol/openid-connect/userinfo**', async (route) =>
    json(route, {
      sub: USER.id,
      preferred_username: USER.username,
      given_name: USER.firstName,
      family_name: USER.lastName,
      email: USER.email,
      email_verified: true,
    }),
  );

  await context.route('**/realms/timeasy/protocol/openid-connect/logout**', async (route) => {
    const url = new URL(route.request().url());
    const redirect =
      url.searchParams.get('post_logout_redirect_uri') ??
      url.searchParams.get('redirect_uri') ??
      'http://localhost:4200/';
    await route.fulfill({ status: 302, headers: { location: redirect } });
  });
}
