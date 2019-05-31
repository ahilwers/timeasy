import 'package:flutter/material.dart';
import 'package:timeasy/timeentry_repository.dart';

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
        stopTiming();
        break;
    }
  }

  void _startTiming() async {
    var repository = new TimeEntryRepository();
    await repository.closeLatestTimeEntry();
    await repository.getLatestOpenTimeEntryOrCreateNew();
    _setAppState(AppState.RUNNING);
  }

  void stopTiming() async {
    var repository = new TimeEntryRepository();
    await repository.closeLatestTimeEntry();
    _setAppState(AppState.STOPPED);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('timeasy'),
      ),
      body: Center(
        child : new RawMaterialButton(
          onPressed: _toggleState,
          child: _currentState==AppState.STOPPED ? new Icon(
            Icons.play_arrow,
            color: Colors.blue,
            size: 128.0,
          ) : new Icon(
            Icons.stop,
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

}
