import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:intl/intl.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/models/timeentry.dart';
import 'package:timeasy/repositories/timeentry_repository.dart';
import 'package:timeasy/tools/date_tools.dart';
import 'package:timeasy/tools/duration_formatter.dart';
import 'package:timeasy/tools/excel_export.dart';
import 'package:timeasy/tools/excel_export_onelineperday.dart';
import 'package:timeasy/tools/excel_export_allentries.dart';
import 'package:timeasy/views/timeentry/timeentry_edit_view.dart';

enum ExportType {
  AllEntries,
  OneLinePerDay,
}

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
  DateTimeRange? _currentDateRange;
  final Project _project;
  Locale? locale;
  String exportMessage = "";
  String exportDate = "";
  String exportStart = "";
  String exportEnd = "";
  String exportPause = "";

  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final DurationFormatter _durationFormatter = new DurationFormatter();

  _DataListState(this._project);

  @override
  void initState() {
    super.initState();
    _loadTimeEntries(_getInitialDateRange());
  }

  @override
  Widget build(BuildContext context) {
    if (timeEntries == null) {
      exportMessage = AppLocalizations.of(context)!.timesExported;
      exportDate = AppLocalizations.of(context)!.date;
      exportStart = AppLocalizations.of(context)!.start;
      exportEnd = AppLocalizations.of(context)!.end;
      exportPause = AppLocalizations.of(context)!.pause;
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
                AppLocalizations.of(context)!.export,
                style: Theme.of(context)
                    .textTheme
                    .titleMedium!
                    .copyWith(color: Colors.white),
              ),
              onPressed: () {
                _showExportDialog(context);
              },
            ),
          ],
        ),
        body: Column(
          children: <Widget>[
            TextButton(
              child: Text(
                _getCurrentDateRangeText(),
                style: Theme.of(context)
                    .textTheme
                    .titleMedium!
                    .copyWith(fontWeight: FontWeight.bold),
              ),
              onPressed: _selectDateRange,
            ),
            SingleChildScrollView(
                scrollDirection: Axis.vertical, child: _dataBody()),
          ],
        ),
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

  _getCurrentDateRangeText() {
    var dateRange = _currentDateRange ?? _getInitialDateRange();
    var formatter = new DateFormat.yMd(locale.toString());
    return formatter.format(dateRange.start) +
        " - " +
        formatter.format(dateRange.end);
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
      _loadTimeEntries(_currentDateRange ?? _getInitialDateRange());
    });
  }

  void _loadTimeEntries(DateTimeRange dateRange) {
    _currentDateRange = dateRange;
    _timeEntryRepository
        .getTimeEntries(_project.id, dateRange.start, dateRange.end)
        .then((List<TimeEntry> value) {
      setState(() {
        timeEntries = value;
      });
    });
  }

  String _getTitle() {
    return "${AppLocalizations.of(context)!.times} (${_project.name})";
  }

  Future<void> _saveTimeEntries(
      BuildContext context, ExportType exportType) async {
    String? selectedDirectory = await FilePicker.platform.getDirectoryPath();
    if (selectedDirectory == null) {
      return;
    }
    var dateRange = _currentDateRange ?? _getInitialDateRange();
    _exportTimeEntries(context, exportType, selectedDirectory, dateRange);
  }

  void _exportTimeEntries(BuildContext context, ExportType exportType,
      String directory, DateTimeRange dateRange) {
    var filename = _generateFilename(dateRange);
    ExcelExport? export = null;
    switch (exportType) {
      case ExportType.AllEntries:
        export =
            ExcelExportAllEntries(directory, filename, dateRange, _project.id);
        break;
      case ExportType.OneLinePerDay:
        export = ExcelExportOneLinePerDay(
            directory, filename, dateRange, _project.id);
        break;
      default:
    }
    export?.addTranslation("date", exportDate);
    export?.addTranslation("start", exportStart);
    export?.addTranslation("end", exportEnd);
    export?.addTranslation("pause", exportPause);

    export?.Export().then((value) => showToast(exportMessage));
  }

  String _generateFilename(DateTimeRange dateRange) {
    var fromDate =
        "${dateRange.start.year}-${dateRange.start.month.toString().padLeft(2, '0')}-${dateRange.start.day.toString().padLeft(2, '0')}";
    var toDate =
        "${dateRange.end.year}-${dateRange.end.month.toString().padLeft(2, '0')}-${dateRange.end.day.toString().padLeft(2, '0')}";
    return "${fromDate} - ${toDate} ${_project.name}.xlsx";
  }

  void showToast(String message) {
    Fluttertoast.showToast(
      msg: message,
      toastLength: Toast.LENGTH_SHORT,
      gravity: ToastGravity.BOTTOM,
      timeInSecForIosWeb: 1,
      backgroundColor: Colors.black54,
      textColor: Colors.white,
      fontSize: 16.0,
    );
  }

  DateTimeRange _getInitialDateRange() {
    var dateTools = DateTools();
    var year = DateTime.now().year;
    var weekNumber = dateTools.getWeekNumber(DateTime.now());
    var lastDayOfWeek = dateTools.getLastDayOfWeek(weekNumber, year);
    var firstDayOfWeek = lastDayOfWeek.subtract(new Duration(days: 31));
    return DateTimeRange(start: firstDayOfWeek, end: lastDayOfWeek);
  }

  void _selectDateRange() {
    var dateRange = _currentDateRange ?? _getInitialDateRange();
    showDateRangePicker(
      context: context,
      initialDateRange: dateRange,
      firstDate: DateTime.fromMillisecondsSinceEpoch(0),
      lastDate: dateRange.end,
    ).then((value) => _loadTimeEntries(value ?? dateRange));
  }

  void _showExportDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(15.0),
          ),
          title: Text('Wähle eine Option'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  minimumSize: Size(double.infinity, 50),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(10.0),
                  ),
                ),
                onPressed: () {
                  Navigator.of(context).pop(); // Close the dialog
                  _saveTimeEntries(
                      context, ExportType.AllEntries); // Call the method
                },
                child: Text('Alle Zeiteinträge'),
              ),
              SizedBox(height: 10),
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  minimumSize: Size(double.infinity, 50),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(10.0),
                  ),
                ),
                onPressed: () {
                  Navigator.of(context).pop(); // Close the dialog
                  _saveTimeEntries(
                      context, ExportType.OneLinePerDay); // Call the method
                },
                child: Text('Eine Zeile pro Tag'),
              ),
            ],
          ),
        );
      },
    );
  }
}
