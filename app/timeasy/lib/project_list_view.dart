import 'package:flutter/material.dart';

import 'package:timeasy/project_repository.dart';
import 'package:timeasy/project.dart';

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

  @override
  void initState() {
    super.initState();
    var projectRepository = new ProjectRepository();
    projectRepository.getAllProjects().then((List<Project> projectsFromDb) {
      setState(() {
        projects = projectsFromDb;
      });
    });
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
            // Add your onPressed code here!
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

  String _getTitle() {
    return "Projekte";
  }


}