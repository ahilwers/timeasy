import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:intl/intl.dart';
import 'package:timeasy/bloc/authentication/authentication_bloc.dart';
import 'package:timeasy/bloc/authentication/authentication_event.dart';
import 'package:timeasy/bloc/authentication/authentication_state.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_bloc.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_state.dart';
import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/bloc/synchronization/synchronization_state.dart';
import 'package:timeasy/repositories/settings_repository.dart';
import 'package:url_launcher/url_launcher.dart';

class SettingsView extends StatefulWidget {
  @override
  State<StatefulWidget> createState() {
    return SettingsViewState();
  }
}

class SettingsViewState extends State<SettingsView> {
  Locale? _locale;

  @override
  Widget build(BuildContext context) {
    _locale = Localizations.localeOf(context);
    final localizations = AppLocalizations.of(context)!;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      appBar: AppBar(
        title: Text(localizations.settings),
      ),
      body: SingleChildScrollView(
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: 20.0, vertical: 24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.start,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              BlocBuilder<AuthenticationBloc, AuthenticationState>(
                builder: (context, state) {
                  if (state is AuthenticationAuthenticated) {
                    return Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Text(
                          "Hallo ${state.credentials.name ?? state.credentials.username}!",
                          style: TextStyle(
                              fontSize: 20, fontWeight: FontWeight.bold),
                        ),
                        SizedBox(height: 20),
                        ElevatedButton(
                          child: Text('Logout'),
                          onPressed: () => _logout(),
                          style: ElevatedButton.styleFrom(
                            padding: EdgeInsets.symmetric(
                                horizontal: 30, vertical: 15),
                            textStyle: TextStyle(fontSize: 16),
                          ),
                        ),
                        SizedBox(height: 20),
                        BlocBuilder<InternetConnectionBloc,
                            InternetConnectionState>(builder: (context, state) {
                          if (state is InternetConnectionConnected) {
                            return Text(
                              "Internet Connected",
                            );
                          } else {
                            return Text(
                              "Internet Disconnected",
                            );
                          }
                        }),
                        SizedBox(height: 20),
                        BlocBuilder<SynchronizationBloc, SynchronizationState>(
                          builder: (context, state) {
                            if (state is SynchronizationSuccess ||
                                state is SynchronizationInitial) {
                              return FutureBuilder<String>(
                                future: _getLastSyncDate(),
                                builder: (context, snapshot) {
                                  if (snapshot.connectionState ==
                                      ConnectionState.waiting) {
                                    return Text(
                                        "Letzte Synchronisierung: Lädt...");
                                  } else if (snapshot.hasError) {
                                    return Text("Fehler beim Laden des Datums");
                                  } else {
                                    return Text(
                                        "Letzte Synchronisierung: ${snapshot.data ?? ''}");
                                  }
                                },
                              );
                            } else if (state is SynchronizationError) {
                              return Text(
                                "Fehler während der Synchronisierung: ${state.message}",
                              );
                            } else {
                              return Text("");
                            }
                          },
                        ),
                      ],
                    );
                  } else {
                    // Promotional content for non-logged in users
                    return Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Heading with emphasis
                        Text(
                          localizations.freeUseHeading,
                          style: TextStyle(
                            fontSize: 20,
                            fontWeight: FontWeight.bold,
                            color: Theme.of(context).primaryColor,
                          ),
                        ),
                        SizedBox(height: 12),

                        // Description text
                        Text(
                          localizations.freeUseDescription,
                          style: TextStyle(
                            fontSize: 16,
                            color: isDark ? Colors.white : Colors.black87,
                          ),
                        ),
                        SizedBox(height: 24),

                        // Subscription benefits
                        Text(
                          localizations.subscriptionBenefits,
                          style: TextStyle(
                            fontSize: 16,
                            color: isDark ? Colors.white : Colors.black87,
                          ),
                        ),
                        SizedBox(height: 24),

                        // Call to action with website link
                        GestureDetector(
                          onTap: () => _launchWebsite(),
                          child: Text(
                            localizations.signInPromo,
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              color: Theme.of(context).primaryColor,
                              decoration: TextDecoration.underline,
                            ),
                          ),
                        ),
                        SizedBox(height: 32),

                        // Already have account section
                        Text(
                          localizations.alreadyHaveAccount,
                          style: TextStyle(
                            fontSize: 16,
                            color: isDark ? Colors.white : Colors.black87,
                          ),
                        ),
                        SizedBox(height: 8),
                        GestureDetector(
                          onTap: () => _login(),
                          child: Text(
                            localizations.signInHere,
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              color: Theme.of(context).primaryColor,
                              decoration: TextDecoration.underline,
                            ),
                          ),
                        ),
                      ],
                    );
                  }
                },
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _login() {
    context.read<AuthenticationBloc>().add(LoginEvent());
  }

  void _logout() {
    context.read<AuthenticationBloc>().add(LogoutEvent());
  }

  void _launchWebsite() async {
    final Uri url = Uri.parse('https://timeasy.org');
    if (!await launchUrl(url)) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Could not open website')),
      );
    }
  }

  Future<String> _getLastSyncDate() async {
    var repository = new SettingsRepository();
    var settings = await repository.getSettings();
    var formatter = new DateFormat.yMd(_locale.toString()).add_Hm();
    return settings.lastSyncTime == null
        ? ""
        : formatter.format(settings.lastSyncTime!.toLocal());
  }
}
