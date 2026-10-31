import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:timeasy/l10n/app_localizations.dart';
import 'package:intl/intl.dart';
import 'package:timeasy/bloc/authentication/authentication_bloc.dart';
import 'package:timeasy/bloc/authentication/authentication_event.dart';
import 'package:timeasy/bloc/authentication/authentication_state.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_bloc.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_state.dart';
import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/bloc/synchronization/synchronization_state.dart';
import 'package:timeasy/repositories/settings_repository.dart';
import 'package:timeasy/services/background_sync_service.dart';
import 'package:timeasy/services/billing_service.dart';
import 'package:timeasy/services/review_service.dart';
import 'package:url_launcher/url_launcher.dart';

class SettingsView extends StatefulWidget {
  @override
  State<StatefulWidget> createState() {
    return SettingsViewState();
  }
}

class SettingsViewState extends State<SettingsView>
    with WidgetsBindingObserver {
  PurchaseOrigin _origin = PurchaseOrigin.none;
  bool _originLoading = true;
  StreamSubscription<PurchaseOrigin>? _originUpdates;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    // RevenueCat pushes the customer info whenever the entitlement changes
    // (login, purchase, restore, renewal, expiry). Apply it directly instead
    // of polling, so the card cannot keep showing a stale status.
    _originUpdates = BillingService().purchaseOriginUpdates.listen((origin) {
      if (!mounted) {
        return;
      }
      if (origin == PurchaseOrigin.none) {
        // The SDK reports no entitlement. The server is the source of truth
        // for gating, so let it confirm before showing "No active
        // subscription" (e.g. accounts granted access without RevenueCat).
        _loadPurchaseOrigin();
        return;
      }
      setState(() {
        _origin = origin;
        _originLoading = false;
      });
    });
    _loadPurchaseOrigin();
  }

  @override
  void dispose() {
    _originUpdates?.cancel();
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    // Coming back from the store's subscription management may have changed
    // the subscription, so refresh the status.
    if (state == AppLifecycleState.resumed) {
      _loadPurchaseOrigin();
    }
  }

  Future<void> _loadPurchaseOrigin() async {
    setState(() {
      _originLoading = true;
    });
    final origin = await BillingService().activePurchaseOrigin();
    if (!mounted) {
      return;
    }
    setState(() {
      _origin = origin;
      _originLoading = false;
    });
  }

  static String? _userIdOf(AuthenticationState state) {
    return state is AuthenticationAuthenticated
        ? state.credentials.userId
        : null;
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(localizations.settings),
      ),
      body: BlocListener<AuthenticationBloc, AuthenticationState>(
        // Logging in or out changes whose subscription is shown; reload the
        // status so the card never keeps the previous state.
        listenWhen: (previous, current) =>
            _userIdOf(previous) != _userIdOf(current),
        listener: (context, state) {
          _loadPurchaseOrigin();
        },
        child: RefreshIndicator(
          onRefresh: _loadPurchaseOrigin,
          child: ListView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(16),
            children: <Widget>[
              BlocBuilder<AuthenticationBloc, AuthenticationState>(
                builder: (context, state) {
                  if (state is AuthenticationAuthenticated) {
                    return _buildAuthenticated(context, localizations, state);
                  }
                  return _buildLoggedOut(context, localizations);
                },
              ),
              // Rating is independent of the account, so it is shown in both
              // states.
              _SettingsCard(
                title: localizations.aboutSection,
                child: _buildRateAppRow(context, localizations),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildAuthenticated(BuildContext context,
      AppLocalizations localizations, AuthenticationAuthenticated state) {
    final theme = Theme.of(context);
    final displayName =
        state.credentials.name ?? state.credentials.username ?? '';
    final email = state.credentials.email;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        _SettingsCard(
          title: localizations.accountSection,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Text(
                localizations.greeting(displayName),
                style: theme.textTheme.titleLarge
                    ?.copyWith(fontWeight: FontWeight.bold),
              ),
              if (email != null && email.isNotEmpty) ...[
                const SizedBox(height: 4),
                Text(
                  email,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
              const SizedBox(height: 16),
              Align(
                alignment: Alignment.centerLeft,
                child: OutlinedButton.icon(
                  onPressed: _logout,
                  icon: const Icon(Icons.logout, size: 18),
                  label: Text(localizations.logout),
                ),
              ),
            ],
          ),
        ),
        _SettingsCard(
          title: localizations.subscriptionSection,
          child: _buildSubscriptionStatus(context, localizations),
        ),
        _SettingsCard(
          title: localizations.syncSection,
          child: _buildSyncStatus(context, localizations),
        ),
      ],
    );
  }

  Widget _buildSubscriptionStatus(
      BuildContext context, AppLocalizations localizations) {
    if (_originLoading) {
      return Row(
        children: <Widget>[
          const SizedBox(
            width: 16,
            height: 16,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
          const SizedBox(width: 8),
          Text(
            localizations.lastSyncLoading,
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ],
      );
    }
    // Store compliance: only store-managed subscriptions may be managed
    // through a link from the app. Web purchases (Paddle on timeasy.org) are
    // valid but must only be described, never linked (Apple 3.1.1 / Google
    // Play payments policy).
    if (_origin == PurchaseOrigin.store) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          _InfoRow(
            icon: Icons.verified,
            color: Colors.green,
            text: localizations.subscriptionActiveStore,
          ),
          const SizedBox(height: 8),
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton.icon(
              onPressed: _openSubscriptionManagement,
              icon: const Icon(Icons.open_in_new, size: 18),
              label: Text(localizations.manageSubscription),
            ),
          ),
        ],
      );
    }
    if (_origin == PurchaseOrigin.testStore) {
      // Simulated store from test builds: show the status, but there is no
      // real store management to link to.
      return _InfoRow(
        icon: Icons.verified,
        color: Colors.green,
        text: localizations.subscriptionActiveStore,
      );
    }
    if (_origin == PurchaseOrigin.web) {
      return _InfoRow(
        icon: Icons.language,
        text: localizations.subscriptionManagedOnWeb,
      );
    }
    if (_origin == PurchaseOrigin.server) {
      // The server confirmed an active subscription, but the RevenueCat SDK
      // could not tell which store it was bought from. Show the status
      // without a management link, which is safe for store compliance.
      return _InfoRow(
        icon: Icons.verified,
        color: Colors.green,
        text: localizations.subscriptionActive,
      );
    }
    return _InfoRow(
      icon: Icons.info_outline,
      text: localizations.subscriptionInactive,
    );
  }

  Widget _buildSyncStatus(
      BuildContext context, AppLocalizations localizations) {
    final theme = Theme.of(context);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        BlocBuilder<InternetConnectionBloc, InternetConnectionState>(
          builder: (context, state) {
            final connected = state is InternetConnectionConnected;
            return _InfoRow(
              icon: connected ? Icons.wifi : Icons.wifi_off,
              color: connected ? Colors.green : theme.colorScheme.error,
              text: connected
                  ? localizations.internetConnected
                  : localizations.internetDisconnected,
            );
          },
        ),
        const SizedBox(height: 8),
        BlocBuilder<SynchronizationBloc, SynchronizationState>(
          builder: (context, state) {
            if (state is SynchronizationInProgress) {
              return _InfoRow(
                icon: Icons.sync,
                color: theme.colorScheme.primary,
                text: localizations.syncing,
              );
            }
            if (state is SynchronizationError) {
              return _InfoRow(
                icon: Icons.error_outline,
                color: theme.colorScheme.error,
                text: localizations.syncError(state.message),
              );
            }
            return FutureBuilder<String>(
              future: _getLastSyncDate(context),
              builder: (context, snapshot) {
                if (snapshot.connectionState == ConnectionState.waiting) {
                  return _InfoRow(
                    icon: Icons.schedule,
                    text: localizations.lastSyncLoading,
                  );
                }
                if (snapshot.hasError) {
                  return _InfoRow(
                    icon: Icons.schedule,
                    text: localizations.lastSyncError,
                  );
                }
                return _InfoRow(
                  icon: Icons.schedule,
                  text: localizations.lastSyncLabel(
                      snapshot.data ?? localizations.lastSyncNever),
                );
              },
            );
          },
        ),
        const SizedBox(height: 8),
        Align(
          alignment: Alignment.centerLeft,
          child: TextButton.icon(
            onPressed: _syncNow,
            icon: const Icon(Icons.sync, size: 18),
            label: Text(localizations.syncNow),
          ),
        ),
      ],
    );
  }

  Widget _buildLoggedOut(
      BuildContext context, AppLocalizations localizations) {
    final theme = Theme.of(context);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        _SettingsCard(
          title: 'timeasy',
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Text(
                localizations.freeUseHeading,
                style: theme.textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.bold,
                  color: theme.colorScheme.primary,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                localizations.freeUseDescription,
                style: theme.textTheme.bodyMedium,
              ),
              const SizedBox(height: 12),
              Text(
                localizations.subscriptionBenefits,
                style: theme.textTheme.bodyMedium,
              ),
            ],
          ),
        ),
        _SettingsCard(
          title: localizations.accountSection,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Text(
                localizations.signInPromo,
                style: theme.textTheme.bodyMedium
                    ?.copyWith(fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 12),
              // The Keycloak user ID is the RevenueCat App User ID, so an
              // account is required before subscribing. Offer both paths
              // directly: sign in or create an account.
              ElevatedButton(
                onPressed: _login,
                child: Text(localizations.signIn),
              ),
              const SizedBox(height: 8),
              OutlinedButton(
                onPressed: _register,
                child: Text(localizations.register),
              ),
            ],
          ),
        ),
      ],
    );
  }

  void _login() {
    context.read<AuthenticationBloc>().add(LoginEvent());
  }

  void _register() {
    context.read<AuthenticationBloc>().add(RegisterEvent());
  }

  void _logout() {
    context.read<AuthenticationBloc>().add(LogoutEvent());
  }

  void _syncNow() {
    context.read<BackgroundSyncService>().synchronize();
  }

  void _rateApp() {
    ReviewService().openStoreListing();
  }

  Widget _buildRateAppRow(
      BuildContext context, AppLocalizations localizations) {
    final theme = Theme.of(context);
    return InkWell(
      onTap: _rateApp,
      borderRadius: BorderRadius.circular(8),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 4, horizontal: 4),
        child: Row(
          children: <Widget>[
            Icon(Icons.star_outline, size: 22, color: theme.colorScheme.primary),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    localizations.rateApp,
                    style: theme.textTheme.titleSmall
                        ?.copyWith(fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    localizations.rateAppDescription,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
            Icon(Icons.chevron_right, color: theme.colorScheme.onSurfaceVariant),
          ],
        ),
      ),
    );
  }

  Future<void> _openSubscriptionManagement() async {
    final localizations = AppLocalizations.of(context)!;
    final url = await BillingService().managementUrl();
    if (!mounted) {
      return;
    }
    if (url == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(localizations.noActiveSubscription)),
      );
      return;
    }
    if (!await launchUrl(Uri.parse(url),
        mode: LaunchMode.externalApplication)) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(localizations.noActiveSubscription)),
      );
    }
  }

  Future<String> _getLastSyncDate(BuildContext context) async {
    final localizations = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    var repository = new SettingsRepository();
    var settings = await repository.getSettings();
    final lastSyncTime = settings.lastSyncTime;
    if (lastSyncTime == null) {
      return localizations.lastSyncNever;
    }
    var formatter = new DateFormat.yMMMd(locale).add_Hm();
    return formatter.format(lastSyncTime.toLocal());
  }
}

/// Card with a small section title, used to structure the settings screen.
class _SettingsCard extends StatelessWidget {
  final String title;
  final Widget child;

  const _SettingsCard({Key? key, required this.title, required this.child})
      : super(key: key);

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      elevation: 0,
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: BorderSide(color: theme.dividerColor.withValues(alpha: 0.4)),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              title.toUpperCase(),
              style: theme.textTheme.labelMedium?.copyWith(
                letterSpacing: 1.0,
                fontWeight: FontWeight.w600,
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 12),
            child,
          ],
        ),
      ),
    );
  }
}

/// Icon + text row used inside the settings cards.
class _InfoRow extends StatelessWidget {
  final IconData icon;
  final String text;
  final Color? color;

  const _InfoRow({Key? key, required this.icon, required this.text, this.color})
      : super(key: key);

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Icon(
          icon,
          size: 18,
          color: color ?? theme.colorScheme.onSurfaceVariant,
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            text,
            style: theme.textTheme.bodyMedium,
          ),
        ),
      ],
    );
  }
}
