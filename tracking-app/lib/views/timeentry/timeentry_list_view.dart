import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:intl/intl.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/models/timeentry.dart';
import 'package:timeasy/repositories/timeentry_repository.dart';
import 'package:timeasy/tools/date_tools.dart';
import 'package:timeasy/tools/duration_formatter.dart';
import 'package:timeasy/tools/excel_export.dart';
import 'package:timeasy/views/timeentry/timeentry_edit_view.dart';

class TimeEntryListView extends StatelessWidget {
  final Project _project;

  TimeEntryListView(this._project);

  @override
  Widget build(BuildContext context) {
    return Scaffold(body: new DataList(_project));
  }
}

class DataList extends StatefulWidget {
  final Project _project;

  DataList(this._project, {Key? key}) : super(key: key);

  @override
  _DataListState createState() {
    return new _DataListState(_project);
  }
}

class _DataListState extends State<DataList> {
  List<TimeEntry>? timeEntries;
  final Project _project;
  Locale? locale;

  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final DurationFormatter _durationFormatter = new DurationFormatter();

  _DataListState(this._project);

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
          title: new Text(AppLocalizations.of(context)!.loadingTimes),
        ),
      );
    } else {
      locale = Localizations.localeOf(context);
      return Scaffold(
        appBar: AppBar(
          title: Text(_getTitle()),
          backgroundColor: Theme.of(context).primaryColor,
          actions: <Widget>[
            TextButton(
              child: Text(
                "Speichern",
                style: Theme.of(context)
                    .textTheme
                    .titleMedium!
                    .copyWith(color: Colors.white),
              ),
              onPressed: () {
                _saveTimeEntries();
              },
            ),
          ],
        ),
        body: SingleChildScrollView(
            scrollDirection: Axis.vertical, child: _dataBody()),
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
          DataColumn(
              label: Text(AppLocalizations.of(context)!.start),
              numeric: false,
              tooltip: AppLocalizations.of(context)!.tooltipTimeStart),
          DataColumn(
              label: Text(AppLocalizations.of(context)!.end),
              numeric: false,
              tooltip: AppLocalizations.of(context)!.tooltipTimeEnd),
          DataColumn(
              label: Text(AppLocalizations.of(context)!.hours),
              numeric: true,
              tooltip: AppLocalizations.of(context)!.tooltipHours),
        ],
        rows: timeEntries!
            .map((timeEntry) => DataRow(
                  cells: [
                    DataCell(
                        Text(timeFormatter
                            .format(timeEntry.startTime.toLocal())), onTap: () {
                      _addOrEditTimeEntry(timeEntryIdToEdit: timeEntry.id);
                    }),
                    DataCell(
                        Text(timeEntry.endTime != null
                            ? timeFormatter.format(timeEntry.endTime!.toLocal())
                            : ""), onTap: () {
                      _addOrEditTimeEntry(timeEntryIdToEdit: timeEntry.id);
                    }),
                    DataCell(
                        Text(timeEntry.endTime != null
                            ? _durationFormatter.formatDuration(timeEntry
                                .endTime!
                                .difference(timeEntry.startTime))
                            : ""), onTap: () {
                      _addOrEditTimeEntry(timeEntryIdToEdit: timeEntry.id);
                    })
                  ],
                ))
            .toList());
  }

  void _addOrEditTimeEntry({String? timeEntryIdToEdit}) {
    Navigator.of(context)
        .push(
      MaterialPageRoute(
        builder: (context) => TimeEntryEditView(_project.id, timeEntryIdToEdit),
        fullscreenDialog: true,
      ),
    )
        .then((value) {
      _loadTimeEntries();
    });
  }

  void _loadTimeEntries() {
    _timeEntryRepository
        .getAllTimeEntries(_project.id)
        .then((List<TimeEntry> value) {
      setState(() {
        timeEntries = value;
      });
    });
  }

  String _getTitle() {
    return "${AppLocalizations.of(context)!.times} (${_project.name})";
  }

  Future<void> _saveTimeEntries() async {
    String? selectedDirectory = await FilePicker.platform.getDirectoryPath();

    if (selectedDirectory != null) {
      showDateRangePicker(
        context: context,
        firstDate: DateTime.fromMillisecondsSinceEpoch(0),
        lastDate: DateTime.now(),
        initialDateRange: _getInitialDataRange(),
      ).then((value) => _exportTimeEntries(selectedDirectory, value!));
    }
  }

  void _exportTimeEntries(String directory, DateTimeRange dateRange) {
    var export = ExcelExport(directory, dateRange, _project.id);
    export.Export();
  }

  DateTimeRange _getInitialDataRange() {
    var dateTools = DateTools();
    var year = DateTime.now().year;
    var weekNumber = dateTools.getWeekNumber(DateTime.now());
    var firstDayOfWeek = dateTools.getFirstDayOfWeek(weekNumber, year);
    var lastDayOfWeek = dateTools.getLastDayOfWeek(weekNumber, year);
    return DateTimeRange(start: firstDayOfWeek, end: lastDayOfWeek);
  }
}
