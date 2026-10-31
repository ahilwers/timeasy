import {computed, inject, Injectable, signal} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import Keycloak from 'keycloak-js';
import {Observable, catchError, firstValueFrom, from, of, tap} from 'rxjs';
import {
  CustomerInfo,
  ErrorCode,
  Offering,
  Package,
  Purchases,
  PurchasesError
} from '@revenuecat/purchases-js';
import {environment} from '../../environments/environment';

interface BillingStatus {
  enabled: boolean;
  webApiKey?: string | null;
}

interface SubscriptionStatus {
  enabled: boolean;
  active: boolean;
}

export type BillingInterval = 'monthly' | 'annual';

export interface BillingPackages {
  monthly: Package | null;
  annual: Package | null;
}

@Injectable({
  providedIn: 'root'
})
export class BillingService {
  private readonly apiUrl = `${environment.apiUrl}`;
  private readonly http = inject(HttpClient);
  private readonly keycloak = inject(Keycloak);

  private readonly billingEnabledSignal = signal<boolean | null>(null);
  private readonly webApiKeySignal = signal<string | null>(null);
  private purchases: Purchases | null = null;

  /** Whether billing is enabled on the server (null = not loaded yet). */
  billingEnabled = computed(() => this.billingEnabledSignal());

  /**
   * Loads the (public) billing status so clients can hide billing UI on
   * self-hosted instances. The response also carries the public RevenueCat web
   * API key when the server has web billing configured, so the key is not
   * baked into this client's build. Returns a cold observable — callers must
   * subscribe (or compose it), otherwise no request is made.
   */
  loadBillingStatus(): Observable<BillingStatus> {
    return this.http.get<BillingStatus>(`${this.apiUrl}/billing/status`).pipe(
      tap((status) => {
        this.billingEnabledSignal.set(status.enabled);
        this.webApiKeySignal.set(status.enabled ? (status.webApiKey ?? null) : null);
      }),
      catchError(() => {
        this.billingEnabledSignal.set(false);
        this.webApiKeySignal.set(null);
        return of({enabled: false});
      })
    );
  }

  /** Force-refreshes the current user's subscription state (bypasses the server-side cache). */
  getSubscriptionStatus(): Observable<SubscriptionStatus> {
    return this.http.get<SubscriptionStatus>(`${this.apiUrl}/subscription/status`);
  }

  /** Loads the monthly/annual packages (including localized prices) of the current offering. */
  async getPackages(): Promise<BillingPackages> {
    const purchases = await this.getPurchases();
    const offerings = await purchases.getOfferings();
    return this.packagesFrom(offerings.current);
  }

  /**
   * Opens the embedded checkout (Paddle via RevenueCat) for the given interval.
   * Resolves with the customer info after a successful purchase.
   */
  async purchase(interval: BillingInterval, locale?: string): Promise<CustomerInfo> {
    const packages = await this.getPackages();
    const rcPackage = packages[interval];
    if (!rcPackage) {
      throw new Error(`RevenueCat offering has no ${interval} package`);
    }
    const result = await (await this.getPurchases()).purchase({
      rcPackage,
      customerEmail: this.customerEmail(),
      selectedLocale: locale,
      defaultLocale: locale,
      skipSuccessPage: true
    });
    return result.customerInfo;
  }

  /**
   * Returns the management URL for the user's active subscription (Paddle
   * customer portal for web purchases, App Store / Play Store for mobile
   * purchases), or null when there is no active subscription.
   */
  getManagementURL(): Observable<string | null> {
    return from(this.loadManagementURL());
  }

  /** True when the user aborted the purchase flow — not an error worth showing. */
  isUserCancelled(error: unknown): boolean {
    return error instanceof PurchasesError && error.errorCode === ErrorCode.UserCancelledError;
  }

  private async loadManagementURL(): Promise<string | null> {
    const purchases = await this.getPurchases();
    const customerInfo = await purchases.getCustomerInfo();
    return customerInfo.managementURL;
  }

  private packagesFrom(offering: Offering | null): BillingPackages {
    return {
      monthly: offering?.monthly ?? null,
      annual: offering?.annual ?? null
    };
  }

  /**
   * Lazily configures the RevenueCat Web SDK exactly once, using the Keycloak
   * user ID as the RevenueCat App User ID. This is the same ID the server uses
   * for entitlement checks and the mobile apps use after login, so a purchase
   * on any platform unlocks the same account.
   */
  /**
   * Lazily configures the RevenueCat Web SDK exactly once, using the Keycloak
   * user ID as the RevenueCat App User ID. The public API key comes from the
   * server's billing status endpoint, so it is not baked into this build.
   */
  private async getPurchases(): Promise<Purchases> {
    if (!this.purchases) {
      let apiKey = this.webApiKeySignal();
      if (!apiKey) {
        // The page may not have loaded the billing status yet (e.g. "manage
        // subscription" clicked immediately after opening the profile).
        await firstValueFrom(this.loadBillingStatus());
        apiKey = this.webApiKeySignal();
      }
      if (!apiKey) {
        throw new Error('RevenueCat web billing is not configured on this server');
      }
      const appUserId = this.keycloak.subject;
      if (!appUserId) {
        throw new Error('RevenueCat requires an authenticated Keycloak user');
      }
      this.purchases = Purchases.configure({apiKey, appUserId});
    }
    return this.purchases;
  }

  private customerEmail(): string | undefined {
    const email = this.keycloak.tokenParsed?.['email'];
    return typeof email === 'string' && email.length > 0 ? email : undefined;
  }
}
