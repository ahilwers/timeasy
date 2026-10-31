import {Component, OnDestroy, OnInit, inject, signal} from '@angular/core';
import {TranslateModule} from '@ngx-translate/core';
import {ButtonModule} from 'primeng/button';
import {ProgressSpinnerModule} from 'primeng/progressspinner';
import {BillingService} from '../../../services/billing.service';
import {CardComponent} from '../../card/card.component';

@Component({
  selector: 'app-upgrade-success',
  standalone: true,
  imports: [TranslateModule, CardComponent, ButtonModule, ProgressSpinnerModule],
  templateUrl: './upgrade-success.component.html',
  styleUrl: './upgrade-success.component.css'
})
export class UpgradeSuccessComponent implements OnInit, OnDestroy {
  private readonly billingService = inject(BillingService);

  private static readonly MAX_CHECK_ATTEMPTS = 10;
  private static readonly CHECK_INTERVAL_MS = 3000;

  isLoading = signal(true);
  isActive = signal(false);
  checkFailed = signal(false);

  private checkAttempts = 0;
  private retryTimer: ReturnType<typeof setTimeout> | null = null;

  ngOnInit(): void {
    this.checkSubscription();
  }

  ngOnDestroy(): void {
    this.cancelRetry();
  }

  checkSubscription(): void {
    this.cancelRetry();
    this.isLoading.set(true);
    this.checkFailed.set(false);
    this.checkAttempts++;
    this.billingService.getSubscriptionStatus().subscribe({
      next: (status) => {
        this.isLoading.set(false);
        if (status.active) {
          this.isActive.set(true);
          return;
        }
        this.scheduleRetry();
      },
      error: () => {
        this.isLoading.set(false);
        this.scheduleRetry();
      }
    });
  }

  private scheduleRetry(): void {
    if (this.checkAttempts >= UpgradeSuccessComponent.MAX_CHECK_ATTEMPTS) {
      this.checkFailed.set(true);
      return;
    }
    this.retryTimer = setTimeout(() => this.checkSubscription(), UpgradeSuccessComponent.CHECK_INTERVAL_MS);
  }

  private cancelRetry(): void {
    if (this.retryTimer !== null) {
      clearTimeout(this.retryTimer);
      this.retryTimer = null;
    }
  }
}
