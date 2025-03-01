import 'dart:convert';

import 'change_type.dart';

class ProjectSyncData {
  final String id;
  final String name;
  final ChangeType changeType;
  final DateTime changeTimestamp;

  ProjectSyncData({
    required this.id,
    required this.name,
    required this.changeType,
    required this.changeTimestamp,
  });

  factory ProjectSyncData.fromJson(Map<String, dynamic> json) {
    return ProjectSyncData(
      id: json['id'] as String,
      name: json['name'] as String,
      changeType:
          ChangeTypeHelper.convertFromString(json['changeType'] as String),
      changeTimestamp:
          DateTime.parse(json['changeTimestamp'] as String).toUtc(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'changeType': ChangeTypeHelper.convertToString(changeType),
      'changeTimestamp': changeTimestamp.toUtc().toIso8601String(),
    };
  }

  static List<ProjectSyncData> listFromJson(String jsonString) {
    final List<dynamic> jsonData = json.decode(jsonString);
    return jsonData.map((item) => ProjectSyncData.fromJson(item)).toList();
  }
}
