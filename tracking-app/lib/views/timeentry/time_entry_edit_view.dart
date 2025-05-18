import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:intl/intl.dart';
import 'package:timeasy/components/custom_datetime_picker.dart';
import 'package:timeasy/models/time_entry.dart';
import 'package:timeasy/repositories/time_entry_repository.dart';

enum ConfirmAction { CANCEL, ACCEPT }

class TimeEntryEditView extends StatelessWidget {
  final String? _timeEntryId;
  final String _projectId;

  TimeEntryEditView(this._projectId, [this._timeEntryId]);

  @override
  Widget build(BuildContext context) {
    return Scaffold(body: new TimeEntryEditWidget(_projectId, _timeEntryId));
  }
}

class TimeEntryEditWidget extends StatefulWidget {
  final String? _timeEntryId;
  final String _projectId;

  TimeEntryEditWidget(this._projectId, this._timeEntryId);

  @override
  _TimeEntryEditWidgetState createState() {
    return new _TimeEntryEditWidgetState(_projectId, _timeEntryId);
  }
}

class _TimeEntryEditWidgetState extends State<TimeEntryEditWidget> {
  TimeEntry? _timeEntry;
  bool _endTimeWasEmpty = false;
  final String? _timeEntryId;
  final String _projectId;
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final _formEditTimeEntryKey = GlobalKey<FormState>();

  _TimeEntryEditWidgetState(this._projectId, this._timeEntryId);

