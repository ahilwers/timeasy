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

  getLatestOpenTimeEntryOrCreateNew() async {
    var latestEntry = await getLatestOpenTimeEntry();
    if (latestEntry==null) {
      latestEntry = new TimeEntry();
      await addTimeEntry(latestEntry);
    }
    return latestEntry;
  }

  getLatestOpenTimeEntry() async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName, where: "${TimeEntry.endTimeColumn} = ?", whereArgs: [0]);
    return queryResult.isNotEmpty ? TimeEntry.fromMap(queryResult.first) : null;
  }

  Future<List<TimeEntry>> getAllTimeEntries() async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(TimeEntry.tableName);
    return queryResult.isNotEmpty ? queryResult.map((entry) => TimeEntry.fromMap(entry)).toList() : [];
  }
}