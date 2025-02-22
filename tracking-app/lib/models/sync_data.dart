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

    return SyncData(
      timeEntries: (jsonData['TimeEntries'] as List<dynamic>)
          .map((entry) => TimeEntrySyncData.fromJson(entry))
          .toList(),
      projects: (jsonData['Projects'] as List<dynamic>)
          .map((project) => ProjectSyncData.fromJson(project))
          .toList(),
    );
  }
}
