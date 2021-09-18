import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:timeasy/weeklystatistics_widget.dart';
import 'package:timeasy/project.dart';
import 'package:timeasy/date_tools.dart';

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
  int _year = 0;
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
    final dateTools = new DateTools();
    _calendarWeek = dateTools.getWeekNumber(DateTime.now()) -
        1; //need to subtract one because the page is flipped forward once on startup.
    _year = DateTime.now().year;
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
                if (_calendarWeek > 52) {
                  _calendarWeek = 1;
                  _year++;
                }
              } else if ((position < _lastPosition) && (_calendarWeek > 0)) {
                _calendarWeek--;
                if (_calendarWeek < 1) {
                  _calendarWeek = 52;
                  _year--;
                }
              }
              _lastPosition = position;
              return new WeeklyStatisticsWidget(_project, _calendarWeek, _year);
            }));
  }

  String _getTitle() {
    return "Wochenübersicht (${_project.name})";
  }
}
