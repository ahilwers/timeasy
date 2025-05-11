import 'package:timeasy/models/change_type.dart';
import 'package:timeasy/models/project_sync_data.dart';
import 'package:timeasy/models/sync_data.dart';
import 'package:timeasy/models/time_entry_sync_data.dart';
import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/repositories/settings_repository.dart';
import 'package:timeasy/repositories/time_entry_repository.dart';
import 'package:timeasy/services/synchronization_api_service.dart';

class SyncDataSender {
  final SynchronizationApiService _apiService;
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final ProjectRepository _projectRepository = new ProjectRepository();
  final SettingsRepository _settingsRepository = new SettingsRepository();

  DateTime? _latestProjectUpdateTime;
  DateTime? _latestTimeEntryUpdateTime;

  SyncDataSender(this._apiService) {}

  Future<void> sendNewestEntries() async {
    var syncData = await createSyncData();
    await _apiService.sendSyncData(syncData);
    await _saveUpdateTimestamps();
  }

  Future<SyncData> createSyncData() async {
    var settings = await _settingsRepository.getSettings();
    var projects =
        await getProjectsToSend(settings.latestLocalProjectTimestamp);
    var timeEntries =
        await getTimeEntriesToSend(settings.latestLocalTimeEntryTimestamp);
    return new SyncData(timeEntries: timeEntries, projects: projects);
  }

  Future<List<ProjectSyncData>> getProjectsToSend(
      DateTime? changedAfter) async {
    var projects =
        await _projectRepository.getProjectsChangedAfter(changedAfter);
    var projectSyncDataList = <ProjectSyncData>[];
    for (var project in projects) {
      var changeType = ChangeType.CHANGED;
      if (project.deleted) {
        changeType = ChangeType.DELETED;
      } else if (project.created == project.updated) {
        changeType = ChangeType.NEW;
      }
      var projectSyncData = new ProjectSyncData(
        id: project.id,
        name: project.name,
        color: project.color,
        changeType: changeType,
        changeTimestamp: project.updated,
      );
      projectSyncDataList.add(projectSyncData);
      if (_latestProjectUpdateTime == null ||
          project.updated.isAfter(_latestProjectUpdateTime!)) {
        _latestProjectUpdateTime = project.updated;
      }
    }
    return projectSyncDataList;
  }

  Future<List<TimeEntrySyncData>> getTimeEntriesToSend(
      DateTime? changedAfter) async {
    var timeEntries =
        await _timeEntryRepository.getTimeEntriesChangedAfter(changedAfter);
    var timeEntrySyncDataList = <TimeEntrySyncData>[];
    for (var timeEntry in timeEntries) {
      var changeType = ChangeType.CHANGED;
      if (timeEntry.deleted) {
        changeType = ChangeType.DELETED;
      } else if (timeEntry.created == timeEntry.updated) {
        changeType = ChangeType.NEW;
      }
      var timeEntrieSyncData = new TimeEntrySyncData(
        projectId: timeEntry.projectId,
        id: timeEntry.id,
        description: timeEntry.description ?? "",
        startTime: timeEntry.startTime,
        endTime: timeEntry.endTime,
        changeType: changeType,
        changeTimestamp: timeEntry.updated,
      );
      timeEntrySyncDataList.add(timeEntrieSyncData);
      if (_latestTimeEntryUpdateTime == null ||
          timeEntry.updated.isAfter(_latestTimeEntryUpdateTime!)) {
        _latestTimeEntryUpdateTime = timeEntry.updated;
      }
    }
    return timeEntrySyncDataList;
  }

  Future<void> _saveUpdateTimestamps() async {
    var settings = await _settingsRepository.getSettings();
    settings.latestLocalTimeEntryTimestamp = _latestTimeEntryUpdateTime;
    settings.latestLocalProjectTimestamp = _latestProjectUpdateTime;
    await _settingsRepository.saveSettings(settings);
  }
}
