import 'package:timeasy/models/change_type.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/models/project_sync_data.dart';
import 'package:timeasy/models/sync_data.dart';
import 'package:timeasy/models/time_entry.dart';
import 'package:timeasy/models/time_entry_sync_data.dart';
import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/repositories/settings_repository.dart';
import 'package:timeasy/repositories/time_entry_repository.dart';
import 'package:timeasy/services/synchronization_api_service.dart';
import 'package:timeasy/tools/retrieve_changes_result.dart';

class SyncDataRetriever {
  final SynchronizationApiService _apiService;
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final ProjectRepository _projectRepository = new ProjectRepository();
  final SettingsRepository _settingsRepository = new SettingsRepository();

  SyncDataRetriever(this._apiService) {}

  Future<RetrieveChangesResult> retrieveNewestEntries(
      int? sinceChangelogId, String? clientId) async {
    var syncData = await _apiService.getChangedData(sinceChangelogId, clientId);
    await _saveEntries(syncData);
    return new RetrieveChangesResult(
        syncData.projects.isNotEmpty, syncData.timeEntries.isNotEmpty);
  }

  Future<void> _saveEntries(SyncData syncData) async {
    await _saveTimeEntries(syncData.timeEntries);
    await _saveProjects(syncData.projects);
    if (syncData.latestChangelogId != null) {
      await _updateRemoteChangelogIdSettings(syncData.latestChangelogId!);
    }
  }

  Future<void> _saveTimeEntries(List<TimeEntrySyncData> syncData) async {
    for (var entry in syncData) {
      await _saveTimeEntry(entry);
    }
  }

  Future<void> _saveTimeEntry(TimeEntrySyncData syncData) async {
    var timeEntry = _createTimeEntryFromSyncData(syncData);
    var existingTimeEntry =
        await _timeEntryRepository.getTimeEntryById(timeEntry.id);
    switch (syncData.changeType) {
      case ChangeType.NEW:
      case ChangeType.CHANGED:
        if (existingTimeEntry == null) {
          await _timeEntryRepository.addTimeEntryFromSync(timeEntry);
        } else {
          await _timeEntryRepository.updateTimeEntryFromSync(timeEntry);
        }
        break;
      case ChangeType.DELETED:
        if (existingTimeEntry != null) {
          await _timeEntryRepository.deleteTimeEntryFromSync(timeEntry);
        }
        break;
    }
  }

  TimeEntry _createTimeEntryFromSyncData(TimeEntrySyncData entry) {
    var timeEntry = new TimeEntry(entry.projectId);
    timeEntry.id = entry.id;
    timeEntry.description = entry.description;
    timeEntry.startTime = entry.startTime;
    timeEntry.endTime = entry.endTime;
    timeEntry.created = DateTime.now();
    return timeEntry;
  }

  Future<void> _saveProjects(List<ProjectSyncData> syncData) async {
    for (var project in syncData) {
      await _saveProject(project);
    }
  }

  Future<void> _saveProject(ProjectSyncData syncData) async {
    var project = _createProjectFromSyncData(syncData);
    var existingProject = await _projectRepository.getProjectById(project.id);
    switch (syncData.changeType) {
      case ChangeType.NEW:
      case ChangeType.CHANGED:
        if (existingProject == null) {
          _projectRepository.addProjectFromSync(project);
        } else if (syncData.changeTimestamp.isAfter(existingProject.updated)) {
          _projectRepository.updateProjectFromSync(project);
        }
        break;
      case ChangeType.DELETED:
        if (existingProject != null &&
            syncData.changeTimestamp.isAfter(existingProject.updated)) {
          _projectRepository.deleteProjectFromSync(project);
        }
        break;
    }
  }

  Project _createProjectFromSyncData(ProjectSyncData syncData) {
    var project = new Project();
    project.id = syncData.id;
    project.name = syncData.name;
    project.color = syncData.color;
    project.updated = syncData.changeTimestamp;
    project.created = DateTime.now();
    return project;
  }

  Future<void> _updateRemoteChangelogIdSettings(int latestChangelogId) async {
    var settings = await _settingsRepository.getSettings();
    settings.latestRemoteChangelogId = latestChangelogId;
    await _settingsRepository.saveSettings(settings);
  }
}
