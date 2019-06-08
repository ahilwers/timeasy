import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:timeasy/weekly_statistics_builder.dart';
import 'package:timeasy/weekly_statistics.dart';


class WeeklyStatisticsWidget extends StatefulWidget {

  final String title;
  int _calendarWeek;

  WeeklyStatisticsWidget(int calendarWeek, {Key key, this.title}) : super(key: key) {
    _calendarWeek = calendarWeek;
  }

  @override
  _WeeklyStatisticsState createState() {
    return new _WeeklyStatisticsState(_calendarWeek);
  }

}

class _WeeklyStatisticsState extends State<WeeklyStatisticsWidget> {

  int _calendarWeek = 0;
  WeeklyStatistics _weeklyStatistics;
  // Need to define a page controller with a high initial page because otherwise
  // we could not swipe below the current week
  final _weeklyStatisticsBuilder = new WeeklyStatisticsBuilder();

  _WeeklyStatisticsState(int calendarWeek) {
    _calendarWeek = calendarWeek;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        body: _weeklyStatistics==null ? _getData() : _buildLayout(context)
    );
  }

  _getData() {
    _weeklyStatisticsBuilder.build(_calendarWeek).then((WeeklyStatistics statistics) {
      setState(() {
        _weeklyStatistics = statistics;
      });
    });
    Text("Lade Woche ${_calendarWeek.toString()}...");
  }

  _buildLayout(BuildContext context) {
    Locale locale = Localizations.localeOf(context);
    var formatter = new DateFormat.yMd(locale.toString());

    var startDate = formatter.format(_weeklyStatisticsBuilder.getFirstDayOfWeek(_calendarWeek));
    var endDate = formatter.format(_weeklyStatisticsBuilder.getLastDayOfWeek(_calendarWeek));

    return Column(
      children: <Widget>[
        Text("Woche "+_calendarWeek.toString()+" "+startDate+" - "+endDate),
        _buildDayEntry(1),
        _buildDayEntry(2),
        _buildDayEntry(3),
        _buildDayEntry(4),
        _buildDayEntry(5),
        _buildDayEntry(6),
        _buildDayEntry(7),
      ],
    );
  }

  _buildDayEntry(int weekday) {
    return Row(
      children: <Widget>[
        Text(_getNameOfDay(weekday)),
        Text(_getHours(weekday))
      ],
    );
  }

  String _getNameOfDay(int weekday) {
    switch (weekday) {
      case 1 : return "Montag";
      case 2 : return "Dienstag";
      case 3 : return "Mittwoch";
      case 4 : return "Donnerstag";
      case 5 : return "Freitag";
      case 6 : return "Samstag";
      case 7 : return "Sonntag";
    }
  }

  String _getHours(int weekday) {
    if (_weeklyStatistics==null) {
      return "";
    }
    var dayEntry = _weeklyStatistics.getEntryForWeekDay(weekday);
    if ((dayEntry==null) || (dayEntry.seconds==0)) {
      return "";
    }
    return (dayEntry.seconds/60/60).toString();
  }

}