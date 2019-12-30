import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:timeasy/weekly_statistics_builder.dart';
import 'package:timeasy/weekly_statistics.dart';
import 'package:timeasy/project.dart';


class WeeklyStatisticsWidget extends StatefulWidget {

  int _calendarWeek;
  int _year;
  Project _project;

  WeeklyStatisticsWidget(Project project, int calendarWeek, int year, {Key key}) : super(key: key) {
    _calendarWeek = calendarWeek;
    _year = year;
    _project = project;
  }

  @override
  _WeeklyStatisticsState createState() {
    return new _WeeklyStatisticsState(_project, _calendarWeek, _year);
  }

}

class _WeeklyStatisticsState extends State<WeeklyStatisticsWidget> {

  int _calendarWeek = 0;
  int _year = 0;
  Project _project;
  WeeklyStatistics _weeklyStatistics;
  // Need to define a page controller with a high initial page because otherwise
  // we could not swipe below the current week
  final _weeklyStatisticsBuilder = new WeeklyStatisticsBuilder();

  _WeeklyStatisticsState(Project project, int calendarWeek, int year) {
    _calendarWeek = calendarWeek;
    _year = year;
    _project = project;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        body: _weeklyStatistics==null ? _getData() : _buildLayout(context)
    );
  }

  _getData() {
    _weeklyStatisticsBuilder.build(_project, _calendarWeek, _year).then((WeeklyStatistics statistics) {
      setState(() {
        _weeklyStatistics = statistics;
      });
    });
    Text("Lade Woche ${_calendarWeek.toString()}...");
  }

  _buildLayout(BuildContext context) {
    Locale locale = Localizations.localeOf(context);
    var formatter = new DateFormat.yMd(locale.toString());

    var startDate = formatter.format(_weeklyStatisticsBuilder.getFirstDayOfWeek(_calendarWeek, _year));
    var endDate = formatter.format(_weeklyStatisticsBuilder.getLastDayOfWeek(_calendarWeek, _year));

    return Card (
      child: ListView(
        children: <Widget>[
          ListTile(
            title: Text("${_calendarWeek.toString()}. Kalenderwoche", style: TextStyle(fontWeight: FontWeight.w500)),
            subtitle: Text("$startDate - $endDate"),
          ),
          Divider(),
          _buildDayEntry(1),
          _buildDayEntry(2),
          _buildDayEntry(3),
          _buildDayEntry(4),
          _buildDayEntry(5),
          _buildDayEntry(6),
          _buildDayEntry(7),
          Divider(),
          _buildSumEntry()
        ]
      ),
    );
  }

  _buildDayEntry(int weekday) {
    return ListTile(
      title: Text(_getNameOfDay(weekday)),
      trailing: Text(_getHoursAsString(weekday)),
    );
  }

  _buildSumEntry() {
    return ListTile(
      title: Text("Summe:", style: TextStyle(fontWeight: FontWeight.w500)),
      trailing: Text("${getSumAsString()} Stunden", style: TextStyle(fontWeight: FontWeight.w500))
    );
  }

  String _getNameOfDay(int weekday) {
    switch (weekday) {
      case 1 :
        return "Montag";
      case 2 :
        return "Dienstag";
      case 3 :
        return "Mittwoch";
      case 4 :
        return "Donnerstag";
      case 5 :
        return "Freitag";
      case 6 :
        return "Samstag";
      case 7 :
        return "Sonntag";
      default:
        return "";
    }
  }

  String _getHoursAsString(int weekday) {
    if (_weeklyStatistics==null) {
      return "";
    }
    var dayEntry = _weeklyStatistics.getEntryForWeekDay(weekday);
    if ((dayEntry==null) || (dayEntry.seconds==0)) {
      return "";
    }
    return (dayEntry.seconds/60/60).toStringAsFixed(2);
  }

  String getSumAsString() {
    return (_weeklyStatistics.getSumInSeconds()/60/60).toStringAsFixed(2);
  }

}