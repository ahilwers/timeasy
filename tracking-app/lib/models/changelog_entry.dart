import 'package:timeasy/models/change_type.dart';

class ChangelogEntry {
  final String entityType;
  final String entityId;
  final ChangeType changeType;
  final DateTime timestamp;
  final int? changelogId; // This will be set by the server or local DB
  final bool isFromServer; // True if this change came from server sync

  ChangelogEntry({
    required this.entityType,
    required this.entityId,
    required this.changeType,
    required this.timestamp,
    this.changelogId,
    this.isFromServer = false, // Default to local change
  });

  factory ChangelogEntry.fromMap(Map<String, dynamic> map) {
    return ChangelogEntry(
      changelogId: map['id'] as int?,
      entityType: map['entityType'],
      entityId: map['entityId'],
      changeType: ChangeTypeHelper.convertFromString(map['changeType']),
      timestamp: DateTime.fromMillisecondsSinceEpoch(map['timestamp']),
      isFromServer: map['isFromServer'] == 1, // SQLite stores as 0/1
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'entityType': entityType,
      'entityId': entityId,
      'changeType': ChangeTypeHelper.convertToString(changeType),
      'timestamp': timestamp.millisecondsSinceEpoch,
      'isFromServer': isFromServer ? 1 : 0, // Convert to 0/1 for SQLite
    };
  }
}
