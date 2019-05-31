import 'dart:io';

import 'package:path/path.dart';
import 'package:sqflite/sqflite.dart';
import 'package:path_provider/path_provider.dart';

class DBProvider {
  DBProvider._();
  static final DBProvider dbProvider = DBProvider._();

  static Database _database;

  Future<Database> get database async {
    if (_database==null) {
      _database = await _initDB();
    }
    return _database;
  }

  _initDB() async {
    Directory documentsDirectory = await getApplicationDocumentsDirectory();
    String path = join(documentsDirectory.path, join("timeasy", "timeasy.db"));
    return await openDatabase(path, version: 1,
        onOpen: (dbProvider)  async {},
        onCreate: (Database db, int version) async {
          await _createTables(db);
        }
    );
  }

  _createTables(Database db) async {
    _createTimeEntryTable(db);
  }

  _createTimeEntryTable(Database db) async {
    db.execute("CREATE TABLE TimeEntries ("
      "id TEXT, "
      "startTime INTEGER, "
      "endTime INTEGER, "
      "description TEXT "
      ")");
  }

}