import 'dart:convert';
import 'dart:io';

import 'package:flutter/services.dart';
import 'package:purchases_flutter/purchases_flutter.dart';
import 'package:timeasy/environment.dart';
import 'package:timeasy/services/api_service.dart';
import 'package:timeasy/utils/app_logger.dart';

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
  Future<void> login(String? userId) async {
    if (userId == null || userId.isEmpty) {
      return;
    }
    if (_configured && _userId == userId) {
      return;
    }
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
        final configuration = PurchasesConfiguration(apiKey)
          ..appUserID = userId;
        await Purchases.configure(configuration);
        _configured = true;
      } else {
        await Purchases.logIn(userId);
      }
      _userId = userId;
      AppLogger.i('RevenueCat configured for user $userId');
    } catch (e, stackTrace) {
      AppLogger.e('Failed to configure RevenueCat',
          error: e, stackTrace: stackTrace);
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
  Future<Offerings?> loadOfferings() async {
    if (!_configured) {
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
  Future<bool> purchase(Package package) async {
    final result = await Purchases.purchase(PurchaseParams.package(package));
    return _hasActiveEntitlement(result.customerInfo);
  }

  /// Restores previous purchases. Returns true when the entitlement is active.
  Future<bool> restorePurchases() async {
    if (!_configured) {
      return false;
    }
    final customerInfo = await Purchases.restorePurchases();
    return _hasActiveEntitlement(customerInfo);
  }

  /// Management URL for the active subscription (App Store / Play Store), or
  /// null when there is no active subscription or billing is not configured.
  Future<String?> managementUrl() async {
    if (!_configured) {
      return null;
    }
    try {
      final customerInfo = await Purchases.getCustomerInfo();
      return customerInfo.managementURL;
    } catch (e) {
      AppLogger.w('Could not read subscription management URL', error: e);
      return null;
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

  bool _hasActiveEntitlement(CustomerInfo customerInfo) {
    return customerInfo.entitlements.active.containsKey(entitlementId);
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
