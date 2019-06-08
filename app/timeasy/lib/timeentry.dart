import 'package:uuid/uuid.dart';

class TimeEntry {

  static final String tableName = "TimeEntries";
  static final String idColumn = "id";
  static final String startTimeColumn = "startTime";
  static final String endTimeColumn = "endTime";
  static final String descriptionColumn = "description";

  String id;
  DateTime startTime;
  DateTime endTime;
  String description;

  TimeEntry() {
    var uuid = new Uuid();
    id = uuid.v1();
    startTime = DateTime.now().toUtc();
  }

  TimeEntry.fromMap(Map<String, dynamic> map) {
    id = map[idColumn];
    int startTimeMillis = map[startTimeColumn];
    startTime = new DateTime.fromMillisecondsSinceEpoch(startTimeMillis, isUtc: true);
    int endTimeMillis = map[endTimeColumn];
    endTime = new DateTime.fromMillisecondsSinceEpoch(endTimeMillis, isUtc: true);
    description = map[descriptionColumn];
  }

  Map<String, dynamic> toMap() {
    var map = <String, dynamic>{
      idColumn : id,
      startTimeColumn : startTime.millisecondsSinceEpoch,
      endTimeColumn : 0,
      descriptionColumn : description
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