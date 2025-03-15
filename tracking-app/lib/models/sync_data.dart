import 'dart:convert';

import 'project_sync_data.dart';
import 'time_entry_sync_data.dart';

class SyncData {
  final List<TimeEntrySyncData> timeEntries;
  final List<ProjectSyncData> projects;

  SyncData({
    required this.timeEntries,
    required this.projects,
  });

  factory SyncData.fromJson(String jsonString) {
    final Map<String, dynamic> jsonData = json.decode(jsonString);

    var timeEntries = jsonData['TimeEntries'] ?? [];
    var projects = jsonData['Projects'] ?? [];

    return SyncData(
      timeEntries: (timeEntries as List<dynamic>)
          .map((entry) => TimeEntrySyncData.fromJson(entry))
          .toList(),
      projects: (projects as List<dynamic>)
          .map((project) => ProjectSyncData.fromJson(project))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'TimeEntries': timeEntries.map((entry) => entry.toJson()).toList(),
      'Projects': projects.map((project) => project.toJson()).toList(),
    };
  }
}
