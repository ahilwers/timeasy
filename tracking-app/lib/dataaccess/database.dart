import 'dart:io';

import 'package:path/path.dart';
import 'package:path_provider/path_provider.dart';
import 'package:sqflite/sqflite.dart';
import 'package:sqflite_migration/sqflite_migration.dart';

class DBProvider {
  DBProvider._();

  static final DBProvider dbProvider = DBProvider._();

  static Database? _database;

  final initScript = [
    '''
      CREATE TABLE Projects (
        id TEXT, 
        name TEXT, 
        created INTEGER, 
        updated INTEGER 
      );
    ''',
    '''
      CREATE TABLE TimeEntries (
        id TEXT, 
        startTime INTEGER, 
        endTime INTEGER, 
        description TEXT, 
        created INTEGER, 
        updated INTEGER, 
        projectId TEXT, 
        FOREIGN KEY(projectId) REFERENCES Projects(id) 
      );
    '''
  ];

  final migrations = [
    '''
      ALTER TABLE Projects ADD COLUMN deleted INTEGER DEFAULT 0;
    ''',
    '''
     CREATE TABLE ApiCredentials (
        id TEXT,
        username TEXT, 
        name TEXT,
        email TEXT,
        accessToken TEXT, 
        refreshToken TEXT, 
        logoutUrl TEXT,
        credentialJson TEXT
      );
    ''',
    '''
     CREATE TABLE Settings (
        id TEXT,
        lastSyncTime INTEGER,
        latestRemoteTimeEntryTimestamp INTEGER DEFAULT 0,
        latestRemoteProjectTimestamp INTEGER DEFAULT 0,
        latestLocalTimeEntryTimestamp INTEGER DEFAULT 0,
        latestLocalProjectTimestamp INTEGER DEFAULT 0
      );
    ''',
    '''
      ALTER TABLE TimeEntries ADD COLUMN deleted INTEGER DEFAULT 0;
    '''
  ];

  Future<Database> get database async {
    if (_database == null) {
      _database = await _initDB();
    }
    return _database!;
  }

  _initDB() async {
    Directory documentsDirectory = await getApplicationDocumentsDirectory();
    String path = join(documentsDirectory.path, join("timeasy", "timeasy.db"));

    final migrationConfig = MigrationConfig(
        initializationScript: initScript, migrationScripts: migrations);
    return await openDatabaseWithMigration(path, migrationConfig);
  }
}
