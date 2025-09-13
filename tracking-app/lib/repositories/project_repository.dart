import 'package:shared_preferences/shared_preferences.dart';
import 'package:timeasy/dataaccess/database.dart';
import 'package:timeasy/models/change_type.dart';
import 'package:timeasy/models/changelog_entry.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/repositories/changelog_repository.dart';

class ProjectRepository {
  final ChangelogRepository _changelogRepository = ChangelogRepository();

  Future<Project> addProject(Project project) async {
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.insert(Project.tableName, project.toMap());
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'Project',
            entityId: project.id,
            changeType: ChangeType.NEW,
            timestamp: DateTime.now().toUtc(),
          ),
          txn);
    });
    return project;
  }

  // Method for syncing from server - creates changelog entries marked as server-side
  Future<Project> addProjectFromSync(Project project) async {
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.insert(Project.tableName, project.toMap());
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'Project',
            entityId: project.id,
            changeType: ChangeType.NEW,
            timestamp: DateTime.now().toUtc(),
            isFromServer: true, // Mark as server-originated
          ),
          txn);
    });
    return project;
  }

  Future<Project> updateProject(Project project) async {
    project.updated = DateTime.now().toUtc();
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.update(Project.tableName, project.toMap(),
          where: "${Project.idColumn} = ?", whereArgs: [project.id]);
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'Project',
            entityId: project.id,
            changeType: ChangeType.CHANGED,
            timestamp: DateTime.now().toUtc(),
          ),
          txn);
    });
    return project;
  }

  // Method for syncing from server - creates changelog entries marked as server-side
  Future<Project> updateProjectFromSync(Project project) async {
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.update(Project.tableName, project.toMap(),
          where: "${Project.idColumn} = ?", whereArgs: [project.id]);
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'Project',
            entityId: project.id,
            changeType: ChangeType.CHANGED,
            timestamp: DateTime.now().toUtc(),
            isFromServer: true, // Mark as server-originated
          ),
          txn);
    });
    return project;
  }

  Future<Project> deleteProject(Project project) async {
    project.deleted = true;
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.update(Project.tableName, project.toMap(),
          where: "${Project.idColumn} = ?", whereArgs: [project.id]);
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'Project',
            entityId: project.id,
            changeType: ChangeType.DELETED,
            timestamp: DateTime.now().toUtc(),
          ),
          txn);
    });
    return project;
  }

  // Method for syncing from server - creates changelog entries marked as server-side
  Future<Project> deleteProjectFromSync(Project project) async {
    project.deleted = true;
    final db = await DBProvider.dbProvider.database;
    await db.transaction((txn) async {
      await txn.update(Project.tableName, project.toMap(),
          where: "${Project.idColumn} = ?", whereArgs: [project.id]);
      await _changelogRepository.insert(
          ChangelogEntry(
            entityType: 'Project',
            entityId: project.id,
            changeType: ChangeType.DELETED,
            timestamp: DateTime.now().toUtc(),
            isFromServer: true, // Mark as server-originated
          ),
          txn);
    });
    return project;
  }

  Future<Project?> getProjectById(String id) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(Project.tableName,
        where: "${Project.idColumn} = ?", whereArgs: [id]);
    return queryResult.isNotEmpty
        ? queryResult.map((entry) => Project.fromMap(entry)).toList().first
        : null;
  }

  Future<List<Project>> getAllProjects() async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(Project.tableName,
        where: "${Project.deletedColumn} = 0",
        orderBy: "${Project.nameColumn}");
    return queryResult.isNotEmpty
        ? queryResult.map((entry) => Project.fromMap(entry)).toList()
        : [];
  }

  /// Check if any projects exist in the database
  Future<bool> hasProjects() async {
    final projects = await getAllProjects();
    return projects.isNotEmpty;
  }

  /// Creates a default poject if no project exists. Otherwise it returns the first
  /// one.
  Future<Project> createDefaultProjectIfNotExists(
      String defaultProjectName) async {
    var projects = await getAllProjects();
    if (projects.isEmpty) {
      var newProject = new Project();
      newProject.name = defaultProjectName;
      await addProject(newProject);
      return newProject;
    } else {
      return projects[0];
    }
  }

  Future<Project?> getLastUsedProjectOrDefault() async {
    SharedPreferences prefs = await SharedPreferences.getInstance();
    String? lastUsedProjectId = prefs.getString("currentProjectId");
    Project? lastUsedProject;

    if (lastUsedProjectId != null) {
      lastUsedProject = await getProjectById(lastUsedProjectId);
    }

    // If no last used project or it was deleted, just return null
    // instead of creating a default project
    if ((lastUsedProjectId == null) ||
        (lastUsedProject != null) && (lastUsedProject.deleted)) {
      // Check if any projects exist
      List<Project> projects = await getAllProjects();
      if (projects.isNotEmpty) {
        lastUsedProject = projects[0];
      } else {
        return null; // Return null if no projects exist
      }
    }

    return lastUsedProject;
  }

  void saveLastUsedProject(Project project) async {
    SharedPreferences prefs = await SharedPreferences.getInstance();
    prefs.setString("currentProjectId", project.id);
  }

  Future<List<Project>> getProjectsChangedAfter(
      DateTime? changeTimestamp) async {
    var timeStamp = 0;
    if (changeTimestamp != null)
      timeStamp = changeTimestamp.toUtc().millisecondsSinceEpoch;
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(Project.tableName,
        where: "${Project.updatedColumn} > ?",
        whereArgs: [timeStamp],
        orderBy: "${Project.nameColumn}");
    return queryResult.isNotEmpty
        ? queryResult.map((entry) => Project.fromMap(entry)).toList()
        : [];
  }
}
