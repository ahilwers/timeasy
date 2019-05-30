import 'package:timeasy/Database.dart';
import 'package:timeasy/TimeEntry.dart';

class TimeEntryRepository {

  addTimeEntry(TimeEntry timeEntry) async {
    final db = await DBProvider.dbProvider.database;
    return await db.insert(TimeEntry.tableName, timeEntry.toMap());
    /*
    return await db.rawInsert("INSERT INTO ${timeEntry.tableName} "
      "(${timeEntry.idColumn}, ${timeEntry.startTimeColumn}, ${timeEntry.endTimeColumn}, ${timeEntry.descriptionColumn}) "
      "VALUES (\"${timeEntry.id}\", \"${timeEntry.startTime}\", \"${timeEntry.endTime}\", \"${timeEntry.description}\") ");
  */
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
}