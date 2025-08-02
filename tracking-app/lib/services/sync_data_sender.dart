import 'package:timeasy/models/change_type.dart';
import 'package:timeasy/models/changelog_entry.dart';
import 'package:timeasy/models/project_sync_data.dart';
import 'package:timeasy/models/sync_data.dart';
import 'package:timeasy/models/time_entry_sync_data.dart';
import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/repositories/settings_repository.dart';
import 'package:timeasy/repositories/time_entry_repository.dart';
import 'package:timeasy/repositories/changelog_repository.dart';
import 'package:timeasy/services/synchronization_api_service.dart';

class SyncDataSender {
  final SynchronizationApiService _apiService;
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final ProjectRepository _projectRepository = new ProjectRepository();
  final SettingsRepository _settingsRepository = new SettingsRepository();
  final ChangelogRepository _changelogRepository = new ChangelogRepository();

  SyncDataSender(this._apiService) {}

  Future<void> sendNewestEntries(String? clientId) async {
    var syncData = await createSyncData();
    await _apiService.sendSyncData(syncData, clientId);
    await _saveLatestLocalChangelogId();
  }

  Future<SyncData> createSyncData() async {
    var settings = await _settingsRepository.getSettings();
    var changelogEntries = await _changelogRepository.getUnsentChanges(settings.latestLocalChangelogId);
    
    var projects = <ProjectSyncData>[];
    var timeEntries = <TimeEntrySyncData>[];
    
    for (var changelogEntry in changelogEntries) {
      if (changelogEntry.entityType == 'Project') {
        var projectSyncData = await _createProjectSyncDataFromChangelog(changelogEntry);
        if (projectSyncData != null) {
          projects.add(projectSyncData);
        }
      } else if (changelogEntry.entityType == 'TimeEntry') {
        var timeEntrySyncData = await _createTimeEntrySyncDataFromChangelog(changelogEntry);
        if (timeEntrySyncData != null) {
          timeEntries.add(timeEntrySyncData);
        }
      }
    }
    
    return new SyncData(timeEntries: timeEntries, projects: projects);
  }

  Future<ProjectSyncData?> _createProjectSyncDataFromChangelog(ChangelogEntry changelogEntry) async {
    var project = await _projectRepository.getProjectById(changelogEntry.entityId);
    if (project != null || changelogEntry.changeType == ChangeType.DELETED) {
      return ProjectSyncData(
        id: changelogEntry.entityId,
        name: project?.name ?? '',
        color: project?.color ?? '#1E90FF',
        changeType: changelogEntry.changeType,
        changeTimestamp: changelogEntry.timestamp,
      );
    }
    return null;
  }

  Future<TimeEntrySyncData?> _createTimeEntrySyncDataFromChangelog(ChangelogEntry changelogEntry) async {
    var timeEntry = await _timeEntryRepository.getTimeEntryById(changelogEntry.entityId);
    if (timeEntry != null || changelogEntry.changeType == ChangeType.DELETED) {
      return TimeEntrySyncData(
        id: changelogEntry.entityId,
        projectId: timeEntry?.projectId ?? '',
        description: timeEntry?.description ?? '',
        startTime: timeEntry?.startTime ?? DateTime.now(),
        endTime: timeEntry?.endTime,
        changeType: changelogEntry.changeType,
        changeTimestamp: changelogEntry.timestamp,
      );
    }
    return null;
  }

  Future<void> _saveLatestLocalChangelogId() async {
    var settings = await _settingsRepository.getSettings();
    var changelogEntries = await _changelogRepository.getUnsentChanges(settings.latestLocalChangelogId);
    
    if (changelogEntries.isNotEmpty) {
      var latestChangelogId = changelogEntries.map((e) => e.changelogId!).reduce((a, b) => a > b ? a : b);
      settings.latestLocalChangelogId = latestChangelogId;
      await _settingsRepository.saveSettings(settings);
      
      // Clean up sent changelog entries
      await _changelogRepository.deleteSentChanges(latestChangelogId);
    }
  }
}
