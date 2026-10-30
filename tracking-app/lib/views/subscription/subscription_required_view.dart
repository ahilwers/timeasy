import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:purchases_flutter/purchases_flutter.dart';
import 'package:timeasy/services/billing_service.dart';

/// Shows a modal paywall informing the user that sync requires an active
/// timeasy Pro subscription. The app stays fully usable locally — only sync
/// is gated.
///
/// Store compliance: the subscription is bought with native in-app purchase
/// (StoreKit / Google Play Billing via RevenueCat). The dialog never links to
/// an external purchase flow (Apple Guideline 3.1.1 / Google Play payments
/// policy). Resolves to true when a purchase or a restore activated the
/// entitlement.
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
      final active = await _billing.purchase(package);
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
    } catch (e) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isBusy = false;
        _error = _billing.isPurchaseCancelled(e)
            ? null
            : AppLocalizations.of(context)!.subscriptionPurchaseFailed;
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

    return AlertDialog(
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
      content: SingleChildScrollView(
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
                _packageButton(localizations.subscriptionMonthly, monthly),
              if (monthly != null && annual != null) const SizedBox(height: 8),
              if (annual != null)
                _packageButton(localizations.subscriptionYearly, annual),
              if (monthly == null && annual == null)
                Text(
                  localizations.subscriptionNoOfferings,
                  style: TextStyle(fontSize: 13, color: subtleTextColor),
                ),
              if (_isBusy) ...[
                const SizedBox(height: 12),
                const LinearProgressIndicator(),
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
                style: TextStyle(fontSize: 12, color: subtleTextColor),
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
    );
  }

  Widget _packageButton(String label, Package package) {
    return SizedBox(
      width: double.infinity,
      child: ElevatedButton(
        onPressed: _isBusy ? null : () => _purchase(package),
        child: Text('$label · ${package.storeProduct.priceString}'),
      ),
    );
  }
}
