import 'package:timeasy/database.dart';
import 'package:timeasy/project.dart';

class ProjectRepository {

  addProject(Project project) async {
    final db = await DBProvider.dbProvider.database;
    return await db.insert(Project.tableName, project.toMap());
  }

  updateProject(Project project) async {
    project.updated = DateTime.now().toUtc();
    final db = await DBProvider.dbProvider.database;
    return await db.update(Project.tableName, project.toMap(),
        where: "${Project.idColumn} = ?", whereArgs: [project.id]);
  }

  Future<Project> getProjectById(String id) async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(Project.tableName, where: "${Project.idColumn} = ?", whereArgs: [id]);
    return queryResult.isNotEmpty ? queryResult.map((entry) => Project.fromMap(entry)).toList().first : null;
  }

  Future<List<Project>> getAllProjects() async {
    final db = await DBProvider.dbProvider.database;
    var queryResult = await db.query(Project.tableName, orderBy: "${Project.nameColumn}");
    return queryResult.isNotEmpty ? queryResult.map((entry) => Project.fromMap(entry)).toList() : [];
  }

   /// Creates a default poject if no project exists. Otherwise it returns the first
   /// one.
  Future<Project> createDefaultProjectIfNotExists() async {
    var projects = await getAllProjects();
    if (projects.isEmpty) {
      var newProject = new Project();
      newProject.name = "Default";
      addProject(newProject);
      return newProject;
    } else {
      return projects[0];
    }
  }

}
