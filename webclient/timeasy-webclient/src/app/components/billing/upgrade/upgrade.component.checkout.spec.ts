import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { provideTranslateService, TranslateLoader } from '@ngx-translate/core';
import { of } from 'rxjs';
import { BillingService } from '../../../services/billing.service';
import { UpgradeComponent } from './upgrade.component';

// The switch ships disabled; these tests turn the checkout on to exercise the
// enabled path, where a failing Paddle checkout has to fall back to the iOS
// notice instead of showing an error text.
describe('UpgradeComponent with web checkout enabled', () => {
    function setup(purchase: () => Promise<unknown>) {
        const router = { navigate: vi.fn().mockName('navigate') };
        const billingService = {
            billingEnabled: signal(true),
            loadBillingStatus: () => of({ enabled: true }),
            getSubscriptionStatus: () => of({ enabled: true, active: false }),
            getPackages: () => Promise.resolve({ monthly: null, annual: null }),
            purchase: vi.fn().mockName('purchase').mockImplementation(purchase),
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
        fixture.componentInstance.checkoutEnabled.set(true);
        fixture.componentInstance.checkoutUnavailable.set(false);
        return { fixture, router, billingService };
    }

    it('does not show the iOS notice while checkout is enabled', () => {
        const { fixture } = setup(() => Promise.resolve(undefined));

        fixture.detectChanges();

        expect(fixture.componentInstance.checkoutEnabled()).toBe(true);
        expect(fixture.nativeElement.querySelector('.upgrade-notice')).toBeNull();
    });

    it('falls back to the iOS notice when the checkout cannot be created', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => undefined);
        const { fixture, router, billingService } = setup(() => Promise.reject(new Error('transaction_creation_blocked')));
        fixture.detectChanges();

        fixture.componentInstance.startCheckout('monthly');
        await fixture.whenStable();
        fixture.detectChanges();

        expect(billingService.purchase).toHaveBeenCalledWith('monthly', undefined);
        expect(consoleError).toHaveBeenCalled();
        expect(fixture.componentInstance.checkoutUnavailable()).toBe(true);
        const link = fixture.nativeElement.querySelector('.upgrade-store-link') as HTMLAnchorElement;
        expect(link.getAttribute('href')).toBe('https://apps.apple.com/app/id1620757668');
        expect(router.navigate).not.toHaveBeenCalledWith(['/upgrade/success']);
        expect(fixture.nativeElement.textContent).not.toContain('transaction_creation_blocked');
        consoleError.mockRestore();
    });
});
