import 'package:uuid/uuid.dart';

class Settings {
  static final String tableName = "Settings";
  static final String idColumn = "id";
  static final String lastSyncTimeColumn = "lastSyncTime";
  static final String latestRemoteTimeEntryTimestampColumn = "latestRemoteTimeEntryTimestamp";
  static final String latestRemoteProjectTimestampColumn = "latestRemoteProjectTimestamp";
  static final String latestLocalTimeEntryTimestampColumn = "latestLocalTimeEntryTimestamp";
  static final String latestLocalProjectTimestampColumn = "latestLocalProjectTimestamp";

  late String id;
  DateTime? lastSyncTime;
  DateTime? latestRemoteTimeEntryTimestamp;
  DateTime? latestRemoteProjectTimestamp;
  DateTime? latestLocalTimeEntryTimestamp;
  DateTime? latestLocalProjectTimestamp;

  Settings() {
    var uuid = new Uuid();
    id = uuid.v4();
  }

  void clear() {
    lastSyncTime = null;
    latestRemoteTimeEntryTimestamp = null;
    latestRemoteProjectTimestamp = null;
  }

  Settings.fromMap(Map<String, dynamic> map) {
    id = map[idColumn];
    int syncTimeMillis = map[lastSyncTimeColumn];
    if (syncTimeMillis > 0) {
      lastSyncTime = new DateTime.fromMillisecondsSinceEpoch(syncTimeMillis, isUtc: true);
    }
    syncTimeMillis = map[latestRemoteTimeEntryTimestampColumn];
    if (syncTimeMillis > 0) {
      latestRemoteTimeEntryTimestamp = new DateTime.fromMillisecondsSinceEpoch(syncTimeMillis, isUtc: true);
    }
    syncTimeMillis = map[latestRemoteProjectTimestampColumn];
    if (syncTimeMillis > 0) {
      latestRemoteProjectTimestamp = new DateTime.fromMillisecondsSinceEpoch(syncTimeMillis, isUtc: true);
    }
    syncTimeMillis = map[latestLocalTimeEntryTimestampColumn];
    if (syncTimeMillis > 0) {
      latestLocalTimeEntryTimestamp = new DateTime.fromMillisecondsSinceEpoch(syncTimeMillis, isUtc: true);
    }
    syncTimeMillis = map[latestLocalProjectTimestampColumn];
    if (syncTimeMillis > 0) {
      latestLocalProjectTimestamp = new DateTime.fromMillisecondsSinceEpoch(syncTimeMillis, isUtc: true);
    }
  }

  Map<String, dynamic> toMap() {
    var map = <String, dynamic>{
      idColumn: id,
      lastSyncTimeColumn: lastSyncTime?.millisecondsSinceEpoch ?? 0,
      latestRemoteTimeEntryTimestampColumn: latestRemoteTimeEntryTimestamp?.millisecondsSinceEpoch ?? 0,
      latestRemoteProjectTimestampColumn: latestRemoteProjectTimestamp?.millisecondsSinceEpoch ?? 0,
      latestLocalTimeEntryTimestampColumn: latestLocalTimeEntryTimestamp?.millisecondsSinceEpoch ?? 0,
      latestLocalProjectTimestampColumn: latestLocalProjectTimestamp?.millisecondsSinceEpoch ?? 0
    };
    return map;
  }
}
