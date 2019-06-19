import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:timeasy/timeentry_repository.dart';
import 'package:timeasy/timeentry.dart';

class TimeEntryEditView extends StatelessWidget {

  String _timeEntryId;
  String _projectId;

  TimeEntryEditView(String projectId, {String timeEntryId}) {
    _timeEntryId = timeEntryId;
    _projectId = projectId;
  }

  @override
  Widget build(BuildContext context) {

    return Scaffold(
        body: new TimeEntryEditWidget(_projectId, _timeEntryId)
    );
  }

}

class TimeEntryEditWidget extends StatefulWidget {

  String _timeEntryId;
  String _projectId;

  TimeEntryEditWidget(String projectId, String timeEntryId) {
    _timeEntryId = timeEntryId;
    _projectId = projectId;
  }

  @override
  _TimeEntryEditWidgetState createState() {
    return new _TimeEntryEditWidgetState(_projectId, _timeEntryId);
  }

}

class _TimeEntryEditWidgetState extends State<TimeEntryEditWidget> {

  String _timeEntryId;
  String _projectId;
  TimeEntry _timeEntry;
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final _formEditTimeEntryKey = GlobalKey<FormState>();

  _TimeEntryEditWidgetState(String projectId, String timeEntryId) {
    _timeEntryId = timeEntryId;
    _projectId = projectId;
  }

  @override
  void initState() {
    super.initState();
    if (_timeEntryId!=null) {
      _timeEntryRepository.getTimeEntryById(_timeEntryId).then((TimeEntry timeEntryFromDb) {
        setState(() {
          _timeEntry = timeEntryFromDb;
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
          title: new Text("Lade Zeiteintrag..."),
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
            FlatButton(
              onPressed: () {
                final form = _formEditTimeEntryKey.currentState;
                if (form.validate()) {
                  _saveProject(form);
                  Navigator.pop(context);
                }
              },
              child: Text("Speichern",
                style: Theme.of(context)
                  .textTheme
                  .subhead
                  .copyWith(color: Colors.white),
              ),
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
                TextFormField(
                  decoration: InputDecoration(
                    labelText: 'Beschreibung',
                    border: OutlineInputBorder(),
                  ),
                  keyboardType: TextInputType.text,
                  initialValue: _timeEntry.description,
                  onSaved: (value) => _timeEntry.description = value,
                ),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    FlatButton(
                      onPressed: () {
                        _selectDate(context, _timeEntry.startTime).then((DateTime picked) {
                          if (picked != null) {
                            setState(() {
                              var localStartTime = _timeEntry.startTime.toLocal();
                              _timeEntry.startTime = new DateTime(picked.year, picked.month, picked.day, localStartTime.hour, localStartTime.minute).toUtc();
                            });
                          }
                        });

                      },
                      child: Text(dateFormatter.format(_timeEntry.startTime.toLocal())),

                    ),
                    FlatButton(
                      onPressed: () {
                        var startTime = TimeOfDay.fromDateTime(_timeEntry.startTime.toLocal());
                        _selectTime(context, startTime).then((TimeOfDay picked) {
                          if (picked!=null) {
                            setState(() {
                              var localStartTime = _timeEntry.startTime.toLocal();
                              _timeEntry.startTime = new DateTime(localStartTime.year, localStartTime.month, localStartTime.day, picked.hour, picked.minute).toUtc();
                            });
                          }
                        });

                      },
                      child: Text(timeFormatter.format(_timeEntry.startTime.toLocal())),

                    ),
                  ],
                ),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    FlatButton(
                      onPressed: () {
                        var endTime = _timeEntry.endTime!=null ? _timeEntry.endTime : DateTime.now().toUtc();
                        _selectDate(context, endTime).then((DateTime picked) {
                          if (picked != null) {
                            setState(() {
                              var localEndTime = _timeEntry.endTime!=null ? _timeEntry.endTime.toLocal() : DateTime.now();
                              _timeEntry.endTime = new DateTime(picked.year, picked.month, picked.day, localEndTime.hour, localEndTime.minute).toUtc();
                            });
                          }
                        });
                      },
                      child: Text(_timeEntry.endTime != null ? dateFormatter.format(_timeEntry.endTime.toLocal()) : dateFormatter.format(DateTime.now())),

                    ),
                    FlatButton(
                      onPressed: () {
                        var endTime = TimeOfDay.fromDateTime(_timeEntry.endTime!=null ? _timeEntry.endTime.toLocal() : DateTime.now());
                        _selectTime(context, endTime).then((TimeOfDay picked) {
                          if (picked!=null) {
                            setState(() {
                              var localEndTime = _timeEntry.endTime!=null ? _timeEntry.endTime.toLocal() : DateTime.now();
                              _timeEntry.endTime = new DateTime(localEndTime.year, localEndTime.month, localEndTime.day, picked.hour, picked.minute).toUtc();
                            });
                          }
                        });

                      },
                      child: Text(_timeEntry.endTime!=null ? timeFormatter.format(_timeEntry.endTime.toLocal()) : dateFormatter.format(DateTime.now())),

                    ),
                  ],
                )

              ],
            )
          ),
        )
      );
    }
  }

  Future<DateTime> _selectDate(BuildContext context, DateTime initialDate) async {
    final DateTime picked = await showDatePicker(
      context: context,
      initialDate: initialDate,
      firstDate: DateTime(2010),
      lastDate: DateTime(2201),
    );
    return picked;
  }

  Future<TimeOfDay> _selectTime(BuildContext context, TimeOfDay initialSelectedTime) async {
    final TimeOfDay picked = await showTimePicker(
      context: context,
      initialTime: initialSelectedTime,
    );
    return picked;
  }



  String _getTitle() {
    if (_timeEntryId==null) {
      return "Zeiteintrag hinzufügen";
    } else {
      return "Zeiteintrag bearbeiten";
    }
  }

  void _saveProject(FormState form) {
    form.save();
    if (_timeEntryId!=null) {
      _timeEntryRepository.updateTimeEntry(_timeEntry);
    } else {
      _timeEntryRepository.addTimeEntry(_timeEntry);
    }
  }



}