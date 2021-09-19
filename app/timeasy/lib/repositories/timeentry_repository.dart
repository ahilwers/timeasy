import 'package:flutter/cupertino.dart';
import 'package:timeasy/dataaccess/database.dart';
import 'package:timeasy/models/timeentry.dart';

class TimeEntryRepository {
  addTimeEntry(TimeEntry timeEntry) async {
    final db = await DBProvider.dbProvider.database;
    return await db.insert(TimeEntry.tableName, timeEntry.toMap());
  }

  updateTimeEntry(TimeEntry timeEntry) async {
    timeEntry.updated = DateTime.now().toUtc();
    final db = await DBProvider.dbProvider.database;
    return await db.update(TimeEntry.tableName, timeEntry.toMap(), where: "${TimeEntry.idColumn} = ?", whereArgs: [timeEntry.id]);
  }

  deleteTimeEntry(TimeEntry timeEntry) async {
    timeEntry.updated = DateTime.now().toUtc();
    final db = await DBProvider.dbProvider.database;
    return await db.delete(TimeEntry.tableName, where: "${TimeEntry.idColumn} = ?", whereArgs: [timeEntry.id]);
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

  Future<TimeEntry> getLatestOpenTimeEntry(String projectId) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName, where: "${TimeEntry.endTimeColumn} = ? AND ${TimeEntry.projectIdColumn} = ?", whereArgs: [0, projectId]);
    return queryResult.isNotEmpty ? TimeEntry.fromMap(queryResult.first) : null;
  }

  Future<TimeEntry> getTimeEntryById(String id) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName, where: "${TimeEntry.idColumn} = ?", whereArgs: [id]);
    return queryResult.isNotEmpty ? TimeEntry.fromMap(queryResult.first) : null;
  }

  Future<List<TimeEntry>> getAllTimeEntries(String projectId) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName,
        where: "${TimeEntry.projectIdColumn} = ?", whereArgs: [projectId], orderBy: "${TimeEntry.startTimeColumn} desc, ${TimeEntry.endTimeColumn} desc");
    return queryResult.isNotEmpty ? queryResult.map((entry) => TimeEntry.fromMap(entry)).toList() : [];
  }

  Future<List<TimeEntry>> getTimeEntries(String projectId, DateTime startDate, DateTime endDate) async {
    var startMillis = getDateWithoutTime(startDate).millisecondsSinceEpoch;
    var endMillis = getDateWithoutTime(endDate).add(new Duration(days: 1)).millisecondsSinceEpoch;

    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName,
        where: "${TimeEntry.projectIdColumn} = ? AND ${TimeEntry.startTimeColumn} >= ? AND ${TimeEntry.endTimeColumn} < ?",
        whereArgs: [projectId, startMillis, endMillis],
        orderBy: TimeEntry.startTimeColumn);
    return queryResult.isNotEmpty ? queryResult.map((entry) => TimeEntry.fromMap(entry)).toList() : [];
  }

  DateTime getDateWithoutTime(DateTime date) {
    return new DateTime(date.year, date.month, date.day);
  }
}
