import {Component, OnInit, computed, inject, signal} from '@angular/core';
import {Router} from '@angular/router';
import {TranslateModule, TranslateService} from '@ngx-translate/core';
import {ButtonModule} from 'primeng/button';
import {ProgressSpinnerModule} from 'primeng/progressspinner';
import {catchError, forkJoin, of} from 'rxjs';
import {BillingInterval, BillingService} from '../../../services/billing.service';
import {CardComponent} from '../../card/card.component';

@Component({
  selector: 'app-upgrade',
  standalone: true,
  imports: [TranslateModule, CardComponent, ButtonModule, ProgressSpinnerModule],
  templateUrl: './upgrade.component.html',
  styleUrl: './upgrade.component.css'
})
export class UpgradeComponent implements OnInit {
  private readonly billingService = inject(BillingService);
  private readonly router = inject(Router);
  private readonly translateService = inject(TranslateService);

  billingEnabled = computed(() => this.billingService.billingEnabled());
  isLoading = signal(true);
  purchasingInterval = signal<BillingInterval | null>(null);
  monthlyPrice = signal<string | null>(null);
  annualPrice = signal<string | null>(null);
  errorKey = signal<string | null>(null);

  ngOnInit(): void {
    this.isLoading.set(true);
    forkJoin({
      billing: this.billingService.loadBillingStatus(),
      // Errors are treated as "not subscribed" so the paywall still renders.
      subscription: this.billingService.getSubscriptionStatus().pipe(catchError(() => of(null)))
    }).subscribe(({billing, subscription}) => {
      // Users who already have an active subscription (including self-hosted
      // instances, where every user counts as active) must not see the paywall.
      // This happens e.g. after logging out of an unsubscribed account and back
      // in as a subscribed one — the /upgrade URL is preserved by the login flow.
      if (subscription?.active) {
        this.router.navigate(['/']);
        return;
      }
      this.isLoading.set(false);
      if (billing.enabled) {
        this.loadPrices();
      }
    });
  }

  startCheckout(interval: BillingInterval): void {
    this.purchasingInterval.set(interval);
    this.errorKey.set(null);
    // Pass the active UI language so the checkout is localized (Paddle falls
    // back to its default for unsupported locales).
    this.billingService.purchase(interval, this.translateService.currentLang).then(() => {
      this.router.navigate(['/upgrade/success']);
    }).catch((error) => {
      this.purchasingInterval.set(null);
      if (this.billingService.isUserCancelled(error)) {
        return;
      }
      this.errorKey.set('license.upgrade.checkoutError');
    });
  }

  private loadPrices(): void {
    this.billingService.getPackages().then(({monthly, annual}) => {
      this.monthlyPrice.set(monthly?.webBillingProduct.price.formattedPrice ?? null);
      this.annualPrice.set(annual?.webBillingProduct.price.formattedPrice ?? null);
    }).catch(() => {
      // Prices are cosmetic; purchasing surfaces a real error if the offering
      // cannot be loaded.
    });
  }
}
