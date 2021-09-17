import 'package:flutter/material.dart';

import 'package:timeasy/project_repository.dart';
import 'package:timeasy/project.dart';

class ProjectEditView extends StatelessWidget {

  String _projectId;

  ProjectEditView({String projectId}) {
    _projectId = projectId;
  }

  @override
  Widget build(BuildContext context) {

    return Scaffold(
        body: new ProjectEditWidget(_projectId)
    );
  }

}

class ProjectEditWidget extends StatefulWidget {

  String _projectId;

  ProjectEditWidget(String projectId) {
    _projectId = projectId;
  }

  @override
  _ProjectEditWidgetState createState() {
    return new _ProjectEditWidgetState(_projectId);
  }

}

class _ProjectEditWidgetState extends State<ProjectEditWidget> {

  String _projectId;
  Project _project;
  final ProjectRepository _projectRepository = new ProjectRepository();
  final _formEditProjectKey = GlobalKey<FormState>();

  _ProjectEditWidgetState(String projectId) {
    _projectId = projectId;
  }

  List<Project> projects;

  @override
  void initState() {
    super.initState();
    if (_projectId!=null) {
      _projectRepository.getProjectById(_projectId).then((Project projectFromDb) {
        setState(() {
          _project = projectFromDb;
        });
      });
    } else {
      _project = new Project();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_project == null) {
      return Scaffold(
        appBar: new AppBar(
          title: new Text("Lade Projekt..."),
        ),
      );
    } else {
      return Scaffold(
        appBar: AppBar(
          title: Text(_getTitle()),
          actions: <Widget>[
            FlatButton(
              onPressed: () {
                final form = _formEditProjectKey.currentState;
                if (form.validate()) {
                  _saveProject(form);
                  Navigator.pop(context);
                }
              },
              child: Text("Speichern",
                style: Theme.of(context)
                  .textTheme
                  .subtitle1
                  .copyWith(color: Colors.white),
              ),
            ),

          ],

        ),
        body: Container(
          margin: EdgeInsets.all(16.0),
          child: Form(
            key: _formEditProjectKey,
            child: TextFormField(
              decoration: InputDecoration(
                labelText: 'Projektname',
                border: OutlineInputBorder(),
              ),
              keyboardType: TextInputType.text,
              initialValue: _project.name,
              validator: (value) {
                return _validateProjectName(value);
              },
              onSaved: (value) => _project.name = value,
            ),

          ),
        )
      );
    }
  }

  String _getTitle() {
    if (_projectId==null) {
      return "Projekt hinzufügen";
    } else {
      return "Projekt bearbeiten";
    }
  }

  String _validateProjectName(String value) {
    if (value.isEmpty) {
      return "Bitte geben Sie Ihrem Projekt einen Namen.";
    } else {
      return null;
    }
  }

  void _saveProject(FormState form) {
    form.save();
    if (_projectId!=null) {
      _projectRepository.updateProject(_project);
    } else {
      _projectRepository.addProject(_project);
    }
  }



}