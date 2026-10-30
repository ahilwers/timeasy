import 'dart:convert';
import 'dart:io';

import 'package:flutter/foundation.dart' show kDebugMode;
import 'package:flutter/services.dart';
import 'package:purchases_flutter/purchases_flutter.dart';
import 'package:timeasy/environment.dart';
import 'package:timeasy/repositories/api_credentials_repository.dart';
import 'package:timeasy/services/api_service.dart';
import 'package:timeasy/utils/app_logger.dart';

/// Where the active subscription was bought. The distinction matters for App
/// Store compliance: the app may only deep-link into subscription management
/// for store purchases (Apple guideline 3.1.1 / Google Play payments policy).
/// Web purchases (Paddle on timeasy.org) are equally valid subscriptions and
/// unlock Pro the same way, but their management lives outside the stores, so
/// the app never links to it. [testStore] is RevenueCat's simulated store and
/// only occurs in test builds.
enum PurchaseOrigin { store, web, testStore, none }

/// RevenueCat-based billing: Paddle powers web purchases, the App Store and
/// Google Play power in-app purchases. The Go server checks the same RevenueCat
/// entitlement (`timeasy_pro`) using the Keycloak user ID as App User ID, so a
/// purchase on any platform unlocks the same account.
///
/// Billing is strictly opt-in. On self-hosted instances without billing the
/// server reports `enabled: false` and this service never configures the
/// RevenueCat SDK, so no RevenueCat or store calls are made.
class BillingService {
  static const String entitlementId = 'timeasy_pro';

  // Public RevenueCat API keys (per platform, injected at build time with
  // --dart-define). Public by design — they identify the app, not the user.
  static const String _iosApiKey =
      String.fromEnvironment('REVENUECAT_IOS_API_KEY');
  static const String _androidApiKey =
      String.fromEnvironment('REVENUECAT_ANDROID_API_KEY');

  BillingService._internal();
  static final BillingService _instance = BillingService._internal();
  factory BillingService() => _instance;

  bool? _billingEnabled;
  bool _configured = false;
  String? _userId;
  String? _pendingUserId;
  String? _pendingEmail;
  String? _pendingDisplayName;
  Future<void>? _pendingLogin;

  bool get isConfigured => _configured;

  /// Whether the server has billing enabled. Failures are not cached so the
  /// check is retried on the next call.
  Future<bool> isBillingEnabled() async {
    if (_billingEnabled != null) {
      return _billingEnabled!;
    }
    try {
      final api = ApiService(baseUrl: Environment.apiBaseUrl);
      final response = await api.get('/billing/status');
      if (api.isSuccessStatusCode(response.statusCode)) {
        final body = jsonDecode(response.body) as Map<String, dynamic>;
        _billingEnabled = body['enabled'] == true;
        return _billingEnabled!;
      }
      return false;
    } catch (e) {
      AppLogger.w('Could not load billing status', error: e);
      return false;
    }
  }

  /// Configures RevenueCat (once) and logs the authenticated Keycloak user in.
  /// The email and display name from the Keycloak ID token are stored as
  /// customer attributes, so the RevenueCat dashboard shows a readable
  /// customer instead of only the Keycloak user ID.
  ///
  /// Safe to call repeatedly: concurrent calls (auth listener and paywall
  /// retry) are collapsed into a single configuration attempt.
  Future<void> login(String? userId,
      {String? email, String? displayName}) {
    if (userId == null || userId.isEmpty) {
      return Future.value();
    }
    _pendingUserId = userId;
    _pendingEmail = email ?? _pendingEmail;
    _pendingDisplayName = displayName ?? _pendingDisplayName;
    if (_configured && _userId == userId) {
      return _applyCustomerAttributes();
    }
    return _pendingLogin ??= _configure(userId).whenComplete(() {
      _pendingLogin = null;
    });
  }

