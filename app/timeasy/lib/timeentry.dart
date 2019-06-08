import 'package:uuid/uuid.dart';

class TimeEntry {

  static final String tableName = "TimeEntries";
  static final String idColumn = "id";
  static final String startTimeColumn = "startTime";
  static final String endTimeColumn = "endTime";
  static final String descriptionColumn = "description";
  static final String projectIdColumn = "projectId";
  static final String createdColumn = "created";
  static final String updatedColumn = "updated";

  String id;
  DateTime startTime;
  DateTime endTime;
  String description;
  String projectId;
  DateTime created = DateTime.now().toUtc();
  DateTime updated = DateTime.now().toUtc();

  TimeEntry(String forProjectId) {
    var uuid = new Uuid();
    id = uuid.v1();
    projectId = forProjectId;
    startTime = DateTime.now().toUtc();
  }

  TimeEntry.fromMap(Map<String, dynamic> map) {
    id = map[idColumn];
    int startTimeMillis = map[startTimeColumn];
    startTime = new DateTime.fromMillisecondsSinceEpoch(startTimeMillis, isUtc: true);
    int endTimeMillis = map[endTimeColumn];
    if (endTimeMillis>0) {
      endTime = new DateTime.fromMillisecondsSinceEpoch(endTimeMillis, isUtc: true);
    }
    description = map[descriptionColumn];
    projectId = map[projectIdColumn];
    int createdMillis = map[createdColumn];
    created = new DateTime.fromMillisecondsSinceEpoch(createdMillis, isUtc: true);
    int updatedMillis = map[updatedColumn];
    updated = new DateTime.fromMillisecondsSinceEpoch(updatedMillis, isUtc: true);
  }

  Map<String, dynamic> toMap() {
    var map = <String, dynamic>{
      idColumn : id,
      startTimeColumn : startTime.millisecondsSinceEpoch,
      endTimeColumn : 0,
      descriptionColumn : description,
      projectIdColumn: projectId,
      createdColumn : created.millisecondsSinceEpoch,
      updatedColumn : updated.millisecondsSinceEpoch,
    };
    if (endTime!=null) {
      map[endTimeColumn] = endTime.millisecondsSinceEpoch;
    }
    return map;
  }

  int getSeconds() {
    var myEndTime = endTime;
    if (myEndTime==null) {
      myEndTime = DateTime.now();
    }
    return myEndTime.difference(startTime).inSeconds;
  }


}