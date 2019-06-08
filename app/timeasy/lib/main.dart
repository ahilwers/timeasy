import 'package:flutter/material.dart';
import 'package:intl/date_symbol_data_local.dart';

import 'package:timeasy/timeentry.dart';
import 'package:timeasy/timeentry_repository.dart';
import 'package:timeasy/timeentrylist.dart';
import 'package:timeasy/weeklyview.dart';
import 'package:timeasy/project.dart';
import 'package:timeasy/project_repository.dart';

void main() => runApp(MyApp());

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'timeasy',
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

enum AppState {
  RUNNING,
  STOPPED
}

class _MainPageState extends State<MainPage> {

  AppState _currentState = AppState.STOPPED;
  Project _currentProject;


  @override
  void initState() {
    super.initState();
    initializeDateFormatting();

    var projectRepository = new ProjectRepository();
    projectRepository.createDefaultProjectIfNotExists().then((Project project) {
      // Create the default project if it does not exist
      setState(() {
        _currentProject = project;
      });
      var timeEntryRepository = new TimeEntryRepository();
      // Set the current state if there's a timing already running:
      timeEntryRepository.getLatestOpenTimeEntry(_currentProject.id).then((TimeEntry entry) {
        if (entry != null) {
          _setAppState(AppState.RUNNING);
        }
      });
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
    if (_currentProject==null) {
      return;
    }
    var repository = new TimeEntryRepository();
    await repository.closeLatestTimeEntry(_currentProject.id);
    await repository.getLatestOpenTimeEntryOrCreateNew(_currentProject.id);
    _setAppState(AppState.RUNNING);
  }

  void _stopTiming() async {
    if (_currentProject==null) {
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
              child: Text('Drawer Header'),
              decoration: BoxDecoration(
                color: Colors.blue,
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
                Navigator.push(context, MaterialPageRoute(builder: (context) => TimeEntryList(_currentProject)));
              },
            ),
            ListTile(
              title: Text('Projekte'),
              onTap: () {
                // Update the state of the app
                // ...
                // Then close the drawer
                Navigator.pop(context);
              },
            ),
          ],
        )
      ),
      body: Center(
        child : new RawMaterialButton(
          onPressed: _toggleState,
          child: new Icon(
            _getIcon(),
            color: Colors.blue,
            size: 128.0,
          ),
          shape: new CircleBorder(),
          elevation: 2.0,
          fillColor: Colors.white,
          padding: const EdgeInsets.all(15.0),
        ),
      ),
    );
  }

  _getIcon() {
    var icon = Icons.add_circle_outline;
    if (_currentProject!=null) {
      switch (_currentState) {
        case AppState.STOPPED :
          icon = Icons.play_arrow;
          break;
        case AppState.RUNNING :
          icon = Icons.stop;
          break;
      }
    }
    return icon;
  }

}