  Future<void> _configure(String userId) async {
    if (!await isBillingEnabled()) {
      return;
    }
    final apiKey = _platformApiKey();
    if (apiKey.isEmpty) {
      AppLogger.w('RevenueCat API key missing for this platform');
      return;
    }
    try {
      if (!_configured) {
        if (kDebugMode) {
          // Surfaces configuration and offerings problems in the `flutter run`
          // console (e.g. "There is no current offering configured").
          await Purchases.setLogLevel(LogLevel.debug);
        }
        final configuration = PurchasesConfiguration(apiKey)
          ..appUserID = userId;
        await Purchases.configure(configuration);
        _configured = true;
      } else {
        await Purchases.logIn(userId);
      }
      _userId = userId;
      await _applyCustomerAttributes();
      final customerInfo = await Purchases.getCustomerInfo();
      final entitlements = customerInfo.entitlements.active;
      AppLogger.i(
          'RevenueCat configured for user $userId (active entitlements: '
          '${entitlements.isEmpty ? "none" : entitlements.map((id, info) => MapEntry(id, info.store.toString()))})');
    } catch (e, stackTrace) {
      AppLogger.e('Failed to configure RevenueCat',
          error: e, stackTrace: stackTrace);
    }
  }

  /// Sends the Keycloak email/display name to RevenueCat so the dashboard and
  /// store receipts show a human readable customer. Best effort: an attribute
  /// failure must never block billing.
  Future<void> _applyCustomerAttributes() async {
    if (!_configured) {
      return;
    }
    try {
      final email = _pendingEmail;
      if (email != null && email.isNotEmpty) {
        await Purchases.setEmail(email);
      }
      final displayName = _pendingDisplayName;
      if (displayName != null && displayName.isNotEmpty) {
        await Purchases.setDisplayName(displayName);
      }
    } catch (e) {
      AppLogger.w('Could not set RevenueCat customer attributes', error: e);
    }
  }

  /// Waits for an in-flight configuration or retries it, so callers that run
  /// right after login (paywall, settings) never see a false "not configured"
  /// state.
  Future<void> _ensureConfigured() async {
    if (_configured) {
      return;
    }
    final pending = _pendingLogin;
    if (pending != null) {
      await pending;
      return;
    }
    final userId = _pendingUserId;
    if (userId != null) {
      await login(userId);
    }
  }

  Future<void> logout() async {
    if (!_configured) {
      return;
    }
    try {
      await Purchases.logOut();
    } catch (e) {
      AppLogger.w('RevenueCat logout failed', error: e);
    }
    _userId = null;
  }

  /// Loads the current offering including localized prices. Returns null when
  /// billing is not configured or the offerings cannot be loaded.
  ///
  /// Configures RevenueCat on demand when the paywall opens before the auth
  /// listener finished configuring it (sync starts immediately after login).
  Future<Offerings?> loadOfferings() async {
    if (!_configured) {
      await _ensureConfigured();
    }
    if (!_configured) {
      AppLogger.w('RevenueCat is not configured; cannot load offerings '
          '(check the REVENUECAT_*_API_KEY build define and '
          '/billing/status on ${Environment.apiBaseUrl})');
      return null;
    }
    try {
      return await Purchases.getOfferings();
    } catch (e, stackTrace) {
      AppLogger.e('Failed to load RevenueCat offerings',
          error: e, stackTrace: stackTrace);
      return null;
    }
  }

  /// Purchases the given package. Returns true when the entitlement is active.
  ///
  /// The store purchase can succeed while RevenueCat is still processing the
  /// transaction, so the returned customer info may not contain the
  /// entitlement yet (the dialog then stayed open until "restore purchases").
  /// Poll the customer info briefly before reporting a failure.
  Future<bool> purchase(Package package) async {
    if (!_configured) {
      await _ensureConfigured();
    }
    final result = await Purchases.purchase(PurchaseParams.package(package));
    if (_hasActiveEntitlement(result.customerInfo)) {
      await _refreshServerSubscription();
      return true;
    }
    for (var attempt = 0; attempt < 4; attempt++) {
      await Future.delayed(Duration(milliseconds: 500 * (attempt + 1)));
      await Purchases.invalidateCustomerInfoCache();
      final customerInfo = await Purchases.getCustomerInfo();
      if (_hasActiveEntitlement(customerInfo)) {
        await _refreshServerSubscription();
        return true;
      }
    }
    return false;
  }

  /// Restores previous purchases. Returns true when the entitlement is active.
  Future<bool> restorePurchases() async {
    if (!_configured) {
      await _ensureConfigured();
    }
    if (!_configured) {
      return false;
    }
    try {
      final customerInfo = await Purchases.restorePurchases();
      if (_hasActiveEntitlement(customerInfo)) {
        await _refreshServerSubscription();
        return true;
      }
      return false;
    } catch (e) {
      // RevenueCat's Test Store has no store receipts to restore; fall back to
      // the current customer info so an active entitlement still unlocks.
      AppLogger.w('Restore purchases failed, reading customer info',
          error: e);
      final customerInfo = await Purchases.getCustomerInfo();
      if (_hasActiveEntitlement(customerInfo)) {
        await _refreshServerSubscription();
        return true;
      }
      return false;
    }
  }

