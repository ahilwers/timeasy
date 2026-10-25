import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:url_launcher/url_launcher.dart';

import 'package:timeasy/environment.dart';

/// Shows a modal paywall informing the user that sync requires an active
/// timeasy Pro subscription. The app stays fully usable locally — only sync
/// is gated.
///
/// Store compliance: App-store builds (iOS/Android) must not direct users to
/// an external purchase flow (Apple Guideline 3.1.1 / Google Play payments
/// policy), so those builds show information only — the subscription is
/// purchased on the web client at timeasy.org. Only dev builds get a direct
/// "Subscribe" button so the full checkout flow can be tested end-to-end.
Future<void> showSubscriptionRequiredDialog(BuildContext context) async {
  final localizations = AppLocalizations.of(context)!;
  final isDark = Theme.of(context).brightness == Brightness.dark;
  final isDevBuild =
      Environment.getEnvironment() == Environment.DEV_ENVIRONMENT;

  await showDialog(
    context: context,
    barrierDismissible: true,
    builder: (dialogContext) => AlertDialog(
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
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(localizations.subscriptionRequiredMessage),
          const SizedBox(height: 12),
          Text(
            localizations.subscriptionManagedExternally,
            style: TextStyle(
              fontSize: 13,
              color: isDark ? Colors.white70 : Colors.black54,
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(dialogContext).pop(),
          child: Text(localizations.subscriptionRequiredLater),
        ),
        if (isDevBuild)
          ElevatedButton(
            onPressed: () async {
              Navigator.of(dialogContext).pop();
              await _openUpgradePage(context);
            },
            child: Text(localizations.subscriptionRequiredSubscribe),
          ),
      ],
    ),
  );
}

/// Opens the web client's upgrade page. Only reachable from dev builds —
/// store builds must not link to the external purchase flow.
Future<void> _openUpgradePage(BuildContext context) async {
  final upgradeUrl = Uri.parse('${Environment.webBaseUrl}/upgrade');
  if (!await launchUrl(upgradeUrl, mode: LaunchMode.externalApplication)) {
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Could not open the website')),
      );
    }
  }
}
