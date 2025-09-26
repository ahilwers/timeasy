import 'package:timeasy/models/change_type.dart';
import 'package:timeasy/models/changelog_entry.dart';
import 'package:timeasy/models/project_sync_data.dart';
import 'package:timeasy/models/sync_data.dart';
import 'package:timeasy/models/time_entry_sync_data.dart';
import 'package:timeasy/repositories/changelog_repository.dart';
import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/repositories/settings_repository.dart';
import 'package:timeasy/repositories/time_entry_repository.dart';
import 'package:timeasy/services/synchronization_api_service.dart';
import 'package:timeasy/utils/app_logger.dart';

class SyncDataSender {
  final SynchronizationApiService _apiService;
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final ProjectRepository _projectRepository = new ProjectRepository();
  final SettingsRepository _settingsRepository = new SettingsRepository();
  final ChangelogRepository _changelogRepository = new ChangelogRepository();

  SyncDataSender(this._apiService) {}

  Future<void> sendNewestEntries(String? clientId) async {
    final prepared = await _createPreparedSync();
    AppLogger.i(
        'About to send ${prepared.payload.timeEntries.length} time entries, ${prepared.payload.projects.length} projects. Max changelog ID: ${prepared.maxLocalChangelogIdIncluded}',
        method: 'sync');

    try {
      await _apiService.sendSyncData(prepared.payload, clientId);
      AppLogger.i('Successfully sent data to server', method: 'sync');

      await _saveLatestLocalChangelogId(prepared.maxLocalChangelogIdIncluded);
      AppLogger.i(
          'Updated latestLocalChangelogId to ${prepared.maxLocalChangelogIdIncluded} and deleted sent entries',
          method: 'sync');
    } catch (e) {
      AppLogger.e('FAILED to send data to server', error: e, method: 'sync');
      rethrow;
    }
  }

  // class moved to top-level below

  Future<_PreparedSyncPayload> _createPreparedSync() async {
    var settings = await _settingsRepository.getSettings();
    var changelogEntries = await _changelogRepository
        .getUnsentChanges(settings.latestLocalChangelogId);

    int? maxLocalId;
    if (changelogEntries.isNotEmpty) {
      maxLocalId = changelogEntries
          .map((e) => e.changelogId!)
          .reduce((a, b) => a > b ? a : b);
    }

    // Coalesce per-entity changes to a single final operation
    final Map<String, ChangelogEntry> finalChanges = {};
    for (final entry in changelogEntries) {
      final key = '${entry.entityType}:${entry.entityId}';
      final existing = finalChanges[key];
      if (existing == null) {
        finalChanges[key] = entry;
        continue;
      }
      switch (entry.changeType) {
        case ChangeType.DELETED:
          // Deletion wins; keep latest timestamp for auditing
          finalChanges[key] = ChangelogEntry(
            entityType: existing.entityType,
            entityId: existing.entityId,
            changeType: ChangeType.DELETED,
            timestamp: entry.timestamp.isAfter(existing.timestamp)
                ? entry.timestamp
                : existing.timestamp,
            changelogId: entry.changelogId,
          );
          break;
        case ChangeType.NEW:
          // Keep NEW (idempotent)
          finalChanges[key] = ChangelogEntry(
            entityType: existing.entityType,
            entityId: existing.entityId,
            changeType: ChangeType.NEW,
            timestamp: entry.timestamp.isAfter(existing.timestamp)
                ? entry.timestamp
                : existing.timestamp,
            changelogId: entry.changelogId,
          );
          break;
        case ChangeType.CHANGED:
          // If already NEW, keep NEW; else mark as CHANGED
          if (existing.changeType == ChangeType.NEW) {
            // Keep as NEW but bump timestamp
            finalChanges[key] = ChangelogEntry(
              entityType: existing.entityType,
              entityId: existing.entityId,
              changeType: ChangeType.NEW,
              timestamp: entry.timestamp.isAfter(existing.timestamp)
                  ? entry.timestamp
                  : existing.timestamp,
              changelogId: entry.changelogId,
            );
          } else if (existing.changeType == ChangeType.DELETED) {
            // Deletion still wins
            // no-op
          } else {
            finalChanges[key] = ChangelogEntry(
              entityType: existing.entityType,
              entityId: existing.entityId,
              changeType: ChangeType.CHANGED,
              timestamp: entry.timestamp.isAfter(existing.timestamp)
                  ? entry.timestamp
                  : existing.timestamp,
              changelogId: entry.changelogId,
            );
          }
          break;
      }
    }

    var projects = <ProjectSyncData>[];
    var timeEntries = <TimeEntrySyncData>[];

    for (final change in finalChanges.values) {
      if (change.entityType == 'Project') {
        final projectSyncData =
            await _createProjectSyncDataFromChangelog(change);
        if (projectSyncData != null) {
          projects.add(projectSyncData);
        }
      } else if (change.entityType == 'TimeEntry') {
        final timeEntrySyncData =
            await _createTimeEntrySyncDataFromChangelog(change);
        if (timeEntrySyncData != null) {
          timeEntries.add(timeEntrySyncData);
        }
      }
    }

    return _PreparedSyncPayload(
      payload: SyncData(timeEntries: timeEntries, projects: projects),
      maxLocalChangelogIdIncluded: maxLocalId,
    );
  }

  Future<ProjectSyncData?> _createProjectSyncDataFromChangelog(
      ChangelogEntry changelogEntry) async {
    var project =
        await _projectRepository.getProjectById(changelogEntry.entityId);
    if (project != null || changelogEntry.changeType == ChangeType.DELETED) {
      return ProjectSyncData(
        id: changelogEntry.entityId,
        name: project?.name ?? '',
        color: project?.color ?? '#1E90FF',
        isActive: project?.isActive ?? true,
        changeType: changelogEntry.changeType,
        changeTimestamp: changelogEntry.timestamp,
      );
    }
    return null;
  }

  Future<TimeEntrySyncData?> _createTimeEntrySyncDataFromChangelog(
      ChangelogEntry changelogEntry) async {
    var timeEntry =
        await _timeEntryRepository.getTimeEntryById(changelogEntry.entityId);
    if (timeEntry != null || changelogEntry.changeType == ChangeType.DELETED) {
      return TimeEntrySyncData(
        id: changelogEntry.entityId,
        projectId: timeEntry?.projectId ?? '',
        description: timeEntry?.description ?? '',
        startTime: timeEntry?.startTime ?? DateTime.now(),
        endTime: timeEntry?.endTime,
        changeType: changelogEntry.changeType,
      );
    }
    return null;
  }

  Future<void> _saveLatestLocalChangelogId(int? maxIncludedId) async {
    if (maxIncludedId == null) return;
    var settings = await _settingsRepository.getSettings();
    settings.latestLocalChangelogId = maxIncludedId;
    await _settingsRepository.saveSettings(settings);

    // Clean up sent (or coalesced) changelog entries up to the included max id
    await _changelogRepository.deleteSentChanges(maxIncludedId);
  }
}

// Holds the payload and the max local changelog id included to avoid race
// conditions when cleaning up the local outbox.
class _PreparedSyncPayload {
  final SyncData payload;
  final int? maxLocalChangelogIdIncluded;
  _PreparedSyncPayload(
      {required this.payload, required this.maxLocalChangelogIdIncluded});
}
