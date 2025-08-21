import 'dart:convert';

import 'change_type.dart';

class ProjectSyncData {
  final String id;
  final String name;
  final String color;
  final bool isActive;
  final ChangeType changeType;
  final DateTime changeTimestamp;

  ProjectSyncData({
    required this.id,
    required this.name,
    this.color = '#1E90FF',
    this.isActive = true,
    required this.changeType,
    required this.changeTimestamp,
  });

  factory ProjectSyncData.fromJson(Map<String, dynamic> json) {
    return ProjectSyncData(
      id: json['id'] as String,
      name: json['name'] as String,
      color: json['color'] as String? ?? '#1E90FF',
      isActive: json['isActive'] as bool? ?? true,
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
      'color': color,
      'isActive': isActive,
      'changeType': ChangeTypeHelper.convertToString(changeType),
      'changeTimestamp': changeTimestamp.toUtc().toIso8601String(),
    };
  }

  static List<ProjectSyncData> listFromJson(String jsonString) {
    final List<dynamic> jsonData = json.decode(jsonString);
    return jsonData.map((item) => ProjectSyncData.fromJson(item)).toList();
  }
}
