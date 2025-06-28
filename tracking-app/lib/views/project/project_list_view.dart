import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:timeasy/components/project_header_component.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/services/event_sync_service.dart';
import 'package:timeasy/views/project/project_edit_view.dart';

class ProjectListView extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: ProjectHeader(
        title: AppLocalizations.of(context)!.projects,
        showSettingsButton: false,
        actions: [
          IconButton(
            icon: Icon(
              Icons.add,
              color: Theme.of(context).primaryColor,
            ),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (context) => ProjectEditView(null),
                  fullscreenDialog: true,
                ),
              ).then((_) {
                // Refresh the project list when returning from edit view
                if (context
                        .findAncestorStateOfType<_ProjectListWidgetState>() !=
                    null) {
                  context
                      .findAncestorStateOfType<_ProjectListWidgetState>()!
                      ._loadProjects();
                }
              });
            },
          ),
        ],
      ),
      body: ProjectListWidget(),
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
      return Center(
        child: CircularProgressIndicator(),
      );
    } else {
      return _dataBody(context);
    }
  }

  _dataBody(BuildContext context) {
    return ListView.builder(
      itemCount: projects!.length,
      itemBuilder: (context, index) {
        final project = projects![index];

        return Dismissible(
          key: Key(project.id ?? index.toString()),
          background: Container(
            color: Colors.blue,
            alignment: Alignment.centerLeft,
            padding: EdgeInsets.symmetric(horizontal: 20),
            child: Icon(
              Icons.edit,
              color: Colors.white,
            ),
          ),
          secondaryBackground: Container(
            color: Colors.red,
            alignment: Alignment.centerRight,
            padding: EdgeInsets.symmetric(horizontal: 20),
            child: Icon(
              Icons.delete,
              color: Colors.white,
            ),
          ),
          confirmDismiss: (direction) async {
            if (direction == DismissDirection.endToStart) {
              // Delete action
              final bool? result = await showDialog<bool>(
                context: context,
                builder: (BuildContext context) {
                  return AlertDialog(
                    title: Text(AppLocalizations.of(context)!.delete),
                    content: Text(AppLocalizations.of(context)!
                        .deleteProjectConfirmation),
                    actions: <Widget>[
                      TextButton(
                        onPressed: () => Navigator.of(context).pop(false),
                        child: Text(AppLocalizations.of(context)!.cancel),
                      ),
                      TextButton(
                        onPressed: () => Navigator.of(context).pop(true),
                        child: Text(AppLocalizations.of(context)!.delete),
                      ),
                    ],
                  );
                },
              );
              return result ?? false;
            } else {
              // Edit action - don't actually dismiss, just navigate
              _addOrEditProject(projectIdToEdit: project.id);
              return false;
            }
          },
          onDismissed: (direction) {
            if (direction == DismissDirection.endToStart) {
              // Store a copy of the project for potential undo
              final deletedProject = project;
              final deletedIndex = index;

              // Delete the item
              _projectRepository.deleteProject(deletedProject).then((_) {
                setState(() {
                  projects!.removeAt(deletedIndex);
                });

                // Trigger synchronization after deleting a project
                EventSyncService().sendDataToServer();

                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(AppLocalizations.of(context)!.projectDeleted),
                    action: SnackBarAction(
                      label: AppLocalizations.of(context)!.undo,
                      onPressed: () {
                        // Undo deletion
                        _projectRepository.addProject(deletedProject).then((_) {
                          _loadProjects();
                        });
                      },
                    ),
                  ),
                );
              });
            }
          },
          child: Card(
            margin: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            child: InkWell(
              onTap: () {
                // Navigate to edit screen when tapping on the card
                _addOrEditProject(projectIdToEdit: project.id);
              },
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Row(
                  children: [
                    Container(
                      width: 16,
                      height: 16,
                      decoration: BoxDecoration(
                        color: _hexToColor(project.color),
                        shape: BoxShape.circle,
                      ),
                    ),
                    SizedBox(width: 16),
                    Expanded(
                      child: Text(
                        project.name,
                        style: TextStyle(
                          fontSize: 16,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
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
