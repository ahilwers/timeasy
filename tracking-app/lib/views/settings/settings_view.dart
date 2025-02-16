import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:timeasy/bloc/authentication/authentication_bloc.dart';
import 'package:timeasy/bloc/authentication/authentication_event.dart';
import 'package:timeasy/bloc/authentication/authentication_state.dart';
import 'package:timeasy/bloc/internetconnection/internetconnection_block.dart';
import 'package:timeasy/bloc/internetconnection/internetconnection_state.dart';

class SettingsView extends StatefulWidget {
  @override
  State<StatefulWidget> createState() {
    return SettingsViewState();
  }
}

class SettingsViewState extends State<SettingsView> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Settings'),
        backgroundColor: Theme.of(context).primaryColor,
      ),
      body: Center(
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: 20.0), // Abstand vom Rand
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            // Elemente mittig ausrichten
            crossAxisAlignment: CrossAxisAlignment.center,
            children: <Widget>[
              BlocBuilder<InternetConnectionBloc, InternetConnectionState>(
                  builder: (context, state) {
                if (state is InternetConnectionConnected) {
                  return Text(
                    "Internet Connected",
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                  );
                } else {
                  return Text(
                    "Internet Disconnected",
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                  );
                }
              }),
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
                        SizedBox(height: 20), // Abstand
                        ElevatedButton(
                          child: Text('Logout'),
                          onPressed: () => _logout(),
                          style: ElevatedButton.styleFrom(
                            padding: EdgeInsets.symmetric(
                                horizontal: 30, vertical: 15),
                            textStyle: TextStyle(fontSize: 16),
                          ),
                        ),
                        ElevatedButton(
                          onPressed: () => _refreshToken(),
                          child: Text("Refresh Token"),
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

  _refreshToken() {
    context.read<AuthenticationBloc>().add(RefreshTokenEvent());
  }
}
