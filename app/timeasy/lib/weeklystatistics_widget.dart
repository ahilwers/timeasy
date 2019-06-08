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
  bool _dataAvailable = false;
  // Need to define a page controller with a high initial page because otherwise
  // we could not swipe below the current week
  final _weeklyStatisticsBuilder = new WeeklyStatisticsBuilder();

  _WeeklyStatisticsState(int calendarWeek) {
    _calendarWeek = calendarWeek;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        body: _dataAvailable ? _getData() : _buildLayout(context)
    );
  }

  _getData() {
    _weeklyStatisticsBuilder.build(_calendarWeek).then((WeeklyStatistics statistics) {
      setState(() {
        _dataAvailable = true;
      });
    });
    Text("Lade Woche ${_calendarWeek.toString()}...");
  }

  _buildLayout(BuildContext context) {
    _dataAvailable = false;
    Locale locale = Localizations.localeOf(context);
    var formatter = new DateFormat.yMd(locale.toString());

    var startDate = formatter.format(_weeklyStatisticsBuilder.getFirstDayOfWeek(_calendarWeek));
    var endDate = formatter.format(_weeklyStatisticsBuilder.getLastDayOfWeek(_calendarWeek));

    return Text("Woche "+_calendarWeek.toString()+" "+startDate+" - "+endDate);
  }

}