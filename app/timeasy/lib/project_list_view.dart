import 'package:flutter/material.dart';

import 'package:timeasy/project_repository.dart';
import 'package:timeasy/project.dart';
import 'package:timeasy/project_edit_view.dart';

class ProjectListView extends StatelessWidget {

  @override
  Widget build(BuildContext context) {

    return Scaffold(
        body: new ProjectListWidget()
    );
  }

}

class ProjectListWidget extends StatefulWidget {

  @override
  _ProjectListWidgetState createState() {
    return new _ProjectListWidgetState();
  }

}

class _ProjectListWidgetState extends State<ProjectListWidget> {

  List<Project> projects;

  final ProjectRepository _projectRepository = new ProjectRepository();

  @override
  void initState() {
    super.initState();
    _loadProjects();
  }


  @override
  Widget build(BuildContext context) {
    if (projects == null) {
      return Scaffold(
        appBar: new AppBar(
          title: new Text("Lade Projekte..."),
        ),
      );
    } else {
      return Scaffold(
        appBar: AppBar(
          title: Text(_getTitle()),
        ),
        body: _dataBody(context),
        floatingActionButton: FloatingActionButton(
          onPressed: () {
            _addProject();
          },
          child: Icon(Icons.add),
          backgroundColor: Colors.blue,
        ),
      );
    }
  }

  _dataBody(BuildContext context) {
    return ListView.builder(
      itemCount: projects.length,
      itemBuilder: (context, index) {
        return ListTile(
          title: Text(projects[index].name),
        );
      },
    );
  }

  _addProject() {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => ProjectEditView(),
        fullscreenDialog: true,
      ),
    ).then((value) {
      _loadProjects();
    });
  }

  _loadProjects() {
    _projectRepository.getAllProjects().then((List<Project> projectsFromDb) {
      setState(() {
        projects = projectsFromDb;
      });
    });

  }

  String _getTitle() {
    return "Projekte";
  }


}