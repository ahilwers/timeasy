import 'package:sqflite/sqflite.dart';
import 'package:timeasy/dataaccess/database.dart';
import 'package:timeasy/models/change_type.dart';
import 'package:timeasy/models/changelog_entry.dart';
import 'package:timeasy/models/time_entry.dart';
import 'package:timeasy/repositories/changelog_repository.dart';

class TimeEntryRepository {
  final ChangelogRepository _changelogRepository = ChangelogRepository();

  Future<TimeEntry> addTimeEntry(TimeEntry timeEntry) async {
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.insert(TimeEntry.tableName, timeEntry.toMap());
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'TimeEntry',
            entityId: timeEntry.id,
            changeType: ChangeType.NEW,
            timestamp: DateTime.now().toUtc(),
          ),
          txn);
    });
    return timeEntry;
  }

  // Method for syncing from server - creates changelog entries marked as server-side
  Future<TimeEntry> addTimeEntryFromSync(TimeEntry timeEntry) async {
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.insert(TimeEntry.tableName, timeEntry.toMap());
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'TimeEntry',
            entityId: timeEntry.id,
            changeType: ChangeType.NEW,
            timestamp: DateTime.now().toUtc(),
            isFromServer: true, // Mark as server-originated
          ),
          txn);
    });
    return timeEntry;
  }

  Future<TimeEntry> updateTimeEntry(TimeEntry timeEntry) async {
    timeEntry.updated = DateTime.now().toUtc();
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.update(TimeEntry.tableName, timeEntry.toMap(),
          where: "${TimeEntry.idColumn} = ?", whereArgs: [timeEntry.id]);
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'TimeEntry',
            entityId: timeEntry.id,
            changeType: ChangeType.CHANGED,
            timestamp: DateTime.now().toUtc(),
          ),
          txn);
    });
    return timeEntry;
  }

  // Method for syncing from server - creates changelog entries marked as server-side
  Future<TimeEntry> updateTimeEntryFromSync(TimeEntry timeEntry) async {
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.update(TimeEntry.tableName, timeEntry.toMap(),
          where: "${TimeEntry.idColumn} = ?", whereArgs: [timeEntry.id]);
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'TimeEntry',
            entityId: timeEntry.id,
            changeType: ChangeType.CHANGED,
            timestamp: DateTime.now().toUtc(),
            isFromServer: true, // Mark as server-originated
          ),
          txn);
    });
    return timeEntry;
  }

  Future<TimeEntry> deleteTimeEntry(TimeEntry timeEntry) async {
    timeEntry.deleted = true;
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.update(TimeEntry.tableName, timeEntry.toMap(),
          where: "${TimeEntry.idColumn} = ?", whereArgs: [timeEntry.id]);
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'TimeEntry',
            entityId: timeEntry.id,
            changeType: ChangeType.DELETED,
            timestamp: DateTime.now().toUtc(),
          ),
          txn);
    });
    return timeEntry;
  }

  // Method for syncing from server - creates changelog entries marked as server-side
  Future<TimeEntry> deleteTimeEntryFromSync(TimeEntry timeEntry) async {
    timeEntry.deleted = true;
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.update(TimeEntry.tableName, timeEntry.toMap(),
          where: "${TimeEntry.idColumn} = ?", whereArgs: [timeEntry.id]);
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'TimeEntry',
            entityId: timeEntry.id,
            changeType: ChangeType.DELETED,
            timestamp: DateTime.now().toUtc(),
            isFromServer: true, // Mark as server-originated
          ),
          txn);
    });
    return timeEntry;
  }

  // Delete a time entry by ID
  Future<void> deleteTimeEntryById(String id) async {
    final timeEntry = await getTimeEntryById(id);
    if (timeEntry != null) {
      await deleteTimeEntry(timeEntry);
    }
  }

  closeLatestTimeEntry(String projectId) async {
    var latestTimeEntry = await getLatestOpenTimeEntry(projectId);
    if (latestTimeEntry != null) {
      latestTimeEntry.endTime = DateTime.now().toUtc();
      await updateTimeEntry(latestTimeEntry);
    }
  }

  Future<TimeEntry> getLatestOpenTimeEntryOrCreateNew(String projectId) async {
    var latestEntry = await getLatestOpenTimeEntry(projectId);
    if (latestEntry == null) {
      latestEntry = new TimeEntry(projectId);
      await addTimeEntry(latestEntry);
    }
    return latestEntry;
  }

  Future<TimeEntry?> getLatestOpenTimeEntry(String projectId) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName,
        where:
            "${TimeEntry.endTimeColumn} = ? AND ${TimeEntry.projectIdColumn} = ? AND DELETED = 0",
        whereArgs: [0, projectId]);
    return queryResult.isNotEmpty ? TimeEntry.fromMap(queryResult.first) : null;
  }

  Future<List<TimeEntry>> getOpenTimeEntriesForProject(String projectId) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName,
        where:
            "${TimeEntry.endTimeColumn} = ? AND ${TimeEntry.projectIdColumn} = ? AND DELETED = 0",
        whereArgs: [0, projectId],
        orderBy: "${TimeEntry.startTimeColumn} ASC");
    return queryResult.isNotEmpty
        ? queryResult.map((entry) => TimeEntry.fromMap(entry)).toList()
        : [];
  }

  Future<TimeEntry?> getTimeEntryById(String id) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName,
        where: "${TimeEntry.idColumn} = ?", whereArgs: [id]);
    return queryResult.isNotEmpty ? TimeEntry.fromMap(queryResult.first) : null;
  }

  Future<List<TimeEntry>> getAllTimeEntries(String projectId) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName,
        where: "${TimeEntry.projectIdColumn} = ? AND DELETED=0",
        whereArgs: [projectId],
        orderBy:
            "${TimeEntry.startTimeColumn} desc, ${TimeEntry.endTimeColumn} desc");
    return queryResult.isNotEmpty
        ? queryResult.map((entry) => TimeEntry.fromMap(entry)).toList()
        : [];
  }

  Future<List<TimeEntry>> getTimeEntries(
      String projectId, DateTime startDate, DateTime endDate) async {
    var startMillis = getDateWithoutTime(startDate).millisecondsSinceEpoch;
    var endMillis = getDateWithoutTime(endDate)
        .add(new Duration(days: 1))
        .millisecondsSinceEpoch;

    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName,
        where:
            "${TimeEntry.projectIdColumn} = ? AND ${TimeEntry.startTimeColumn} >= ? AND ${TimeEntry.endTimeColumn} < ? AND DELETED=0",
        whereArgs: [projectId, startMillis, endMillis],
        orderBy:
            "${TimeEntry.startTimeColumn} desc, ${TimeEntry.endTimeColumn} desc");
    return queryResult.isNotEmpty
        ? queryResult.map((entry) => TimeEntry.fromMap(entry)).toList()
        : [];
  }

  Future<List<TimeEntry>> getTimeEntriesChangedAfter(
      DateTime? changeTimestamp) async {
    var timeStamp = 0;
    if (changeTimestamp != null)
      timeStamp = changeTimestamp.toUtc().millisecondsSinceEpoch;
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName,
        where: "${TimeEntry.updatedColumn} > ?", whereArgs: [timeStamp]);
    return queryResult.isNotEmpty
        ? queryResult.map((entry) => TimeEntry.fromMap(entry)).toList()
        : [];
  }

  DateTime getDateWithoutTime(DateTime date) {
    return new DateTime(date.year, date.month, date.day);
  }

  Future<TimeEntry> startTimingWithConflictCheck(
      String projectId, String? description) async {
    final db = await DBProvider.dbProvider.database;

    late TimeEntry resultEntry;
    await db.transaction((txn) async {
      // First, check for any unexpected open entries and log them
      var openEntries =
          await _getOpenTimeEntriesForProjectInTransaction(projectId, txn);
      if (openEntries.isNotEmpty) {
        print(
            'WARNING: Found ${openEntries.length} unexpected open entries for project $projectId before starting new timing. Closing them.');

        // Close all existing open entries
        for (var entry in openEntries) {
          entry.endTime = DateTime.now().toUtc();
          await txn.update(TimeEntry.tableName, entry.toMap(),
              where: "${TimeEntry.idColumn} = ?", whereArgs: [entry.id]);
          await _changelogRepository.insert(
              ChangelogEntry(
                entityType: 'TimeEntry',
                entityId: entry.id,
                changeType: ChangeType.CHANGED,
                timestamp: DateTime.now().toUtc(),
              ),
              txn);
        }
      }

      // Create a new time entry
      final timeEntry = TimeEntry(projectId);
      if (description != null && description.isNotEmpty) {
        timeEntry.description = description;
      }

      // Add the new time entry
      await txn.insert(TimeEntry.tableName, timeEntry.toMap());
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'TimeEntry',
            entityId: timeEntry.id,
            changeType: ChangeType.NEW,
            timestamp: DateTime.now().toUtc(),
          ),
          txn);

      resultEntry = timeEntry;
    });

    return resultEntry;
  }

  Future<void> stopTimingWithConflictCheck(String projectId) async {
    final db = await DBProvider.dbProvider.database;

    await db.transaction((txn) async {
      // Get all open entries for this project
      var openEntries =
          await _getOpenTimeEntriesForProjectInTransaction(projectId, txn);

      if (openEntries.isEmpty) {
        print(
            'WARNING: No open time entries found for project $projectId when trying to stop timing.');
        return;
      }

      if (openEntries.length > 1) {
        print(
            'WARNING: Found ${openEntries.length} open entries for project $projectId when stopping timing. Closing all.');
      }

      // Close all open entries
      final endTime = DateTime.now().toUtc();
      for (var entry in openEntries) {
        entry.endTime = endTime;
        await txn.update(TimeEntry.tableName, entry.toMap(),
            where: "${TimeEntry.idColumn} = ?", whereArgs: [entry.id]);
        await _changelogRepository.insert(
            ChangelogEntry(
              entityType: 'TimeEntry',
              entityId: entry.id,
              changeType: ChangeType.CHANGED,
              timestamp: DateTime.now().toUtc(),
            ),
            txn);
      }
    });
  }

  Future<List<TimeEntry>> _getOpenTimeEntriesForProjectInTransaction(
      String projectId, Transaction txn) async {
    var queryResult = await txn.query(TimeEntry.tableName,
        where:
            "${TimeEntry.endTimeColumn} = ? AND ${TimeEntry.projectIdColumn} = ? AND DELETED = 0",
        whereArgs: [0, projectId],
        orderBy: "${TimeEntry.startTimeColumn} ASC");
    return queryResult.isNotEmpty
        ? queryResult.map((entry) => TimeEntry.fromMap(entry)).toList()
        : [];
  }

  Future<List<String>> getLastUniqueDescriptions(String projectId,
      {int limit = 100}) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.rawQuery('''
      SELECT DISTINCT ${TimeEntry.descriptionColumn} 
      FROM ${TimeEntry.tableName} 
      WHERE ${TimeEntry.projectIdColumn} = ? 
      AND ${TimeEntry.descriptionColumn} IS NOT NULL 
      AND ${TimeEntry.descriptionColumn} != '' 
      AND DELETED = 0
      ORDER BY ${TimeEntry.startTimeColumn} DESC
      LIMIT ?
    ''', [projectId, limit]);

    return queryResult
        .map((entry) => entry[TimeEntry.descriptionColumn] as String)
        .where((description) => description.isNotEmpty)
        .toList();
  }
}
