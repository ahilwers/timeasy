import 'package:sqflite/sqflite.dart';
import 'package:timeasy/dataaccess/database.dart';
import 'package:timeasy/models/changelog_entry.dart';

class ChangelogRepository {
  static final String tableName = "Changelog";
  static final String idColumn = "id";
  static final String entityTypeColumn = "entityType";
  static final String entityIdColumn = "entityId";
  static final String changeTypeColumn = "changeType";
  static final String timestampColumn = "timestamp";

  Future<Database> get _db async => await DBProvider.dbProvider.database;

  Future<ChangelogEntry> insert(ChangelogEntry entry, Transaction txn) async {
    final id = await txn.insert(tableName, entry.toMap());
    return ChangelogEntry(
      changelogId: id,
      entityType: entry.entityType,
      entityId: entry.entityId,
      changeType: entry.changeType,
      timestamp: entry.timestamp,
    );
  }

  Future<List<ChangelogEntry>> getUnsentChanges(int? lastChangelogId) async {
    final db = await _db;
    final List<Map<String, dynamic>> maps = await db.query(
      tableName,
      where: lastChangelogId != null ? '$idColumn > ?' : null,
      whereArgs: lastChangelogId != null ? [lastChangelogId] : null,
      orderBy: '$idColumn ASC',
    );

    return List.generate(maps.length, (i) {
      return ChangelogEntry.fromMap(maps[i]);
    });
  }

  Future<int> deleteSentChanges(int lastChangelogId) async {
    final db = await _db;
    return await db.delete(
      tableName,
      where: '$idColumn <= ?',
      whereArgs: [lastChangelogId],
    );
  }
}
