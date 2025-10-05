import 'dart:convert';

import 'change_type.dart';

class TimeEntrySyncData {
  final String id;
  final String description;
  final DateTime startTime;
  final DateTime? endTime;
  final String projectId;
  final ChangeType changeType;
  final DateTime? changeTimestamp; // Provided by server for ordering/metadata
  final int? changeLogId; // Server change_log id (optional)
  final bool? deleted; // Optional deleted flag

  TimeEntrySyncData({
    required this.id,
    required this.description,
    required this.startTime,
    this.endTime,
    required this.projectId,
    required this.changeType,
    this.changeTimestamp,
    this.changeLogId,
    this.deleted,
  });

  factory TimeEntrySyncData.fromJson(Map<String, dynamic> json) {
    final startTime = DateTime.parse(json['startTime'] as String).toUtc();
    final String? endTimeStr = json['endTime'] as String?;
    DateTime? endTime;
    if (endTimeStr == null || endTimeStr.trim().isEmpty) {
      endTime = null;
    } else {
      endTime = DateTime.parse(endTimeStr).toUtc();
    }

    return TimeEntrySyncData(
      id: json['id'] as String,
      description: json['description'] as String,
      startTime: startTime,
      endTime: endTime,
      projectId: json['projectId'] as String,
      changeType:
          ChangeTypeHelper.convertFromString(json['changeType'] as String),
      changeTimestamp: (json['changeTimestamp'] != null &&
              (json['changeTimestamp'] as String).trim().isNotEmpty)
          ? DateTime.parse(json['changeTimestamp'] as String).toUtc()
          : null,
      changeLogId: json['changeLogId'] as int?,
      deleted: json['deleted'] as bool?,
    );
  }

  Map<String, dynamic> toJson() {
    Map<String, dynamic> json = {
      'id': id,
      'description': description,
      'startTime': startTime.toUtc().toIso8601String(),
      'endTime': endTime != null ? endTime!.toUtc().toIso8601String() : '',
      'projectId': projectId,
      'changeType': ChangeTypeHelper.convertToString(changeType),
    };

    // Only include deleted field if it's not null
    if (deleted != null) {
      json['deleted'] = deleted;
    }

    return json;
  }

  static List<TimeEntrySyncData> listFromJson(String jsonString) {
    final List<dynamic> jsonData = json.decode(jsonString);
    return jsonData.map((item) => TimeEntrySyncData.fromJson(item)).toList();
  }
}
