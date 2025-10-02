import 'package:sqflite/sqflite.dart';
import 'package:timeasy/dataaccess/database.dart';
import 'package:timeasy/models/changelog_entry.dart';
import 'package:timeasy/utils/app_logger.dart';

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
      isFromServer: entry.isFromServer,
    );
  }

  Future<List<ChangelogEntry>> getUnsentChanges(int? lastChangelogId) async {
    final db = await _db;

    // Filter out server-originated changes - only send truly local changes
    String whereClause = 'isFromServer = 0';
    List<dynamic> whereArgs = [];

    if (lastChangelogId != null) {
      whereClause += ' AND $idColumn > ?';
      whereArgs.add(lastChangelogId);
    }

    AppLogger.d('ChangelogRepository: Getting unsent changes with lastChangelogId: $lastChangelogId, whereClause: $whereClause');

    final List<Map<String, dynamic>> maps = await db.query(
      tableName,
      where: whereClause,
      whereArgs: whereArgs.isNotEmpty ? whereArgs : null,
      orderBy: '$idColumn ASC',
    );

    final entries = List.generate(maps.length, (i) {
      return ChangelogEntry.fromMap(maps[i]);
    });

    AppLogger.d('ChangelogRepository: Found ${entries.length} unsent changes: ${entries.map((e) => 'ID:${e.changelogId} ${e.entityType}:${e.entityId} ${e.changeType}').join(', ')}');
    return entries;
  }

  Future<int> deleteSentChanges(int lastChangelogId) async {
    final db = await _db;
    final deletedCount = await db.delete(
      tableName,
      where: '$idColumn <= ?',
      whereArgs: [lastChangelogId],
    );
    AppLogger.d('ChangelogRepository: Deleted $deletedCount changelog entries with ID <= $lastChangelogId');
    return deletedCount;
  }
}
