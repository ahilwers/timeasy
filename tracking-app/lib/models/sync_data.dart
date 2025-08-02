import 'dart:convert';

import 'project_sync_data.dart';
import 'time_entry_sync_data.dart';

class SyncData {
  final List<TimeEntrySyncData> timeEntries;
  final List<ProjectSyncData> projects;
  final int? latestChangelogId;

  SyncData({
    required this.timeEntries,
    required this.projects,
    this.latestChangelogId,
  });

  factory SyncData.fromJson(String jsonString) {
    final Map<String, dynamic> jsonData = json.decode(jsonString);

    var timeEntries = jsonData['TimeEntries'] ?? [];
    var projects = jsonData['Projects'] ?? [];
    var latestChangelogId = jsonData['LatestChangeLogId'];

    return SyncData(
      timeEntries: (timeEntries as List<dynamic>)
          .map((entry) => TimeEntrySyncData.fromJson(entry))
          .toList(),
      projects: (projects as List<dynamic>)
          .map((project) => ProjectSyncData.fromJson(project))
          .toList(),
      latestChangelogId: latestChangelogId,
    );
  }

  Map<String, dynamic> toJson() {
    Map<String, dynamic> json = {
      'TimeEntries': timeEntries.map((entry) => entry.toJson()).toList(),
      'Projects': projects.map((project) => project.toJson()).toList(),
    };
    if (latestChangelogId != null) {
      json['LatestChangeLogId'] = latestChangelogId;
    }
    return json;
  }
}
