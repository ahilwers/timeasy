import 'package:flutter/material.dart';

import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/views/project/project_edit_view.dart';

class ProjectListView extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(body: new ProjectListWidget());
  }
}

class ProjectListWidget extends StatefulWidget {
  @override
  _ProjectListWidgetState createState() {
    return new _ProjectListWidgetState();
  }
}

class _ProjectListWidgetState extends State<ProjectListWidget> {
  List<Project>? projects;

  final ProjectRepository _projectRepository = new ProjectRepository();

  @override
  void initState() {
    super.initState();
    _loadProjects();
  }

  // Convert hex color string to Color object
  Color _hexToColor(String hexString) {
    hexString = hexString.replaceAll('#', '');
    if (hexString.length == 6) {
      hexString = 'FF' + hexString;
    }
    return Color(int.parse(hexString, radix: 16));
  }

  @override
  Widget build(BuildContext context) {
    if (projects == null) {
      return Scaffold(
        appBar: new AppBar(
          title: new Text(AppLocalizations.of(context)!.loadingProjects),
        ),
      );
    } else {
      return Scaffold(
        appBar: AppBar(
          title: Text(_getTitle()),
          backgroundColor: Theme.of(context).primaryColor,
        ),
        body: _dataBody(context),
        floatingActionButton: FloatingActionButton(
          onPressed: () {
            _addOrEditProject();
          },
          child: Icon(Icons.add),
          backgroundColor: Theme.of(context).primaryColor,
        ),
      );
    }
  }

  _dataBody(BuildContext context) {
    return ListView.builder(
      itemCount: projects!.length,
      itemBuilder: (context, index) {
        return ListTile(
          leading: Container(
            width: 16,
            height: 16,
            decoration: BoxDecoration(
              color: _hexToColor(projects![index].color),
              shape: BoxShape.circle,
            ),
          ),
          title: Text(projects![index].name),
          onTap: () {
            _addOrEditProject(projectIdToEdit: projects![index].id);
          },
        );
      },
    );
  }

  void _addOrEditProject({String? projectIdToEdit}) {
    Navigator.of(context)
        .push(
      MaterialPageRoute(
        builder: (context) => ProjectEditView(projectIdToEdit),
        fullscreenDialog: true,
      ),
    )
        .then((value) {
      _loadProjects();
    });
  }

  void _loadProjects() {
    _projectRepository.getAllProjects().then((List<Project> projectsFromDb) {
      setState(() {
        projects = projectsFromDb;
      });
    });
  }

  String _getTitle() {
    return AppLocalizations.of(context)!.projects;
  }
}
