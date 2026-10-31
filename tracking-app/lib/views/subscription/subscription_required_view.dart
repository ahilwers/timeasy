import 'dart:async';

import 'package:flutter/material.dart';
import 'package:timeasy/l10n/app_localizations.dart';
import 'package:purchases_flutter/purchases_flutter.dart';
import 'package:timeasy/environment.dart';
import 'package:timeasy/services/billing_service.dart';
import 'package:url_launcher/url_launcher.dart';

/// Shows a modal paywall informing the user that sync requires an active
/// timeasy Pro subscription. The app stays fully usable locally — only sync
/// is gated.
///
/// Store compliance: the subscription is bought with native in-app purchase
/// (StoreKit / Google Play Billing via RevenueCat). The dialog never links to
/// an external purchase flow (Apple Guideline 3.1.1 / Google Play payments
/// policy) — the only links are the legally required terms and privacy pages.
/// Resolves to true when a purchase or a restore activated the entitlement.
Future<bool> showSubscriptionRequiredDialog(BuildContext context) async {
  final purchased = await showDialog<bool>(
    context: context,
    barrierDismissible: true,
    builder: (_) => const _SubscriptionRequiredDialog(),
  );
  return purchased ?? false;
}

class _SubscriptionRequiredDialog extends StatefulWidget {
  const _SubscriptionRequiredDialog();

  @override
  State<_SubscriptionRequiredDialog> createState() =>
      _SubscriptionRequiredDialogState();
}

