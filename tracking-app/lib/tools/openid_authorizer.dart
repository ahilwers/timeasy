import 'dart:io';

import 'package:openid_client/openid_client.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:openid_client/openid_client_io.dart' as io;


class OpenIdAuthorizer {
  final scopes = ['profile'];
  final clientId = 'timeasy-tracking-app';
  final String keycloakUri;

  OpenIdAuthorizer(this.keycloakUri) {
  }

  Future<Credential> authenticate() async {
    // create a function to open a browser with an url
    urlLauncher(String url) async {
      var uri = Uri.parse(url);
      if (await canLaunchUrl(uri) || Platform.isAndroid) {
        await launchUrl(uri);
      } else {
        throw 'Could not launch $url';
      }
    }

    var client = await getClient();

    // create an authenticator
    var authenticator = io.Authenticator(client,
        scopes: scopes, port: 4000, urlLancher: urlLauncher);

    // starts the authentication
    var c = await authenticator.authorize();

    // close the webview when finished
    if (Platform.isAndroid || Platform.isIOS) {
      closeInAppWebView();
    }

    return c;
  }

  Future<Client> getClient() async {
    var uri = Uri.parse(keycloakUri);
    var issuer = await Issuer.discover(uri);
    return Client(issuer, clientId);
  }

  Future<Credential?> getRedirectResult(Client client,
      {List<String> scopes = const []}) async {
    return null;
  }
}
