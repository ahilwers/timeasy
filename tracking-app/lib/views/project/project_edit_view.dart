import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/services/event_sync_service.dart';

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
  final EventSyncService _eventSyncService = new EventSyncService();
  final _formEditProjectKey = GlobalKey<FormState>();

  // Color options for the project
  final List<Map<String, String>> colors = [
    {'name': 'colorBlue', 'hex': '#1E90FF'},
    {'name': 'colorGreen', 'hex': '#2ECC71'},
    {'name': 'colorRed', 'hex': '#E74C3C'},
    {'name': 'colorOrange', 'hex': '#E67E22'},
    {'name': 'colorPurple', 'hex': '#9B59B6'},
    {'name': 'colorCyan', 'hex': '#1ABC9C'},
    {'name': 'colorYellow', 'hex': '#F1C40F'},
    {'name': 'colorPink', 'hex': '#E91E63'},
    {'name': 'colorGray', 'hex': '#95A5A6'},
    {'name': 'colorBrown', 'hex': '#A0522D'},
  ];

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

  // Convert hex color string to Color object
  Color _hexToColor(String hexString) {
    hexString = hexString.replaceAll('#', '');
    if (hexString.length == 6) {
      hexString = 'FF' + hexString;
    }
    return Color(int.parse(hexString, radix: 16));
  }

  // Get translated color name
  String _getColorName(String colorKey, BuildContext context) {
    switch (colorKey) {
      case 'colorBlue':
        return AppLocalizations.of(context)!.colorBlue;
      case 'colorGreen':
        return AppLocalizations.of(context)!.colorGreen;
      case 'colorRed':
        return AppLocalizations.of(context)!.colorRed;
      case 'colorOrange':
        return AppLocalizations.of(context)!.colorOrange;
      case 'colorPurple':
        return AppLocalizations.of(context)!.colorPurple;
      case 'colorCyan':
        return AppLocalizations.of(context)!.colorCyan;
      case 'colorYellow':
        return AppLocalizations.of(context)!.colorYellow;
      case 'colorPink':
        return AppLocalizations.of(context)!.colorPink;
      case 'colorGray':
        return AppLocalizations.of(context)!.colorGray;
      case 'colorBrown':
        return AppLocalizations.of(context)!.colorBrown;
      default:
        return AppLocalizations.of(context)!.colorBlue;
    }
  }

  // Get color key from hex value
  String _getColorKeyFromHex(String hexValue) {
    final color = colors.firstWhere(
      (color) => color['hex'] == hexValue,
      orElse: () => colors[0], // Default to blue if not found
    );
    return color['name']!;
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
          leading: IconButton(
            icon: Icon(Icons.arrow_back),
            onPressed: () => Navigator.pop(context),
          ),
          title: Text(_getTitle()),
          actions: <Widget>[
            TextButton(
              onPressed: () {
                final form = _formEditProjectKey.currentState;
                if (form!.validate()) {
                  _saveProject(form);
                  Navigator.pop(
                      context, _project); // Return the project to the caller
                }
              },
              child: Text(
                AppLocalizations.of(context)!.save,
                style: TextStyle(
                  fontSize: 16,
                  color: Theme.of(context).primaryColor,
                ),
              ),
            ),
          ],
        ),
        body: Container(
          margin: EdgeInsets.all(16.0),
          child: Form(
            key: _formEditProjectKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                    crossAxisAlignment: CrossAxisAlignment.center,
                    mainAxisAlignment: MainAxisAlignment.start,
                    children: <Widget>[
                      Text('${AppLocalizations.of(context)!.projectName}:',
                          style: TextStyle(fontWeight: FontWeight.bold))
                    ]),
                SizedBox(height: 8),
                TextFormField(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    enabledBorder: OutlineInputBorder(
                      borderSide: BorderSide(
                        color: Theme.of(context).inputDecorationTheme.border
                                is OutlineInputBorder
                            ? (Theme.of(context).inputDecorationTheme.border
                                    as OutlineInputBorder)
                                .borderSide
                                .color
                            : Theme.of(context).dividerColor,
                      ),
                    ),
                    contentPadding:
                        EdgeInsets.symmetric(horizontal: 16, vertical: 16),
                  ),
                  keyboardType: TextInputType.text,
                  initialValue: _project!.name,
                  validator: (value) {
                    return _validateProjectName(value!);
                  },
                  onSaved: (value) => _project!.name = value!,
                  maxLines: 1,
                ),
                SizedBox(height: 24),
                Row(
                    crossAxisAlignment: CrossAxisAlignment.center,
                    mainAxisAlignment: MainAxisAlignment.start,
                    children: <Widget>[
                      Text('${AppLocalizations.of(context)!.projectColor}:',
                          style: TextStyle(fontWeight: FontWeight.bold))
                    ]),
                SizedBox(height: 8),
                DropdownButtonFormField<String>(
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    enabledBorder: OutlineInputBorder(
                      borderSide: BorderSide(
                        color: Theme.of(context).inputDecorationTheme.border
                                is OutlineInputBorder
                            ? (Theme.of(context).inputDecorationTheme.border
                                    as OutlineInputBorder)
                                .borderSide
                                .color
                            : Theme.of(context).dividerColor,
                      ),
                    ),
                    contentPadding:
                        EdgeInsets.symmetric(horizontal: 16, vertical: 16),
                  ),
                  value: _getColorKeyFromHex(_project!.color),
                  items: colors.map((color) {
                    return DropdownMenuItem<String>(
                      value: color['name'],
                      child: Row(
                        children: [
                          Container(
                            width: 24,
                            height: 24,
                            decoration: BoxDecoration(
                              color: _hexToColor(color['hex']!),
                              shape: BoxShape.circle,
                            ),
                          ),
                          SizedBox(width: 8),
                          Text(_getColorName(color['name']!, context)),
                        ],
                      ),
                    );
                  }).toList(),
                  onChanged: (value) {
                    if (value != null) {
                      setState(() {
                        final selectedColor = colors
                            .firstWhere((color) => color['name'] == value);
                        _project!.color = selectedColor['hex']!;
                      });
                    }
                  },
                  onSaved: (value) {
                    if (value != null) {
                      final selectedColor =
                          colors.firstWhere((color) => color['name'] == value);
                      _project!.color = selectedColor['hex']!;
                    }
                  },
                ),
              ],
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
    EventSyncService().synchronizeOnEvent();
  }
}
