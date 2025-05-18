import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:intl/intl.dart';
import 'package:timeasy/bloc/selected_project/selected_project_bloc.dart';
import 'package:timeasy/bloc/selected_project/selected_project_state.dart';
import 'package:timeasy/components/project_header_component.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/models/time_entry.dart';
import 'package:timeasy/repositories/time_entry_repository.dart';
import 'package:timeasy/tools/date_tools.dart';
import 'package:timeasy/tools/duration_formatter.dart';
import 'package:timeasy/views/timeentry/time_entry_edit_view.dart';

class TimeEntryListView extends StatelessWidget {
  final Project _project;

  TimeEntryListView(this._project);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: BlocBuilder<SelectedProjectBloc, SelectedProjectState>(
        builder: (context, state) {
          Project projectToUse = _project;

          if (state is SelectedProjectSet && state.project != null) {
            projectToUse = state.project!;
          }

          return DataList(projectToUse);
        },
      ),
    );
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
      return Scaffold(
        appBar: ProjectHeader(
          actions: [
            IconButton(
              icon: Icon(Icons.add),
              onPressed: () => _addOrEditTimeEntry(),
            ),
          ],
          showSettingsButton: false, // Explicitly disable settings button
        ),
        body: Center(
          child: CircularProgressIndicator(),
        ),
      );
    } else {
      locale = Localizations.localeOf(context);
      return Scaffold(
        appBar: ProjectHeader(
          actions: [
            IconButton(
              icon: Icon(Icons.add),
              onPressed: () => _addOrEditTimeEntry(),
            ),
          ],
          showSettingsButton: false, // Explicitly disable settings button
        ),
        body: Column(
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16.0),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  TextButton(
                    onPressed: _selectDateRange,
                    child: Row(
                      children: [
                        Icon(Icons.date_range),
                        SizedBox(width: 8),
                        Text(_getCurrentDateRangeText()),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            Padding(
              padding:
                  const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
              child: Container(
                padding: EdgeInsets.symmetric(vertical: 12.0, horizontal: 16.0),
                decoration: BoxDecoration(
                  color: Theme.of(context).brightness == Brightness.dark
                      ? Colors.grey[800]
                      : Colors.grey[200],
                  borderRadius: BorderRadius.circular(8.0),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      AppLocalizations.of(context)!.weeklyHourSum,
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 16,
                      ),
                    ),
                    Text(
                      _calculateTotalHours(),
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 16,
                        color: Theme.of(context).primaryColor,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                child: _buildTimeEntryList(),
              ),
            ),
          ],
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

  _buildTimeEntryList() {
    var dateFormatter = new DateFormat.yMd(locale.toString());
    var timeFormatter = new DateFormat.Hm(locale.toString());

    if (timeEntries!.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Text(
            AppLocalizations.of(context)!.noTimeEntries,
            style: TextStyle(fontSize: 16),
          ),
        ),
      );
    }

    return ListView.builder(
      shrinkWrap: true,
      physics: NeverScrollableScrollPhysics(),
      itemCount: timeEntries!.length,
      itemBuilder: (context, index) {
        final timeEntry = timeEntries![index];
        final formattedStartDate =
            dateFormatter.format(timeEntry.startTime.toLocal());
        final formattedStartTime =
            timeFormatter.format(timeEntry.startTime.toLocal());

        final formattedEndDate = timeEntry.endTime != null
            ? dateFormatter.format(timeEntry.endTime!.toLocal())
            : "";
        final formattedEndTime = timeEntry.endTime != null
            ? timeFormatter.format(timeEntry.endTime!.toLocal())
            : "";

        // Calculate duration - use current time if endTime is null
        final duration = timeEntry.endTime != null
            ? _durationFormatter.formatDuration(
                timeEntry.endTime!.difference(timeEntry.startTime))
            : _durationFormatter
                .formatDuration(DateTime.now().difference(timeEntry.startTime));

        return Dismissible(
          key: Key(timeEntry.id ?? index.toString()),
          background: Container(
            color: Colors.blue,
            alignment: Alignment.centerLeft,
            padding: EdgeInsets.symmetric(horizontal: 20),
            child: Icon(
              Icons.edit,
              color: Colors.white,
            ),
          ),
          secondaryBackground: Container(
            color: Colors.red,
            alignment: Alignment.centerRight,
            padding: EdgeInsets.symmetric(horizontal: 20),
            child: Icon(
              Icons.delete,
              color: Colors.white,
            ),
          ),
          confirmDismiss: (direction) async {
            if (direction == DismissDirection.endToStart) {
              // Delete action
              final bool? result = await showDialog<bool>(
                context: context,
                builder: (BuildContext context) {
                  return AlertDialog(
                    title: Text(AppLocalizations.of(context)!.deleteTimeEntry),
                    content: Text(AppLocalizations.of(context)!
                        .deleteTimeEntryConfirmation),
                    actions: <Widget>[
                      TextButton(
                        onPressed: () => Navigator.of(context).pop(false),
                        child: Text(AppLocalizations.of(context)!.cancel),
                      ),
                      TextButton(
                        onPressed: () => Navigator.of(context).pop(true),
                        child: Text(AppLocalizations.of(context)!.delete),
                      ),
                    ],
                  );
                },
              );
              return result ?? false;
            } else {
              // Edit action - don't actually dismiss, just navigate
              _addOrEditTimeEntry(timeEntryIdToEdit: timeEntry.id);
              return false;
            }
          },
          onDismissed: (direction) {
            if (direction == DismissDirection.endToStart) {
              // Store a copy of the time entry for potential undo
              final deletedTimeEntry = timeEntry;
              final deletedIndex = index;

              // Delete the item
              _timeEntryRepository.deleteTimeEntryById(timeEntry.id!).then((_) {
                setState(() {
                  timeEntries!.removeAt(index);
                });
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content:
                        Text(AppLocalizations.of(context)!.timeEntryDeleted),
                    action: SnackBarAction(
                      label: AppLocalizations.of(context)!.undo,
                      onPressed: () {
                        // Undo deletion
                        _timeEntryRepository
                            .addTimeEntry(deletedTimeEntry)
                            .then((_) {
                          _loadTimeEntries(
                              _currentDateRange ?? _getInitialDateRange());
                        });
                      },
                    ),
                  ),
                );
              });
            }
          },
          child: Card(
            margin: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            child: InkWell(
              onTap: () {
                // Navigate to edit screen when tapping on the card
                _addOrEditTimeEntry(timeEntryIdToEdit: timeEntry.id);
              },
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              '${AppLocalizations.of(context)!.start}:',
                              style: TextStyle(
                                fontWeight: FontWeight.bold,
                                fontSize: 14,
                              ),
                            ),
                            SizedBox(height: 4),
                            Text(
                              '$formattedStartDate, $formattedStartTime',
                              style: TextStyle(fontSize: 16),
                            ),
                          ],
                        ),
                        if (timeEntry.endTime != null)
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Text(
                                '${AppLocalizations.of(context)!.end}:',
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 14,
                                ),
                              ),
                              SizedBox(height: 4),
                              Text(
                                '$formattedEndDate, $formattedEndTime',
                                style: TextStyle(fontSize: 16),
                              ),
                            ],
                          ),
                      ],
                    ),
                    if (timeEntry.description != null &&
                        timeEntry.description!.isNotEmpty)
                      Padding(
                        padding: const EdgeInsets.only(top: 8.0),
                        child: Text(
                          timeEntry.description!,
                          style: TextStyle(fontSize: 16),
                        ),
                      ),
                    // Always show hours - for ongoing entries, use current time for calculation
                    Padding(
                      padding: const EdgeInsets.only(top: 12.0),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          Text(
                            '${AppLocalizations.of(context)!.hours}: ',
                            style: TextStyle(
                              fontWeight: FontWeight.bold,
                              fontSize: 14,
                            ),
                          ),
                          Text(
                            duration,
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          // Add a small indicator for ongoing entries
                          if (timeEntry.endTime == null)
                            Padding(
                              padding: const EdgeInsets.only(left: 4.0),
                              child: Icon(
                                Icons.update,
                                size: 16,
                                color: Theme.of(context).colorScheme.primary,
                              ),
                            ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
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

  String _calculateTotalHours() {
    if (timeEntries == null || timeEntries!.isEmpty) {
      return "0:00";
    }

    Duration totalDuration = Duration.zero;

    for (var entry in timeEntries!) {
      if (entry.endTime != null) {
        // For completed entries, use the actual duration
        totalDuration += entry.endTime!.difference(entry.startTime);
      } else {
        // For ongoing entries, calculate duration up to now
        totalDuration += DateTime.now().difference(entry.startTime);
      }
    }

    return _durationFormatter.formatDuration(totalDuration);
  }
}
