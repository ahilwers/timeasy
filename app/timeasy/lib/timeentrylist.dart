import 'package:flutter/material.dart';
import 'package:timeasy/timeentry_repository.dart';
import 'package:timeasy/timeentry.dart';

class TimeEntryList extends StatelessWidget {

  @override
  Widget build(BuildContext context) {

    return Scaffold(
      body: new DataList(title: "timeasy - Zeiten")
    );
  }

}

class DataList extends StatefulWidget {

  final String title;

  DataList({Key key, this.title}) : super(key: key);

  @override
  _DataListState createState() {
    return new _DataListState();
  }

}

class _DataListState extends State<DataList> {

  List<TimeEntry> timeEntries;

  @override
  void initState() {
    super.initState();
    var timeEntryRepository = new TimeEntryRepository();
    timeEntryRepository.getAllTimeEntries().then((List<TimeEntry> value) {
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
          title: Text("timeasy - Zeiten"),
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
      ],
      rows: timeEntries.map((timeEntry) => DataRow(
        cells: [
          DataCell(
            Text(timeEntry.startTime.toLocal().toIso8601String()),
          ),
          DataCell(
            Text(timeEntry.endTime != null ? timeEntry.endTime.toLocal().toIso8601String() : ""),
          )
        ]
      )).toList()
    );

  }


}