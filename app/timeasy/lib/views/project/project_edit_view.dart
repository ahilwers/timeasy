import 'package:flutter/material.dart';

import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/models/project.dart';

enum ConfirmAction { CANCEL, ACCEPT }

class ProjectEditView extends StatelessWidget {
  final String? _projectId;

  ProjectEditView([this._projectId]);

  @override
  Widget build(BuildContext context) {
    return Scaffold(body: new ProjectEditWidget(_projectId));
  }
}

class ProjectEditWidget extends StatefulWidget {
  final String? _projectId;

  ProjectEditWidget(this._projectId);

  @override
  _ProjectEditWidgetState createState() {
    return new _ProjectEditWidgetState(_projectId);
  }
}

class _ProjectEditWidgetState extends State<ProjectEditWidget> {
  String? _projectId;
  Project? _project;
  final ProjectRepository _projectRepository = new ProjectRepository();
  final _formEditProjectKey = GlobalKey<FormState>();

  _ProjectEditWidgetState([String? projectId]) {
    _projectId = projectId;
  }

  @override
  void initState() {
    super.initState();
    if (_projectId != null) {
      _projectRepository.getProjectById(_projectId!).then(
        (Project? projectFromDb) {
          setState(
            () {
              _project = projectFromDb!;
            },
          );
        },
      );
    } else {
      _project = new Project();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_project == null) {
      return Scaffold(
        appBar: new AppBar(
          title: new Text(AppLocalizations.of(context)!.loadingProject),
        ),
      );
    } else {
      return Scaffold(
        appBar: AppBar(
          title: Text(_getTitle()),
          actions: <Widget>[
            TextButton(
              onPressed: () {
                final form = _formEditProjectKey.currentState;
                if (form!.validate()) {
                  _saveProject(form);
                  Navigator.pop(context);
                }
              },
              child: Text(
                AppLocalizations.of(context)!.save,
                style: Theme.of(context).textTheme.subtitle1!.copyWith(color: Colors.white),
              ),
            ),
            _projectId != null
                ? TextButton(
                    onPressed: () {
                      deleteProjectWithRequest(context);
                    },
                    child: Text(
                      AppLocalizations.of(context)!.delete,
                      style: Theme.of(context).textTheme.subtitle1!.copyWith(color: Colors.white),
                    ),
                  )
                : Container(),
          ],
        ),
        body: Container(
          margin: EdgeInsets.all(16.0),
          child: Form(
            key: _formEditProjectKey,
            child: TextFormField(
              decoration: InputDecoration(
                labelText: AppLocalizations.of(context)!.projectName,
                border: OutlineInputBorder(),
              ),
              keyboardType: TextInputType.text,
              initialValue: _project!.name,
              validator: (value) {
                return _validateProjectName(value!);
              },
              onSaved: (value) => _project!.name = value!,
            ),
          ),
        ),
      );
    }
  }

  String _getTitle() {
    if (_projectId == null) {
      return AppLocalizations.of(context)!.addProject;
    } else {
      return AppLocalizations.of(context)!.editProject;
    }
  }

  String? _validateProjectName(String value) {
    if (value.isEmpty) {
      return AppLocalizations.of(context)!.errorMissingProjectName;
    } else {
      return null;
    }
  }

  void _saveProject(FormState form) {
    form.save();
    if (_projectId != null) {
      _projectRepository.updateProject(_project!);
    } else {
      _projectRepository.addProject(_project!);
    }
  }

  Future<ConfirmAction?> deleteProjectWithRequest(BuildContext context) async {
    return showDialog<ConfirmAction>(
      context: context,
      barrierDismissible: false, // user must tap button for close dialog!
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(AppLocalizations.of(context)!.delete),
          content: Text(AppLocalizations.of(context)!.deleteProjectRequest),
          actions: <Widget>[
            TextButton(
              child: Text(AppLocalizations.of(context)!.no),
              onPressed: () {
                Navigator.of(context).pop(ConfirmAction.CANCEL);
              },
            ),
            TextButton(
              child: Text(AppLocalizations.of(context)!.yes),
              onPressed: () {
                deleteProject();
                Navigator.of(context).pop(ConfirmAction.ACCEPT);
                Navigator.pop(context);
              },
            )
          ],
        );
      },
    );
  }

  void deleteProject() {
    if (_projectId != null) {
      _projectRepository.deleteProject(_project!);
    }
  }
}
