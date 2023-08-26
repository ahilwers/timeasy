import 'dart:io';

import 'package:path/path.dart';
import 'package:sqflite/sqflite.dart';
import 'package:sqflite_migration/sqflite_migration.dart';

import 'package:path_provider/path_provider.dart';

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

    final migrationConfig = MigrationConfig(initializationScript: initScript, migrationScripts: migrations);
    return await openDatabaseWithMigration(path, migrationConfig);
  }
}
