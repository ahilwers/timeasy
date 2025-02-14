import 'package:flutter/material.dart';

import '../../tools/openid_authorizer.dart';

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
        title: Text('Setting'),
        backgroundColor: Theme.of(context).primaryColor,
      ),
      body: Column(
        children: <Widget>[
          Text('Settings'),
          ElevatedButton(
            child: Text('Login'),
            onPressed: () => _login(),
          ),
        ],
      ),
    );
  }

  void _login() {
    var authenticator =
        new OpenIdAuthorizer("http://localhost:8180/realms/timeasy");
    authenticator.authenticate();
  }
}
