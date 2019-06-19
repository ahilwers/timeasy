import 'package:flutter/material.dart';

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
      return Scaffold(
        appBar: AppBar(
          title: Text(_getTitle()),
          actions: <Widget>[
            FlatButton(
              onPressed: () {
                final form = _formEditTimeEntryKey.currentState;
                if (form.validate()) {
                  _saveTimeEntry(form);
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
            child: TextFormField(
              decoration: InputDecoration(
                labelText: 'Beschreibung',
                border: OutlineInputBorder(),
              ),
              keyboardType: TextInputType.text,
              initialValue: _timeEntry.description,
              onSaved: (value) => _timeEntry.description = value,
            ),

          ),
        )
      );
    }
  }

  String _getTitle() {
    if (_timeEntryId==null) {
      return "Zeiteintrag hinzufügen";
    } else {
      return "Zeiteintrag bearbeiten";
    }
  }

  void _saveTimeEntry(FormState form) {
    form.save();
    if (_timeEntryId!=null) {
      _timeEntryRepository.updateTimeEntry(_timeEntry);
    } else {
      _timeEntryRepository.addTimeEntry(_timeEntry);
    }
  }



}