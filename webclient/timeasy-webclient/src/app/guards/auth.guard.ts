import { inject } from '@angular/core';
import { ActivatedRouteSnapshot, CanActivateFn, Router, RouterStateSnapshot, UrlTree } from '@angular/router';
import { AuthGuardData, createAuthGuard } from 'keycloak-angular';
import Keycloak from 'keycloak-js';

const isAccessAllowed = async (
    route: ActivatedRouteSnapshot,
    _: RouterStateSnapshot,
    authData: AuthGuardData
): Promise<boolean | UrlTree> => {
    const { authenticated, grantedRoles } = authData;
    const accessAllowed = authenticated;

    if (!authenticated) {
        const keycloak = inject(Keycloak);
        const redirectUrl = `${window.location.origin}/${route.url}`;
        keycloak.login({ redirectUri: redirectUrl });
    }

    const requiredRole = route.data['role'];
    if (requiredRole) {
        const hasRequiredRole = (role: string): boolean =>
            Object.values(grantedRoles.realmRoles).some((roles) => roles.includes(role));
        const accessAllowed = (authenticated && hasRequiredRole(requiredRole));
        if (!accessAllowed) {
            const router = inject(Router);
            return router.parseUrl('/forbidden');
        }
        return accessAllowed;
    }

    if (authenticated) {
        return true;
    }
    return accessAllowed;
};

export const canActivateAuth = createAuthGuard<CanActivateFn>(isAccessAllowed);
