import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import 'package:timeasy/repositories/timeentry_repository.dart';
import 'package:timeasy/models/timeentry.dart';

enum ConfirmAction { CANCEL, ACCEPT }

class TimeEntryEditView extends StatelessWidget {
  final String? _timeEntryId;
  final String _projectId;

  TimeEntryEditView(this._projectId, [this._timeEntryId]) {}

  @override
  Widget build(BuildContext context) {
    return Scaffold(body: new TimeEntryEditWidget(_projectId, _timeEntryId));
  }
}

class TimeEntryEditWidget extends StatefulWidget {
  final String? _timeEntryId;
  final String _projectId;

  TimeEntryEditWidget(this._projectId, this._timeEntryId) {}

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

  _TimeEntryEditWidgetState(this._projectId, this._timeEntryId) {}

  @override
  void initState() {
    super.initState();
    if (_timeEntryId != null) {
      _timeEntryRepository.getTimeEntryById(_timeEntryId!).then((TimeEntry? timeEntryFromDb) {
        setState(() {
          _timeEntry = timeEntryFromDb;
          _endTimeWasEmpty = _timeEntry!.endTime == null; // Indicates that we're editing a time entry that is not completed yet
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
        appBar: new AppBar(
          title: new Text(AppLocalizations.of(context).loadingTimeEntry),
        ),
      );
    } else {
      Locale locale = Localizations.localeOf(context);
      var dateFormatter = new DateFormat.yMd(locale.toString());
      var timeFormatter = new DateFormat.Hm(locale.toString());
      return Scaffold(
          appBar: AppBar(
            title: Text(_getTitle()),
            actions: <Widget>[
              TextButton(
                onPressed: () {
                  final form = _formEditTimeEntryKey.currentState;
                  if (form!.validate()) {
                    // We need to validate the timeEntry separately
                    var errorMessage = "";
                    if (_timeEntry?.startTime == null) {
                      errorMessage = AppLocalizations.of(context).errorMissingStartTime;
                    } else if (_needToSetEndTime()) {
                      errorMessage = AppLocalizations.of(context).errorMissingEndTime;
                    } else if ((_timeEntry!.endTime != null) && (_timeEntry!.endTime!.isBefore(_timeEntry!.startTime))) {
                      errorMessage = AppLocalizations.of(context).errorEndtimeNotAfterStartTime;
                    }
                    if (errorMessage != "") {
                      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(errorMessage)));
                    } else {
                      _saveProject(form);
                      Navigator.pop(context);
                    }
                  }
                },
                child: Text(
                  AppLocalizations.of(context).save,
                  style: Theme.of(context).textTheme.subtitle1!.copyWith(color: Colors.white),
                ),
              ),
              _timeEntryId != null
                  ? TextButton(
                      onPressed: () {
                        deleteTimeEntryWithRequest(context);
                      },
                      child: Text(
                        AppLocalizations.of(context).delete,
                        style: Theme.of(context).textTheme.subtitle1!.copyWith(color: Colors.white),
                      ),
                    )
                  : Container(),
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
                        children: <Widget>[Text('${AppLocalizations.of(context).start}:', style: TextStyle(fontWeight: FontWeight.bold))]),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        TextButton(
                          onPressed: () {
                            _selectDate(context, _timeEntry!.startTime).then((DateTime? picked) {
                              if (picked != null) {
                                setState(() {
                                  var localStartTime = _timeEntry!.startTime.toLocal();
                                  _timeEntry!.startTime = new DateTime(picked.year, picked.month, picked.day, localStartTime.hour, localStartTime.minute).toUtc();
                                  // Also set the end time automatically if it's not already set:
                                  if (_needToSetEndTime()) {
                                    _timeEntry!.endTime = _timeEntry!.startTime;
                                  }
                                });
                              }
                            });
                          },
                          child: Text(dateFormatter.format(_timeEntry!.startTime.toLocal())),
                        ),
                        TextButton(
                          onPressed: () {
                            var startTime = TimeOfDay.fromDateTime(_timeEntry!.startTime.toLocal());
                            _selectTime(context, startTime).then((TimeOfDay? picked) {
                              if (picked != null) {
                                setState(() {
                                  var localStartTime = _timeEntry!.startTime.toLocal();
                                  _timeEntry!.startTime = new DateTime(localStartTime.year, localStartTime.month, localStartTime.day, picked.hour, picked.minute).toUtc();
                                  // Also set the end time automatically if it's not already set:
                                  if (_needToSetEndTime()) {
                                    _timeEntry!.endTime = _timeEntry!.startTime;
                                  }
                                });
                              }
                            });
                          },
                          child: Text(timeFormatter.format(_timeEntry!.startTime.toLocal())),
                        ),
                      ],
                    ),
                    Row(
                        crossAxisAlignment: CrossAxisAlignment.center,
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: <Widget>[Text('${AppLocalizations.of(context).end}:', style: TextStyle(fontWeight: FontWeight.bold))]),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        TextButton(
                          onPressed: () {
                            var endTime = _timeEntry!.endTime != null ? _timeEntry!.endTime : DateTime.now().toUtc();
                            _selectDate(context, endTime!).then((DateTime? picked) {
                              if (picked != null) {
                                setState(() {
                                  var localEndTime = _timeEntry!.endTime != null ? _timeEntry!.endTime?.toLocal() : DateTime.now();
                                  _timeEntry!.endTime = new DateTime(picked.year, picked.month, picked.day, localEndTime!.hour, localEndTime.minute).toUtc();
                                });
                              }
                            });
                          },
                          child: Text(_timeEntry!.endTime != null ? dateFormatter.format(_timeEntry!.endTime!.toLocal()) : AppLocalizations.of(context).endDate),
                        ),
                        TextButton(
                          onPressed: () {
                            var endTime = TimeOfDay.fromDateTime(_timeEntry!.endTime != null ? _timeEntry!.endTime!.toLocal() : DateTime.now());
                            _selectTime(context, endTime).then((TimeOfDay? picked) {
                              if (picked != null) {
                                setState(() {
                                  var localEndTime = _timeEntry!.endTime != null ? _timeEntry!.endTime!.toLocal() : DateTime.now();
                                  _timeEntry!.endTime = new DateTime(localEndTime.year, localEndTime.month, localEndTime.day, picked.hour, picked.minute).toUtc();
                                });
                              }
                            });
                          },
                          child: Text(_timeEntry!.endTime != null ? timeFormatter.format(_timeEntry!.endTime!.toLocal()) : AppLocalizations.of(context).endTime),
                        ),
                      ],
                    )
                  ],
                )),
          ));
    }
  }

  bool _needToSetEndTime() {
    return (!_endTimeWasEmpty) && (_timeEntry.endTime == null);
  }

  Future<DateTime?> _selectDate(BuildContext context, DateTime initialDate) async {
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: initialDate,
      firstDate: DateTime(2010),
      lastDate: DateTime(2201),
    );
    return picked;
  }

  Future<TimeOfDay?> _selectTime(BuildContext context, TimeOfDay initialSelectedTime) async {
    final TimeOfDay? picked = await showTimePicker(
      context: context,
      initialTime: initialSelectedTime,
    );
    return picked;
  }

  String _getTitle() {
    if (_timeEntryId == null) {
      return AppLocalizations.of(context).addTimeEntry;
    } else {
      return AppLocalizations.of(context).editTimeEntry;
    }
  }

  void _saveProject(FormState form) {
    form.save();
    if (_timeEntryId != null) {
      _timeEntryRepository.updateTimeEntry(_timeEntry);
    } else {
      _timeEntryRepository.addTimeEntry(_timeEntry);
    }
  }

  Future<ConfirmAction?> deleteTimeEntryWithRequest(BuildContext context) async {
    return showDialog<ConfirmAction>(
      context: context,
      barrierDismissible: false, // user must tap button for close dialog!
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(AppLocalizations.of(context).delete),
          content: Text(AppLocalizations.of(context).deleteTimeEntryRequest),
          actions: <Widget>[
            TextButton(
              child: Text(AppLocalizations.of(context).no),
              onPressed: () {
                Navigator.of(context).pop(ConfirmAction.CANCEL);
              },
            ),
            TextButton(
              child: Text(AppLocalizations.of(context).yes),
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
      _timeEntryRepository.deleteTimeEntry(_timeEntry);
    }
  }
}
