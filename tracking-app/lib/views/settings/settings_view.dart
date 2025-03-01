import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';
import 'package:timeasy/bloc/authentication/authentication_bloc.dart';
import 'package:timeasy/bloc/authentication/authentication_event.dart';
import 'package:timeasy/bloc/authentication/authentication_state.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_bloc.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_state.dart';
import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/bloc/synchronization/synchronization_state.dart';
import 'package:timeasy/repositories/settings_repository.dart';

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

    return Scaffold(
      appBar: AppBar(
        title: Text('Settings'),
        backgroundColor: Theme.of(context).primaryColor,
      ),
      body: Center(
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: 20.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            // Elemente mittig ausrichten
            crossAxisAlignment: CrossAxisAlignment.center,
            children: <Widget>[
              BlocBuilder<AuthenticationBloc, AuthenticationState>(
                builder: (context, state) {
                  if (state is AuthenticationAuthenticated) {
                    return Column(
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
                      ],
                    );
                  } else {
                    return ElevatedButton(
                      child: Text('Login'),
                      onPressed: () => _login(),
                      style: ElevatedButton.styleFrom(
                        padding:
                            EdgeInsets.symmetric(horizontal: 30, vertical: 15),
                        textStyle: TextStyle(fontSize: 16),
                      ),
                    );
                  }
                },
              ),
              SizedBox(height: 20),
              BlocBuilder<InternetConnectionBloc, InternetConnectionState>(
                  builder: (context, state) {
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
                          return Text("Letzte Synchronisierung: Lädt...");
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

  Future<String> _getLastSyncDate() async {
    var repository = new SettingsRepository();
    var settings = await repository.getSettings();
    var formatter = new DateFormat.yMd(_locale.toString()).add_Hm();
    return settings.lastSyncTime == null
        ? ""
        : formatter.format(settings.lastSyncTime!.toLocal());
  }
}
