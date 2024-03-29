import 'package:shared_preferences/shared_preferences.dart';

import 'package:timeasy/dataaccess/database.dart';
import 'package:timeasy/models/project.dart';

class ProjectRepository {
  addProject(Project project) async {
    final db = await DBProvider.dbProvider.database;
    return await db.insert(Project.tableName, project.toMap());
  }

  updateProject(Project project) async {
    project.updated = DateTime.now().toUtc();
    final db = await DBProvider.dbProvider.database;
    return await db.update(Project.tableName, project.toMap(), where: "${Project.idColumn} = ?", whereArgs: [project.id]);
  }

  deleteProject(Project project) async {
    project.deleted = true;
    return await updateProject(project);
  }

  Future<Project?> getProjectById(String id) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(Project.tableName, where: "${Project.idColumn} = ?", whereArgs: [id]);
    return queryResult.isNotEmpty ? queryResult.map((entry) => Project.fromMap(entry)).toList().first : null;
  }

  Future<List<Project>> getAllProjects() async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(Project.tableName, where: "${Project.deletedColumn} = 0", orderBy: "${Project.nameColumn}");
    return queryResult.isNotEmpty ? queryResult.map((entry) => Project.fromMap(entry)).toList() : [];
  }

  /// Creates a default poject if no project exists. Otherwise it returns the first
  /// one.
  Future<Project> createDefaultProjectIfNotExists(String defaultProjectName) async {
    var projects = await getAllProjects();
    if (projects.isEmpty) {
      var newProject = new Project();
      newProject.name = defaultProjectName;
      addProject(newProject);
      return newProject;
    } else {
      return projects[0];
    }
  }

  Future<Project> getLastUsedProjectOrDefault(String defaultProjectName) async {
    SharedPreferences prefs = await SharedPreferences.getInstance();
    String? lastUsedProjectId = prefs.getString("currentProjectId");
    Project? lastUsedProject;
    if (lastUsedProjectId != null) {
      lastUsedProject = await getProjectById(lastUsedProjectId);
    }
    if ((lastUsedProjectId == null) || (lastUsedProject != null) && (lastUsedProject.deleted)) {
      lastUsedProject = await createDefaultProjectIfNotExists(defaultProjectName);
    }
    return lastUsedProject!;
  }

  void saveLastUsedProject(Project project) async {
    SharedPreferences prefs = await SharedPreferences.getInstance();
    prefs.setString("currentProjectId", project.id);
  }
}
