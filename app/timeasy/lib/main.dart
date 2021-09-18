import 'package:flutter/material.dart';
import 'package:intl/date_symbol_data_local.dart';

import 'package:timeasy/timeentry.dart';
import 'package:timeasy/timeentry_repository.dart';
import 'package:timeasy/timeentry_list_view.dart';
import 'package:timeasy/weekly_view.dart';
import 'package:timeasy/project.dart';
import 'package:timeasy/project_repository.dart';
import 'package:timeasy/project_list_view.dart';
import 'package:flex_color_scheme/flex_color_scheme.dart';

void main() => runApp(MyApp());

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'timeasy',
      // The Mandy red, light theme.
      theme: FlexColorScheme.light(scheme: FlexScheme.deepBlue).toTheme,
      // The Mandy red, dark theme.
      darkTheme: FlexColorScheme.dark(scheme: FlexScheme.deepBlue).toTheme,
      // Use dark or light theme based on system setting.
      themeMode: ThemeMode.system,
      home: MainPage(title: 'timeasy'),
    );
  }
}

class MainPage extends StatefulWidget {
  final String title;

  MainPage({Key key, this.title}) : super(key: key);

  @override
  _MainPageState createState() {
    return new _MainPageState();
  }
}

enum AppState { RUNNING, STOPPED }

class _MainPageState extends State<MainPage> {
  AppState _currentState = AppState.STOPPED;
  Project _currentProject;
  List<Project> _projects;

  final ProjectRepository _projectRepository = new ProjectRepository();
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();

  @override
  void initState() {
    super.initState();
    initializeDateFormatting();

    var projectRepository = new ProjectRepository();
    projectRepository.getLastUsedProjectOrDefault().then((Project project) {
      setState(() {
        _setCurrentProject(project);
      });
      _loadProjects();
      _updateAppState();
    });
  }

  void _setAppState(AppState state) {
    setState(() {
      _currentState = state;
    });
  }

  void _toggleState() {
    switch (_currentState) {
      case AppState.STOPPED:
        _startTiming();
        break;
      case AppState.RUNNING:
        _stopTiming();
        break;
    }
  }

  void _startTiming() async {
    if (_currentProject == null) {
      return;
    }
    var repository = new TimeEntryRepository();
    await repository.closeLatestTimeEntry(_currentProject.id);
    await repository.getLatestOpenTimeEntryOrCreateNew(_currentProject.id);
    _setAppState(AppState.RUNNING);
  }

  void _stopTiming() async {
    if (_currentProject == null) {
      return;
    }
    var repository = new TimeEntryRepository();
    await repository.closeLatestTimeEntry(_currentProject.id);
    _setAppState(AppState.STOPPED);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('timeasy'),
      ),
      drawer: Drawer(
          child: ListView(
        padding: EdgeInsets.zero,
        children: <Widget>[
          DrawerHeader(
            child: Text('timeasy', style: TextStyle(fontWeight: FontWeight.w500, color: Colors.white)),
            decoration: BoxDecoration(
              color: Theme.of(context).primaryColorDark,
            ),
          ),
          ListTile(
            title: Text('Wochenübersicht'),
            onTap: () {
              Navigator.push(context, MaterialPageRoute(builder: (context) => WeeklyView(_currentProject)));
            },
          ),
          ListTile(
              title: Text('Zeiteinträge'),
              onTap: () {
                Navigator.push(context, MaterialPageRoute(builder: (context) => TimeEntryListView(_currentProject)));
              }),
          ListTile(
            title: Text('Projekte'),
            onTap: () {
              Navigator.push(context, MaterialPageRoute(builder: (context) => ProjectListView())).then((_) {
                _loadProjects();
              });
            },
          ),
        ],
      )),
      body: Center(
          child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: <Widget>[
          new RawMaterialButton(
            onPressed: _toggleState,
            child: new Icon(
              _getIcon(),
              color: Theme.of(context).primaryColor,
              size: 128.0,
            ),
            shape: new CircleBorder(),
            elevation: 2.0,
            fillColor: Theme.of(context).backgroundColor,
            padding: const EdgeInsets.all(15.0),
          ),
          _projects == null
              ? Text("Lade Projekte...")
              : new DropdownButton<String>(
                  value: _currentProject.id,
                  items: _projects.map((Project value) {
                    return new DropdownMenuItem<String>(
                      value: value.id,
                      child: new Text(value.name),
                    );
                  }).toList(),
                  onChanged: (String value) {
                    _projectRepository.getProjectById(value).then((Project projectFromDb) {
                      setState(() {
                        _setCurrentProject(projectFromDb);
                      });
                      _updateAppState();
                    });
                  },
                ),
        ],
      )),
    );
  }

  _getIcon() {
    var icon = Icons.add_circle_outline;
    if (_currentProject != null) {
      switch (_currentState) {
        case AppState.STOPPED:
          icon = Icons.play_arrow;
          break;
        case AppState.RUNNING:
          icon = Icons.stop;
          break;
      }
    }
    return icon;
  }

  _loadProjects() {
    _projectRepository.getAllProjects().then((List<Project> projectsFromDb) {
      setState(() {
        _projects = projectsFromDb;
      });
    });
  }

  _setCurrentProject(Project project) {
    _currentProject = project;
    _projectRepository.saveLastUsedProject(project);
  }

  _updateAppState() {
    // Set the current state if there's a timing already running:
    _timeEntryRepository.getLatestOpenTimeEntry(_currentProject.id).then((TimeEntry entry) {
      if (entry != null) {
        _setAppState(AppState.RUNNING);
      } else {
        _setAppState(AppState.STOPPED);
      }
    });
  }
}
