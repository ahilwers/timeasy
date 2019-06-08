import 'package:timeasy/database.dart';
import 'package:timeasy/timeentry.dart';

class TimeEntryRepository {

  addTimeEntry(TimeEntry timeEntry) async {
    final db = await DBProvider.dbProvider.database;
    return await db.insert(TimeEntry.tableName, timeEntry.toMap());
  }

  updateTimeEntry(TimeEntry timeEntry) async {
    final db = await DBProvider.dbProvider.database;
    return await db.update(TimeEntry.tableName, timeEntry.toMap(), where: "${TimeEntry.idColumn} = ?", whereArgs: [timeEntry.id]);

  }

  closeLatestTimeEntry() async {
    var latestTimeEntry = await getLatestOpenTimeEntry();
    if (latestTimeEntry != null) {
      latestTimeEntry.endTime = DateTime.now().toUtc();
      await updateTimeEntry(latestTimeEntry);
    }
  }

  Future<TimeEntry> getLatestOpenTimeEntryOrCreateNew() async {
    var latestEntry = await getLatestOpenTimeEntry();
    if (latestEntry==null) {
      latestEntry = new TimeEntry();
      await addTimeEntry(latestEntry);
    }
    return latestEntry;
  }

  Future<TimeEntry> getLatestOpenTimeEntry() async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName, where: "${TimeEntry.endTimeColumn} = ?", whereArgs: [0]);
    return queryResult.isNotEmpty ? TimeEntry.fromMap(queryResult.first) : null;
  }

  Future<List<TimeEntry>> getAllTimeEntries() async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName);
    return queryResult.isNotEmpty ? queryResult.map((entry) => TimeEntry.fromMap(entry)).toList() : [];
  }

  Future<List<TimeEntry>> getTimeEntries(DateTime startDate, DateTime endDate) async {
    var startMillis = getDateWithoutTime(startDate).millisecondsSinceEpoch;
    var endMillis = getDateWithoutTime(endDate).add(new Duration(days: 1)).millisecondsSinceEpoch;

    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName, where: "${TimeEntry.startTimeColumn} >= ? AND ${TimeEntry.endTimeColumn} < ?", whereArgs: [startMillis, endMillis], orderBy: TimeEntry.startTimeColumn);
    return queryResult.isNotEmpty ? queryResult.map((entry) => TimeEntry.fromMap(entry)).toList() : [];

  }
  
  DateTime getDateWithoutTime(DateTime date) {
    return new DateTime(date.year, date.month, date.day);
  }
}