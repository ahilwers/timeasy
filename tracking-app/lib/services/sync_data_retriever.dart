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

class SyncDataRetriever {
  final SynchronizationApiService _apiService;
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final ProjectRepository _projectRepository = new ProjectRepository();
  final SettingsRepository _settingsRepository = new SettingsRepository();

  SyncDataRetriever(this._apiService) {}

  Future<void> retrieveNewestEntries(DateTime? changedAfter) async {
    var syncData = await _apiService.getChangedData(changedAfter);
    await _saveEntries(syncData);
  }

  Future<void> _saveEntries(SyncData syncData) async {
    await _saveTimeEntries(syncData.timeEntries);
    await _saveProjects(syncData.projects);
  }

  Future<void> _saveTimeEntries(List<TimeEntrySyncData> syncData) async {
    DateTime? latestUpdateTime = null;
    for (var entry in syncData) {
      if (latestUpdateTime == null ||
          entry.changeTimestamp.isAfter(latestUpdateTime)) {
        latestUpdateTime = entry.changeTimestamp;
      }
      await _saveTimeEntry(entry);
    }
    _updateRemoteTimeEntrySettings(latestUpdateTime);
  }

  Future<void> _saveTimeEntry(TimeEntrySyncData syncData) async {
    var timeEntry = _createTimeEntryFromSyncData(syncData);
    var existingTimeEntry =
        await _timeEntryRepository.getTimeEntryById(timeEntry.id);
    switch (syncData.changeType) {
      case ChangeType.NEW:
      case ChangeType.CHANGED:
        if (existingTimeEntry == null) {
          await _timeEntryRepository.addTimeEntry(timeEntry);
        } else if (syncData.changeTimestamp
            .isAfter(existingTimeEntry.updated)) {
          await _timeEntryRepository.updateTimeEntry(timeEntry);
        }
        break;
      case ChangeType.DELETED:
        if (existingTimeEntry != null &&
            syncData.changeTimestamp.isAfter(existingTimeEntry.updated)) {
          await _timeEntryRepository.deleteTimeEntry(timeEntry);
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
    timeEntry.updated = entry.changeTimestamp;
    timeEntry.created = DateTime.now();
    return timeEntry;
  }

  Future<void> _saveProjects(List<ProjectSyncData> syncData) async {
    DateTime? latestUpdateTime = null;
    for (var project in syncData) {
      if (latestUpdateTime == null ||
          project.changeTimestamp.isAfter(latestUpdateTime)) {
        latestUpdateTime = project.changeTimestamp;
      }
      await _saveProject(project);
    }
    await _updateRemoteProjectSettings(latestUpdateTime);
  }

  Future<void> _saveProject(ProjectSyncData syncData) async {
    var project = _createProjectFromSyncData(syncData);
    var existingProject = await _projectRepository.getProjectById(project.id);
    switch (syncData.changeType) {
      case ChangeType.NEW:
      case ChangeType.CHANGED:
        if (existingProject == null) {
          _projectRepository.addProject(project);
        } else if (syncData.changeTimestamp.isAfter(existingProject.updated)) {
          _projectRepository.updateProject(project);
        }
        break;
      case ChangeType.DELETED:
        if (existingProject != null &&
            syncData.changeTimestamp.isAfter(existingProject.updated)) {
          _projectRepository.deleteProject(project);
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

  Future<void> _updateRemoteTimeEntrySettings(
      DateTime? latestUpdateTime) async {
    var settings = await _settingsRepository.getSettings();
    settings.latestRemoteTimeEntryTimestamp = latestUpdateTime;
    await _settingsRepository.saveSettings(settings);
  }

  Future<void> _updateRemoteProjectSettings(DateTime? latestUpdateTime) async {
    var settings = await _settingsRepository.getSettings();
    settings.latestRemoteProjectTimestamp = latestUpdateTime;
    await _settingsRepository.saveSettings(settings);
  }
}
