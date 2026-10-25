import {computed, inject, Injectable, signal} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable, catchError, of, tap} from 'rxjs';
import {environment} from '../../environments/environment';

interface BillingStatus {
  enabled: boolean;
}

interface CheckoutResponse {
  enabled: boolean;
  url: string | null;
}

interface SubscriptionStatus {
  enabled: boolean;
  active: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class BillingService {
  private readonly apiUrl = `${environment.apiUrl}`;
  private readonly http = inject(HttpClient);

  private readonly billingEnabledSignal = signal<boolean | null>(null);

  /** Whether billing is enabled on the server (null = not loaded yet). */
  billingEnabled = computed(() => this.billingEnabledSignal());

  /**
   * Loads the (public) billing status so clients can hide billing UI on
   * self-hosted instances. Returns a cold observable — callers must subscribe
   * (or compose it), otherwise no request is made.
   */
  loadBillingStatus(): Observable<BillingStatus> {
    return this.http.get<BillingStatus>(`${this.apiUrl}/billing/status`).pipe(
      tap((status) => this.billingEnabledSignal.set(status.enabled)),
      catchError(() => {
        this.billingEnabledSignal.set(false);
        return of({enabled: false});
      })
    );
  }

  /**
   * Creates a Polar hosted checkout session for the current user and returns its URL.
   * locale is the active UI language (e.g. "de") so the checkout is localized.
   */
  startUpgrade(locale?: string): Observable<CheckoutResponse> {
    return this.http.post<CheckoutResponse>(`${this.apiUrl}/upgrade`, {locale});
  }

  /** Force-refreshes the current user's subscription state (bypasses the server-side cache). */
  getSubscriptionStatus(): Observable<SubscriptionStatus> {
    return this.http.get<SubscriptionStatus>(`${this.apiUrl}/subscription/status`);
  }

  /** Creates a short-lived Polar customer portal session URL for the current user. */
  createPortalSession(): Observable<CheckoutResponse> {
    return this.http.post<CheckoutResponse>(`${this.apiUrl}/billing/portal`, {});
  }
}