class _SubscriptionRequiredDialogState
    extends State<_SubscriptionRequiredDialog> {
  final BillingService _billing = BillingService();

  bool _isLoading = true;
  bool _isBusy = false;
  Offerings? _offerings;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadOfferings();
  }

  Future<void> _loadOfferings() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    final enabled = await _billing.isBillingEnabled();
    Offerings? offerings;
    if (enabled) {
      offerings = await _billing.loadOfferings();
    }
    if (!mounted) {
      return;
    }
    setState(() {
      _offerings = offerings;
      _isLoading = false;
    });
  }

  Future<void> _purchase(Package package) async {
    setState(() {
      _isBusy = true;
      _error = null;
    });
    try {
      // The store purchase sheet and RevenueCat processing normally take a
      // few seconds. Never leave the user without feedback if it hangs.
      final active = await _billing
          .purchase(package)
          .timeout(const Duration(minutes: 2));
      if (!mounted) {
        return;
      }
      if (active) {
        Navigator.of(context).pop(true);
        return;
      }
      setState(() {
        _isBusy = false;
        _error = AppLocalizations.of(context)!.subscriptionPurchaseFailed;
      });
    } on TimeoutException {
      if (!mounted) {
        return;
      }
      setState(() {
        _isBusy = false;
        _error = AppLocalizations.of(context)!.subscriptionPurchaseTimeout;
      });
    } catch (e) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isBusy = false;
        if (_billing.isPurchaseCancelled(e)) {
          _error = null;
        } else if (_billing.isPurchasePending(e)) {
          // Google Play pending purchases (e.g. cash payment): the
          // entitlement activates once the payment completes.
          _error = AppLocalizations.of(context)!.subscriptionPurchasePending;
        } else {
          _error = AppLocalizations.of(context)!.subscriptionPurchaseFailed;
        }
      });
    }
  }

  Future<void> _restore() async {
    setState(() {
      _isBusy = true;
      _error = null;
    });
    try {
      final active = await _billing.restorePurchases();
      if (!mounted) {
        return;
      }
      if (active) {
        Navigator.of(context).pop(true);
        return;
      }
      setState(() {
        _isBusy = false;
        _error = AppLocalizations.of(context)!.subscriptionRestoreFailed;
      });
    } catch (e) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isBusy = false;
        _error = AppLocalizations.of(context)!.subscriptionRestoreFailed;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context)!;
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final monthly = _offerings?.current?.monthly;
    final annual = _offerings?.current?.annual;
    final subtleTextColor = isDark ? Colors.white70 : Colors.black54;

    return PopScope(
      // Keep the dialog open while the store transaction is in flight.
      canPop: !_isBusy,
      child: AlertDialog(
        backgroundColor: isDark ? Colors.grey[900] : Colors.white,
        title: Row(
          children: [
            Icon(
              Icons.workspace_premium,
              color: Theme.of(context).primaryColor,
            ),
            const SizedBox(width: 8),
            Expanded(child: Text(localizations.subscriptionRequiredTitle)),
          ],
        ),
        content: _isBusy
            ? Padding(
                padding: const EdgeInsets.symmetric(vertical: 24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const SizedBox(
                      width: 36,
                      height: 36,
                      child: CircularProgressIndicator(),
                    ),
                    const SizedBox(height: 16),
                    Text(
                      localizations.subscriptionProcessing,
                      textAlign: TextAlign.center,
                      style: const TextStyle(fontWeight: FontWeight.w600),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      localizations.subscriptionProcessingHint,
                      textAlign: TextAlign.center,
                      style: TextStyle(fontSize: 13, color: subtleTextColor),
                    ),
                  ],
                ),
              )
            : SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(localizations.subscriptionRequiredMessage),
                    const SizedBox(height: 16),
                    if (_isLoading)
                      const Center(
                        child: Padding(
                          padding: EdgeInsets.all(16),
                          child: CircularProgressIndicator(),
                        ),
                      )
                    else ...[
                      if (monthly != null)
                        _packageButton(
                          localizations.subscriptionPriceMonthly(
                              monthly.storeProduct.priceString),
                          monthly,
                        ),
                      if (monthly != null && annual != null)
                        const SizedBox(height: 8),
                      if (annual != null)
                        _packageButton(
                          localizations.subscriptionPriceYearly(
                              annual.storeProduct.priceString),
                          annual,
                        ),
                      if (monthly == null && annual == null) ...[
                        Text(
                          localizations.subscriptionNoOfferings,
                          style: TextStyle(fontSize: 13, color: subtleTextColor),
                        ),
                        TextButton(
                          onPressed: _isBusy ? null : _loadOfferings,
                          child: Text(localizations.subscriptionRetry),
                        ),
                      ],
                      if (_error != null) ...[
                        const SizedBox(height: 12),
                        Text(
                          _error!,
                          style: TextStyle(
                            fontSize: 13,
                            color: Theme.of(context).colorScheme.error,
                          ),
                        ),
                      ],
                      const SizedBox(height: 12),
                      Text(
                        localizations.subscriptionAutoRenewNote,
                        style: TextStyle(
                            fontSize: 12, color: subtleTextColor),
                      ),
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 16,
                        children: [
                          _legalLink(
                            localizations.subscriptionPrivacy,
                            Environment.privacyPolicyUrl,
                          ),
                          if (Environment.termsOfUseUrl != null)
                            _legalLink(
                              localizations.subscriptionTerms,
                              Environment.termsOfUseUrl!,
                            ),
                        ],
                      ),
                    ],
                  ],
                ),
              ),
        actions: [
          TextButton(
            onPressed: _isBusy ? null : _restore,
            child: Text(localizations.subscriptionRestore),
          ),
          TextButton(
            onPressed: _isBusy ? null : () => Navigator.of(context).pop(false),
            child: Text(localizations.subscriptionRequiredLater),
          ),
        ],
      ),
    );
  }

  Widget _packageButton(String label, Package package) {
    return SizedBox(
      width: double.infinity,
      child: ElevatedButton(
        onPressed: _isBusy ? null : () => _purchase(package),
        child: Text(label),
      ),
    );
  }

  /// Legal links (privacy policy / terms of use) are always allowed, even
  /// though the paywall must not link to any purchase flow.
  Widget _legalLink(String label, String url) {
    return InkWell(
      onTap: () =>
          launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 12,
          color: Theme.of(context).primaryColor,
          decoration: TextDecoration.underline,
        ),
      ),
    );
  }
}
