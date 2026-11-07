import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, provideRouter, Router } from '@angular/router';
import Keycloak from 'keycloak-js';
import { canActivateAuth } from './auth.guard';

describe('Website authentication entry points', () => {
    function setup(action: string, locale: string, authenticated = false) {
        const keycloak = {
            authenticated,
            register: vi.fn().mockResolvedValue(undefined),
            login: vi.fn().mockResolvedValue(undefined),
        };
        TestBed.configureTestingModule({
            providers: [provideRouter([]), { provide: Keycloak, useValue: keycloak }],
        });
        const route = new ActivatedRouteSnapshot();
        route.data = { authAction: action };
        route.queryParams = { locale };
        const runGuard = () => TestBed.runInInjectionContext(
            () => canActivateAuth(route, TestBed.inject(Router).routerState.snapshot),
        );
        return { keycloak, runGuard };
    }

    it.each(['de', 'en'])('opens registration in %s with the app root as callback', async (locale) => {
        const { keycloak, runGuard } = setup('register', locale);
        expect(await runGuard()).toBe(false);
        expect(keycloak.register).toHaveBeenCalledWith({ redirectUri: `${window.location.origin}/`, locale });
        expect(keycloak.login).not.toHaveBeenCalled();
    });

    it('opens login for existing users', async () => {
        const { keycloak, runGuard } = setup('login', 'en');
        expect(await runGuard()).toBe(false);
        expect(keycloak.login).toHaveBeenCalledWith({ redirectUri: `${window.location.origin}/`, locale: 'en' });
        expect(keycloak.register).not.toHaveBeenCalled();
    });

    it('uses the Keycloak default language for an unsupported locale', async () => {
        const { keycloak, runGuard } = setup('register', 'invalid');
        await runGuard();
        expect(keycloak.register).toHaveBeenCalledWith({ redirectUri: `${window.location.origin}/`, locale: undefined });
    });

    it.each(['register', 'login'])('sends authenticated visitors from /%s to the dashboard', async (action) => {
        const { keycloak, runGuard } = setup(action, 'de', true);
        expect(await runGuard()).toEqual(TestBed.inject(Router).parseUrl('/'));
        expect(keycloak.register).not.toHaveBeenCalled();
        expect(keycloak.login).not.toHaveBeenCalled();
    });
});
