import 'dart:convert';

import 'change_type.dart';

class TimeEntrySyncData {
  final String id;
  final String description;
  final DateTime startTime;
  final DateTime? endTime;
  final String projectId;
  final ChangeType changeType;

  TimeEntrySyncData({
    required this.id,
    required this.description,
    required this.startTime,
    this.endTime,
    required this.projectId,
    required this.changeType,
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
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'description': description,
      'startTime': startTime.toUtc().toIso8601String(),
      'endTime': endTime != null ? endTime!.toUtc().toIso8601String() : '',
      'projectId': projectId,
      'changeType': ChangeTypeHelper.convertToString(changeType),
    };
  }

  static List<TimeEntrySyncData> listFromJson(String jsonString) {
    final List<dynamic> jsonData = json.decode(jsonString);
    return jsonData.map((item) => TimeEntrySyncData.fromJson(item)).toList();
  }
}
