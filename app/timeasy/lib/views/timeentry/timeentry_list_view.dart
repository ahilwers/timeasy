import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:timeasy/tools/duration_formatter.dart';
import 'package:timeasy/repositories/timeentry_repository.dart';
import 'package:timeasy/models/timeentry.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/views/timeentry/timeentry_edit_view.dart';

class TimeEntryListView extends StatelessWidget {
  Project _project;

  TimeEntryListView(Project project) {
    _project = project;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(body: new DataList(_project));
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
  Locale locale;

  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final DurationFormatter _durationFormatter = new DurationFormatter();

  _DataListState(Project project) {
    _project = project;
  }

  @override
  void initState() {
    super.initState();
    _loadTimeEntries();
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
      locale = Localizations.localeOf(context);
      return Scaffold(
        appBar: AppBar(
          title: Text(_getTitle()),
        ),
        body: SingleChildScrollView(scrollDirection: Axis.vertical, child: _dataBody()),
        floatingActionButton: FloatingActionButton(
          onPressed: () {
            _addOrEditTimeEntry();
          },
          child: Icon(Icons.add),
          backgroundColor: Theme.of(context).primaryColor,
        ),
      );
    }
  }

  _dataBody() {
    var timeFormatter = new DateFormat.yMd(locale.toString()).add_Hm();
    return DataTable(
        columns: [
          DataColumn(label: Text("Start"), numeric: false, tooltip: "The start time"),
          DataColumn(label: Text("Ende"), numeric: false, tooltip: "The end time"),
          DataColumn(label: Text("Stunden"), numeric: true, tooltip: "Anzahl der Stunden"),
        ],
        rows: timeEntries
            .map((timeEntry) => DataRow(
                  cells: [
                    DataCell(Text(timeFormatter.format(timeEntry.startTime.toLocal())), onTap: () {
                      _addOrEditTimeEntry(timeEntryIdToEdit: timeEntry.id);
                    }),
                    DataCell(Text(timeEntry.endTime != null ? timeFormatter.format(timeEntry.endTime.toLocal()) : ""), onTap: () {
                      _addOrEditTimeEntry(timeEntryIdToEdit: timeEntry.id);
                    }),
                    DataCell(Text(timeEntry.endTime != null ? _durationFormatter.formatDuration(timeEntry.endTime.difference(timeEntry.startTime)) : ""), onTap: () {
                      _addOrEditTimeEntry(timeEntryIdToEdit: timeEntry.id);
                    })
                  ],
                ))
            .toList());
  }

  void _addOrEditTimeEntry({String timeEntryIdToEdit}) {
    Navigator.of(context)
        .push(
      MaterialPageRoute(
        builder: (context) => TimeEntryEditView(_project.id, timeEntryId: timeEntryIdToEdit),
        fullscreenDialog: true,
      ),
    )
        .then((value) {
      _loadTimeEntries();
    });
  }

  void _loadTimeEntries() {
    _timeEntryRepository.getAllTimeEntries(_project.id).then((List<TimeEntry> value) {
      setState(() {
        timeEntries = value;
      });
    });
  }

  String _getTitle() {
    return "Zeiten (${_project.name})";
  }
}