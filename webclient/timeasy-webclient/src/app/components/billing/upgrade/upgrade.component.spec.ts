import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { provideTranslateService, TranslateLoader } from '@ngx-translate/core';
import { of } from 'rxjs';
import { BillingService } from '../../../services/billing.service';
import { UpgradeComponent } from './upgrade.component';

describe('UpgradeComponent', () => {
    function setup(active: boolean) {
        const router = { navigate: vi.fn().mockName('navigate') };
        const billingService = {
            billingEnabled: signal(true),
            loadBillingStatus: () => of({ enabled: true }),
            getSubscriptionStatus: () => of({ enabled: true, active }),
            getPackages: () => Promise.resolve({ monthly: null, annual: null }),
            purchase: vi.fn().mockName('purchase').mockResolvedValue(undefined),
            isUserCancelled: () => false
        };

        TestBed.configureTestingModule({
            imports: [UpgradeComponent],
            providers: [
                provideNoopAnimations(),
                provideTranslateService({ loader: { provide: TranslateLoader, useValue: { getTranslation: () => of({}) } } }),
                { provide: Router, useValue: router },
                { provide: BillingService, useValue: billingService }
            ]
        });

        const fixture = TestBed.createComponent(UpgradeComponent);
        return { fixture, router };
    }

    it('redirects subscribers to the dashboard instead of showing the paywall', () => {
        const { fixture, router } = setup(true);

        fixture.detectChanges();

        expect(router.navigate).toHaveBeenCalledWith(['/']);
    });

    it('shows the paywall for users without a subscription', () => {
        const { fixture, router } = setup(false);

        fixture.detectChanges();

        expect(router.navigate).not.toHaveBeenCalled();
        expect(fixture.componentInstance.isLoading()).toBe(false);
    });

    it('points to the iOS app instead of opening an unavailable web checkout', () => {
        const { fixture } = setup(false);

        fixture.detectChanges();

        const notice = fixture.nativeElement.querySelector('.upgrade-notice');
        const link = fixture.nativeElement.querySelector('.upgrade-store-link') as HTMLAnchorElement;
        expect(notice).not.toBeNull();
        expect(link.getAttribute('href')).toBe('https://apps.apple.com/app/id1620757668');
        expect(link.getAttribute('target')).toBe('_blank');
        expect(fixture.componentInstance.checkoutEnabled()).toBe(false);
    });
});