  @override
  void initState() {
    super.initState();
    if (_timeEntryId != null) {
      _timeEntryRepository
          .getTimeEntryById(_timeEntryId!)
          .then((TimeEntry? timeEntryFromDb) {
        setState(() {
          _timeEntry = timeEntryFromDb;
          _endTimeWasEmpty = _timeEntry!.endTime ==
              null; // Indicates that we're editing a time entry that is not completed yet
        });
      });
    } else {
      _timeEntry = new TimeEntry(_projectId);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_timeEntry == null) {
      return Scaffold(
        appBar: AppBar(
          title: Text(
            AppLocalizations.of(context)!.loadingTimeEntry,
            style: TextStyle(
              fontSize: 22,
              fontWeight: FontWeight.bold,
              color: Theme.of(context).brightness == Brightness.dark ? Colors.white : Colors.black,
            ),
          ),
          backgroundColor: Theme.of(context).brightness == Brightness.dark ? Colors.black : Colors.white,
          elevation: 0,
          iconTheme: IconThemeData(
            color: Theme.of(context).brightness == Brightness.dark ? Colors.white : Colors.black,
          ),
          leading: IconButton(
            icon: Icon(Icons.arrow_back),
            onPressed: () => Navigator.pop(context),
          ),
        ),
      );
    } else {
      Locale locale = Localizations.localeOf(context);
      var dateFormatter = new DateFormat.yMd(locale.toString());
      var timeFormatter = new DateFormat.Hm(locale.toString());
      return Scaffold(
          appBar: AppBar(
            title: Text(
              _getTitle(),
              style: TextStyle(
                fontSize: 22,
                fontWeight: FontWeight.bold,
                color: Theme.of(context).brightness == Brightness.dark ? Colors.white : Colors.black,
              ),
            ),
            backgroundColor: Theme.of(context).brightness == Brightness.dark ? Colors.black : Colors.white,
            elevation: 0,
            iconTheme: IconThemeData(
              color: Theme.of(context).brightness == Brightness.dark ? Colors.white : Colors.black,
            ),
            leading: IconButton(
              icon: Icon(Icons.arrow_back),
              onPressed: () => Navigator.pop(context),
            ),
            centerTitle: true,
            actions: <Widget>[
              // Save button with icon
              IconButton(
                icon: Icon(
                  Icons.save,
                  color: Theme.of(context).brightness == Brightness.dark ? Colors.white : Colors.black,
                ),
                tooltip: AppLocalizations.of(context)!.save,
                onPressed: () {
                  final form = _formEditTimeEntryKey.currentState;
                  if (form!.validate()) {
                    // We need to validate the timeEntry separately
                    var errorMessage = "";
                    if (_timeEntry?.startTime == null) {
                      errorMessage =
                          AppLocalizations.of(context)!.errorMissingStartTime;
                    } else if (_needToSetEndTime()) {
                      errorMessage =
                          AppLocalizations.of(context)!.errorMissingEndTime;
                    } else if ((_timeEntry!.endTime != null) &&
                        (_timeEntry!.endTime!
                            .isBefore(_timeEntry!.startTime))) {
                      errorMessage = AppLocalizations.of(context)!
                          .errorEndtimeNotAfterStartTime;
                    }
                    if (errorMessage != "") {
                      ScaffoldMessenger.of(context)
                          .showSnackBar(SnackBar(content: Text(errorMessage)));
                    } else {
                      _saveProject(form);
                      Navigator.pop(context);
                    }
                  }
                },
              ),
              // Delete button with icon (only shown when editing an existing time entry)
              if (_timeEntryId != null)
                IconButton(
                  icon: Icon(
                    Icons.delete,
                    color: Theme.of(context).brightness == Brightness.dark ? Colors.white : Colors.black,
                  ),
                  tooltip: AppLocalizations.of(context)!.delete,
                  onPressed: () {
                    deleteTimeEntryWithRequest(context);
                  },
                ),
            ],
          ),
          body: Container(
            margin: EdgeInsets.all(16.0),
            child: Form(
                key: _formEditTimeEntryKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Row(
                        crossAxisAlignment: CrossAxisAlignment.center,
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: <Widget>[
                          Text('${AppLocalizations.of(context)!.start}:',
                              style: TextStyle(fontWeight: FontWeight.bold))
                        ]),
                    CustomDateTimePicker(
                      dateTime: _timeEntry!.startTime.toLocal(),
                      onDateTimeChanged: (DateTime newDateTime) {
                        setState(() {
                          _timeEntry!.startTime = newDateTime.toUtc();
                          // Also set the end time automatically if it's not already set:
                          if (_needToSetEndTime()) {
                            _timeEntry!.endTime = _timeEntry!.startTime;
                          }
                        });
                      },
                    ),
                    SizedBox(height: 16),
                    Row(
                        crossAxisAlignment: CrossAxisAlignment.center,
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: <Widget>[
                          Text('${AppLocalizations.of(context)!.end}:',
                              style: TextStyle(fontWeight: FontWeight.bold))
                        ]),
                    CustomDateTimePicker(
                      dateTime: _timeEntry!.endTime?.toLocal(),
                      dateHint: AppLocalizations.of(context)!.endDate,
                      timeHint: AppLocalizations.of(context)!.endTime,
                      onDateTimeChanged: (DateTime newDateTime) {
                        setState(() {
                          _timeEntry!.endTime = newDateTime.toUtc();
                        });
                      },
                    ),
                    SizedBox(height: 24),
                    Row(
                        crossAxisAlignment: CrossAxisAlignment.center,
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: <Widget>[
                          Text(
                              '${AppLocalizations.of(context)!.entryDescription}:',
                              style: TextStyle(fontWeight: FontWeight.bold))
                        ]),
                    SizedBox(height: 8),
                    TextFormField(
                      initialValue: _timeEntry!.description ?? '',
                      decoration: InputDecoration(
                        hintText: AppLocalizations.of(context)!.whatDidYouDo,
                        border: OutlineInputBorder(),
                        enabledBorder: OutlineInputBorder(
                          borderSide: BorderSide(
                            color: Theme.of(context).inputDecorationTheme.border
                                    is OutlineInputBorder
                                ? (Theme.of(context).inputDecorationTheme.border
                                        as OutlineInputBorder)
                                    .borderSide
                                    .color
                                : Theme.of(context).dividerColor,
                          ),
                        ),
                        contentPadding:
                            EdgeInsets.symmetric(horizontal: 16, vertical: 16),
                      ),
                      maxLines: 3,
                      onChanged: (value) {
                        setState(() {
                          _timeEntry!.description = value;
                        });
                      },
                    ),
                  ],
                )),
          ));
    }
  }

  bool _needToSetEndTime() {
    return (!_endTimeWasEmpty) && (_timeEntry!.endTime == null);
  }

  String _getTitle() {
    if (_timeEntryId == null) {
      return AppLocalizations.of(context)!.addTimeEntry;
    } else {
      return AppLocalizations.of(context)!.editTimeEntry;
    }
  }

  void _saveProject(FormState form) {
    form.save();
    if (_timeEntryId != null) {
      _timeEntryRepository.updateTimeEntry(_timeEntry!);
    } else {
      _timeEntryRepository.addTimeEntry(_timeEntry!);
    }
  }

  Future<ConfirmAction?> deleteTimeEntryWithRequest(
      BuildContext context) async {
    return showDialog<ConfirmAction>(
      context: context,
      barrierDismissible: false, // user must tap button for close dialog!
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(AppLocalizations.of(context)!.delete),
          content: Text(AppLocalizations.of(context)!.deleteTimeEntryRequest),
          actions: <Widget>[
            TextButton(
              child: Text(AppLocalizations.of(context)!.no),
              onPressed: () {
                Navigator.of(context).pop(ConfirmAction.CANCEL);
              },
            ),
            TextButton(
              child: Text(AppLocalizations.of(context)!.yes),
              onPressed: () {
                deleteTimeEntry();
                Navigator.of(context).pop(ConfirmAction.ACCEPT);
                Navigator.pop(context);
              },
            )
          ],
        );
      },
    );
  }

  void deleteTimeEntry() {
    if (_timeEntryId != null) {
      _timeEntryRepository.deleteTimeEntry(_timeEntry!);
    }
  }
}