  /// Forces the server to re-check the entitlement (bypassing its cache) after
  /// a successful purchase or restore. Without this the sync that runs right
  /// after paying can still be rejected with a cached 402, which made the
  /// paywall pop up again immediately.
  Future<void> _refreshServerSubscription() async {
    try {
      final credentials = await ApiCredentialsRepository().getApiCredentials();
      final token = credentials.accessToken;
      if (token == null || token.isEmpty) {
        return;
      }
      final api = ApiService(baseUrl: Environment.apiBaseUrl, token: token);
      await api.get('/subscription/status');
    } catch (e) {
      AppLogger.w('Could not refresh server subscription status', error: e);
    }
  }

  /// Management URL for the active subscription (App Store / Play Store), or
  /// null when there is no store-managed subscription or billing is not
  /// configured. Web purchases (Paddle) are deliberately excluded: launching
  /// their web management portal from the app would violate Apple guideline
  /// 3.1.1 / Google Play payments policy.
  Future<String?> managementUrl() async {
    if (!_configured) {
      await _ensureConfigured();
    }
    if (!_configured) {
      return null;
    }
    try {
      final customerInfo = await Purchases.getCustomerInfo();
      if (_originOf(customerInfo) != PurchaseOrigin.store) {
        return null;
      }
      return customerInfo.managementURL;
    } catch (e) {
      AppLogger.w('Could not read subscription management URL', error: e);
      return null;
    }
  }

  /// How the active `timeasy_pro` entitlement was purchased, or
  /// [PurchaseOrigin.none] when there is none or billing is unavailable.
  /// Used by the settings view to decide what may be shown and linked.
  Future<PurchaseOrigin> activePurchaseOrigin() async {
    if (!_configured) {
      await _ensureConfigured();
    }
    if (!_configured) {
      return PurchaseOrigin.none;
    }
    try {
      // Bypass the SDK cache: right after a purchase it can still hold the
      // pre-purchase customer info, which made the settings screen show "no
      // subscription" until it was reopened.
      await Purchases.invalidateCustomerInfoCache();
      final customerInfo = await Purchases.getCustomerInfo();
      final origin = _originOf(customerInfo);
      AppLogger.d('Active purchase origin: $origin');
      return origin;
    } catch (e) {
      AppLogger.w('Could not read purchase origin', error: e);
      return PurchaseOrigin.none;
    }
  }

  /// True when the user aborted the purchase sheet.
  bool isPurchaseCancelled(Object error) {
    if (error is PlatformException) {
      return PurchasesErrorHelper.getErrorCode(error) ==
          PurchasesErrorCode.purchaseCancelledError;
    }
    return false;
  }

  /// True when the purchase was accepted but is still pending (e.g. cash
  /// payment methods on Google Play). The entitlement activates once the
  /// payment completes, so the UI should inform instead of reporting an error.
  bool isPurchasePending(Object error) {
    if (error is PlatformException) {
      return PurchasesErrorHelper.getErrorCode(error) ==
          PurchasesErrorCode.paymentPendingError;
    }
    return false;
  }

  bool _hasActiveEntitlement(CustomerInfo customerInfo) {
    return customerInfo.entitlements.active.containsKey(entitlementId);
  }

  PurchaseOrigin _originOf(CustomerInfo customerInfo) {
    final entitlement = customerInfo.entitlements.active[entitlementId];
    if (entitlement == null) {
      return PurchaseOrigin.none;
    }
    switch (entitlement.store) {
      case Store.appStore:
      case Store.macAppStore:
      case Store.playStore:
      case Store.amazon:
      case Store.galaxy:
        return PurchaseOrigin.store;
      case Store.testStore:
        // RevenueCat's simulated store, used in test builds only.
        return PurchaseOrigin.testStore;
      case Store.paddle:
      case Store.rcBilling:
      case Store.stripe:
        return PurchaseOrigin.web;
      default:
        return PurchaseOrigin.none;
    }
  }

  String _platformApiKey() {
    if (Platform.isIOS) {
      return _iosApiKey;
    }
    if (Platform.isAndroid) {
      return _androidApiKey;
    }
    return '';
  }
}
