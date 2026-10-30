import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { TranslateModule } from '@ngx-translate/core';
import { of } from 'rxjs';
import { BillingService } from '../../../services/billing.service';
import { UpgradeComponent } from './upgrade.component';

describe('UpgradeComponent', () => {
  function setup(active: boolean) {
    const router = { navigate: jasmine.createSpy('navigate') };
    const billingService = {
      billingEnabled: signal(true),
      loadBillingStatus: () => of({ enabled: true }),
      getSubscriptionStatus: () => of({ enabled: true, active }),
      getPackages: () => Promise.resolve({ monthly: null, annual: null }),
      purchase: jasmine.createSpy('purchase').and.returnValue(Promise.resolve()),
      isUserCancelled: () => false
    };

    TestBed.configureTestingModule({
      imports: [UpgradeComponent, TranslateModule.forRoot()],
      providers: [
        provideNoopAnimations(),
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
    expect(fixture.componentInstance.isLoading()).toBeFalse();
  });
});
