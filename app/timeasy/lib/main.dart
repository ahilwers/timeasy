import 'package:flutter/material.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:flutter_localizations/flutter_localizations.dart';

import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import 'package:timeasy/models/timeentry.dart';
import 'package:timeasy/repositories/timeentry_repository.dart';
import 'package:timeasy/views/imprint.dart';
import 'package:timeasy/views/timeentry/timeentry_list_view.dart';
import 'package:timeasy/views/statistics/weekly_view.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/views/project/project_list_view.dart';
import 'package:flex_color_scheme/flex_color_scheme.dart';

void main() => runApp(MyApp());

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'timeasy',
      localizationsDelegates: [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: [
        Locale('en', ''),
        Locale('de', ''),
      ],
      theme: FlexColorScheme.light(scheme: FlexScheme.deepBlue).toTheme,
      darkTheme: FlexColorScheme.dark(scheme: FlexScheme.bahamaBlue).toTheme,
      // Use dark or light theme based on system setting.
      themeMode: ThemeMode.system,
      home: MainPage(title: 'timeasy'),
    );
  }
}

class MainPage extends StatefulWidget {
  final String? title;

  MainPage({Key? key, this.title}) : super(key: key);

  @override
  _MainPageState createState() {
    return new _MainPageState();
  }
}

enum AppState { RUNNING, STOPPED }

class _MainPageState extends State<MainPage> with SingleTickerProviderStateMixin {
  AppState _currentState = AppState.STOPPED;
  late Project _currentProject;
  List<Project>? _projects;

  final ProjectRepository _projectRepository = new ProjectRepository();
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  late AnimationController buttonAnimationController;

  @override
  void initState() {
    super.initState();
    initializeDateFormatting();

    buttonAnimationController = new AnimationController(
      vsync: this,
      duration: Duration(milliseconds: 1000),
    );

    var projectRepository = new ProjectRepository();
    projectRepository.getLastUsedProjectOrDefault("Project 1").then((Project project) {
      setState(() {
        _setCurrentProject(project);
      });
      _loadProjects();
      _updateAppState();
    });
  }

  void _setAppState(AppState state) {
    switch (state) {
      case AppState.RUNNING:
        if (_currentState == AppState.STOPPED) {
          buttonAnimationController.forward();
        }
        break;
      case AppState.STOPPED:
        if (_currentState == AppState.RUNNING) {
          buttonAnimationController.reverse();
        }
        break;
    }
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
    var repository = new TimeEntryRepository();
    await repository.closeLatestTimeEntry(_currentProject.id);
    await repository.getLatestOpenTimeEntryOrCreateNew(_currentProject.id);
    _setAppState(AppState.RUNNING);
  }

  void _stopTiming() async {
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
              child: Image.asset("assets/hourglass_lightgrey.png"), // Text('timeasy', style: TextStyle(fontWeight: FontWeight.w500, color: Colors.white)),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.centerRight,
                  end: Alignment.centerLeft,
                  colors: [
                    Color(0xff28b0fe),
                    Color(0xffc80eef),
                  ],
                ),
              ),
            ),
            ListTile(
              title: Text(AppLocalizations.of(context)!.weeklyOverview),
              onTap: () {
                Navigator.push(context, MaterialPageRoute(builder: (context) => WeeklyView(_currentProject)));
              },
            ),
            ListTile(
              title: Text(AppLocalizations.of(context)!.timeEntries),
              onTap: () {
                Navigator.push(context, MaterialPageRoute(builder: (context) => TimeEntryListView(_currentProject))).then(
                  (_) {
                    _updateAppState();
                  },
                );
              },
            ),
            ListTile(
              title: Text(AppLocalizations.of(context)!.projects),
              onTap: () {
                Navigator.push(context, MaterialPageRoute(builder: (context) => ProjectListView())).then(
                  (_) {
                    _loadProjects();
                  },
                );
              },
            ),
            ListTile(
              title: Text(AppLocalizations.of(context)!.info),
              onTap: () {
                Navigator.push(context, MaterialPageRoute(builder: (context) => Imprint()));
              },
            ),
          ],
        ),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            new RawMaterialButton(
              onPressed: _toggleState,
              child: Container(
                child: new AnimatedIcon(
                  icon: AnimatedIcons.play_pause,
                  color: Colors.white,
                  size: 148.0,
                  progress: buttonAnimationController,
                ),
                decoration: ShapeDecoration(
                  shape: new CircleBorder(),
                  gradient: LinearGradient(
                    begin: Alignment.centerRight,
                    end: Alignment.centerLeft,
                    colors: [
                      Color(0xff28b0fe),
                      Color(0xffc80eef),
                    ],
                  ),
                ),
              ),
              shape: new CircleBorder(),
              elevation: 2.0,
              //fillColor: Theme.of(context).primaryColor,
              //padding: const EdgeInsets.all(15.0),
            ),
            _projects == null
                ? Text(AppLocalizations.of(context)!.loadingProject)
                : new DropdownButton<String>(
                    value: _currentProject.id,
                    items: _projects!.map(
                      (Project value) {
                        return new DropdownMenuItem<String>(
                          value: value.id,
                          child: new Text(value.name),
                        );
                      },
                    ).toList(),
                    onChanged: (String? value) {
                      _projectRepository.getProjectById(value!).then(
                        (Project? projectFromDb) {
                          setState(
                            () {
                              _setCurrentProject(projectFromDb!);
                            },
                          );
                          _updateAppState();
                        },
                      );
                    },
                  ),
          ],
        ),
      ),
    );
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
    _timeEntryRepository.getLatestOpenTimeEntry(_currentProject.id).then((TimeEntry? entry) {
      if (entry != null) {
        _setAppState(AppState.RUNNING);
      } else {
        _setAppState(AppState.STOPPED);
      }
    });
  }
}
