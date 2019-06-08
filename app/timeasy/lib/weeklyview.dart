import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:timeasy/weeklystatistics_widget.dart';
import 'package:timeasy/project.dart';

class WeeklyView extends StatelessWidget {

  Project _project;

  WeeklyView(Project project) {
    _project = project;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: WeeklyViewWidget(_project),
    );
  }

}

class WeeklyViewWidget extends StatefulWidget {

  Project _project;

  WeeklyViewWidget(Project project, {Key key}) : super(key: key) {
    _project = project;
  }

  @override
  _WeeklyViewState createState() {
    return new _WeeklyViewState(_project);
  }

}

class _WeeklyViewState extends State<WeeklyViewWidget> {

  Project _project;
  int _calendarWeek = 0;
  int _lastPosition = -1;

  // Need to initialize the first page to such a high value to be able to swipe
  // backwards from the current week:
  final _pageController = new PageController(initialPage: 100000);

  _WeeklyViewState(Project project) {
    _project = project;
  }

  @override
  void initState() {
    super.initState();
    _calendarWeek = _weekNumber(DateTime.now())-1; //need to subtract one because the page is flipped forward once on startup.
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_getTitle()),
      ),
      body: PageView.builder(
        controller: _pageController,
        itemBuilder: (context, position) {
          if (position > _lastPosition) {
            _calendarWeek++;
          } else if ((position < _lastPosition) && (_calendarWeek > 1)) {
            _calendarWeek--;
          }
          _lastPosition = position;
          return new WeeklyStatisticsWidget(_project, _calendarWeek);
        }
      )
    );
  }

  /// Calculates week number from a date as per https://en.wikipedia.org/wiki/ISO_week_date#Calculation
  int _weekNumber(DateTime date) {
    int dayOfYear = int.parse(DateFormat("D").format(date));
    return ((dayOfYear - date.weekday + 10) / 7).floor();
  }

  String _getTitle() {
    return "Wochenübersicht (${_project.name})";
  }

}