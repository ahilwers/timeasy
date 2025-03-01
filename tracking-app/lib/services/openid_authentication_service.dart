import 'dart:convert';
import 'dart:io';

import 'package:openid_client/openid_client.dart';
import 'package:openid_client/openid_client_io.dart' as io;
import 'package:timeasy/models/api_credentials.dart';
import 'package:url_launcher/url_launcher.dart';

class OpenIdAuthenticationService {
  final scopes = ['profile', 'offline_access'];
  final clientId = 'timeasy-tracking-app';
  final String keycloakUri;

  OpenIdAuthenticationService(this.keycloakUri) {}

  Future<ApiCredentials?> authenticate() async {
    var client = await getClient();
    var authenticator = io.Authenticator(client,
        scopes: scopes, port: 4000, urlLancher: _urlLauncher);
    var c = await authenticator.authorize();
    _closeWebView();

    if (c == null) {
      return null;
    }

    var token = await c.getTokenResponse();
    var userInformation = await c.getUserInfo();

    var apiCredential = ApiCredentials();
    apiCredential.accessToken = token.accessToken;
    apiCredential.refreshToken = token.refreshToken;
    apiCredential.logoutUrl = c.generateLogoutUrl()?.toString();
    apiCredential.email = userInformation.email;
    apiCredential.username = userInformation.preferredUsername;
    apiCredential.name = userInformation.name;
    apiCredential.credentialJson = json.encode(c.toJson());
    return apiCredential;
  }

  Future<void> logout(ApiCredentials apiCredential) async {
    if (apiCredential.logoutUrl != null) {
      _urlLauncher(apiCredential.logoutUrl!);
    }
  }

  Future<bool> refreshToken(ApiCredentials apiCredential) async {
    if (apiCredential.credentialJson == null) {
      return false;
    }
    var credential =
        Credential.fromJson(json.decode(apiCredential.credentialJson!));
    var token = await credential.getTokenResponse();
    if (token.accessToken == null) {
      apiCredential.clear();
      return false;
    }
    apiCredential.accessToken = token.accessToken;
    apiCredential.refreshToken = token.refreshToken;
    apiCredential.credentialJson = json.encode(credential.toJson());
    return true;
  }

  _urlLauncher(String url) async {
    var uri = Uri.parse(url);
    if (await canLaunchUrl(uri) || Platform.isAndroid) {
      await launchUrl(uri);
    } else {
      throw 'Could not launch $url';
    }
  }

  void _closeWebView() {
    if (Platform.isAndroid || Platform.isIOS) {
      closeInAppWebView();
    }
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
