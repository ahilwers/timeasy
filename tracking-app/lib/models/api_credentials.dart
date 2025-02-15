import 'package:uuid/uuid.dart';

class ApiCredentials {
  static final String tableName = "ApiCredentials";
  static final String idColumn = "id";
  static final String usernameColumn = "username";
  static final String nameColumn = "name";
  static final String emailColumn = "email";
  static final String accessTokenColumn = "accessToken";
  static final String refreshTokenColumn = "refreshToken";
  static final String logoutUrlColumn = "logoutUrl";
  static final String credentialJsonColumn = "credentialJson";

  late String id;
  String? username;
  String? name;
  String? email;
  String? accessToken;
  String? refreshToken;
  String? logoutUrl;
  String? credentialJson;

  ApiCredentials() {
    var uuid = new Uuid();
    id = uuid.v4();
  }

  void clear() {
    username = null;
    name = null;
    email = null;
    accessToken = null;
    refreshToken = null;
    logoutUrl = null;
    credentialJson = null;
  }

  ApiCredentials.fromMap(Map<String, dynamic> map) {
    id = map[idColumn];
    name = map[nameColumn];
    username = map[usernameColumn];
    email = map[emailColumn];
    accessToken = map[accessTokenColumn];
    refreshToken = map[refreshTokenColumn];
    logoutUrl = map[logoutUrlColumn];
    credentialJson = map[credentialJsonColumn];
  }

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      idColumn: id,
      nameColumn: name,
      usernameColumn: username,
      emailColumn: email,
      accessTokenColumn: accessToken,
      refreshTokenColumn: refreshToken,
      logoutUrlColumn: logoutUrl,
      credentialJsonColumn: credentialJson
    };
  }
}
