import 'package:flutter/material.dart';
import 'package:timeasy/timeentry_repository.dart';
import 'package:timeasy/timeentry.dart';
import 'package:timeasy/project.dart';

class TimeEntryListView extends StatelessWidget {

  Project _project;

  TimeEntryListView(Project project) {
    _project = project;
  }

  @override
  Widget build(BuildContext context) {

    return Scaffold(
      body: new DataList(_project)
    );
  }

}

class DataList extends StatefulWidget {

  Project _project;

  DataList(Project project, {Key key}) : super(key: key) {
    _project = project;
  }

  @override
  _DataListState createState() {
    return new _DataListState(_project);
  }

}

class _DataListState extends State<DataList> {

  List<TimeEntry> timeEntries;
  Project _project;

  _DataListState(Project project) {
    _project = project;
  }

  @override
  void initState() {
    super.initState();
    var timeEntryRepository = new TimeEntryRepository();
    timeEntryRepository.getAllTimeEntries(_project.id).then((List<TimeEntry> value) {
      setState(() {
        timeEntries = value;
      });
    });
  }

  @override
  Widget build(BuildContext context) {
    if (timeEntries == null) {
      return Scaffold(
        appBar: new AppBar(
          title: new Text("Lade Zeiten..."),
        ),
      );
    } else {
      return Scaffold(
        appBar: AppBar(
          title: Text(_getTitle()),
        ),
        body: Column(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          verticalDirection: VerticalDirection.down,
          children: <Widget>[
            Center (
                child: _dataBody()
            ),
          ],
        ),
      );
    }
  }

  _dataBody() {
    return DataTable(
      columns: [
        DataColumn(
            label: Text("Start"),
            numeric: false,
            tooltip: "The start time"
        ),
        DataColumn(
            label: Text("Ende"),
            numeric: false,
            tooltip: "The end time"
        ),
        DataColumn(
            label: Text("Stunden"),
            numeric: true,
            tooltip: "Anzahl der Stunden"
        ),
      ],
      rows: timeEntries.map((timeEntry) => DataRow(
        cells: [
          DataCell(
            Text(timeEntry.startTime.toLocal().toIso8601String()),
          ),
          DataCell(
            Text(timeEntry.endTime != null ? timeEntry.endTime.toLocal().toIso8601String() : ""),
          ),
          DataCell(
            Text(timeEntry.endTime != null ? timeEntry.endTime.difference(timeEntry.startTime).inMinutes.toString() : ""),
          )
        ]
      )).toList()
    );

  }

  String _getTitle() {
    return "Zeiten (${_project.name})";
  }


}