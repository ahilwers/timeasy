import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:timeasy/tools/duration_formatter.dart';

import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import 'package:timeasy/tools/weekly_statistics_builder.dart';
import 'package:timeasy/tools/weekly_statistics.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/tools/date_tools.dart';

class WeeklyStatisticsWidget extends StatefulWidget {
  final int _calendarWeek;
  final int _year;
  final Project _project;

  WeeklyStatisticsWidget(this._project, this._calendarWeek, this._year, {Key? key}) : super(key: key);

  @override
  _WeeklyStatisticsState createState() {
    return new _WeeklyStatisticsState(_project, _calendarWeek, _year);
  }
}

class _WeeklyStatisticsState extends State<WeeklyStatisticsWidget> {
  int _calendarWeek = 0;
  int _year = 0;
  final Project _project;
  WeeklyStatistics? _weeklyStatistics;
  // Need to define a page controller with a high initial page because otherwise
  // we could not swipe below the current week
  final _weeklyStatisticsBuilder = new WeeklyStatisticsBuilder();
  final DurationFormatter _durationFormatter = new DurationFormatter();

  _WeeklyStatisticsState(this._project, int calendarWeek, int year) {
    _calendarWeek = calendarWeek;
    _year = year;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(body: _weeklyStatistics == null ? _getData() : _buildLayout(context));
  }

  _getData() {
    _weeklyStatisticsBuilder.build(_project, _calendarWeek, _year).then((WeeklyStatistics statistics) {
      setState(() {
        _weeklyStatistics = statistics;
      });
    });
    Text("${AppLocalizations.of(context)!.loadingWeek} ${_calendarWeek.toString()}...");
  }

  _buildLayout(BuildContext context) {
    Locale locale = Localizations.localeOf(context);
    final formatter = new DateFormat.yMd(locale.toString());
    final dateTools = new DateTools();

    var startDate = formatter.format(dateTools.getFirstDayOfWeek(_calendarWeek, _year));
    var endDate = formatter.format(dateTools.getLastDayOfWeek(_calendarWeek, _year));

    return Card(
      child: ListView(children: <Widget>[
        ListTile(
          title: Text(AppLocalizations.of(context)!.weekTitle(_calendarWeek), style: TextStyle(fontWeight: FontWeight.w500)),
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
      ]),
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
        title: Text(AppLocalizations.of(context)!.weeklyHourSum, style: TextStyle(fontWeight: FontWeight.w500)),
        trailing: Text("${getSumAsString()}", style: TextStyle(fontWeight: FontWeight.w500)));
  }

  String _getNameOfDay(int weekday) {
    switch (weekday) {
      case 1:
        return AppLocalizations.of(context)!.monday;
      case 2:
        return AppLocalizations.of(context)!.tuesday;
      case 3:
        return AppLocalizations.of(context)!.wednesday;
      case 4:
        return AppLocalizations.of(context)!.thursday;
      case 5:
        return AppLocalizations.of(context)!.friday;
      case 6:
        return AppLocalizations.of(context)!.saturday;
      case 7:
        return AppLocalizations.of(context)!.sunday;
      default:
        return "";
    }
  }

  String _getHoursAsString(int weekday) {
    if (_weeklyStatistics == null) {
      return "";
    }
    var dayEntry = _weeklyStatistics?.getEntryForWeekDay(weekday);
    if ((dayEntry == null) || (dayEntry.seconds == 0)) {
      return "";
    }
    return _durationFormatter.formatDuration(new Duration(seconds: dayEntry.seconds));
  }

  String getSumAsString() {
    return _durationFormatter.formatDuration(new Duration(seconds: _weeklyStatistics?.getSumInSeconds() ?? 0));
  }
}
