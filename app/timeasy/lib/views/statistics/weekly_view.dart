import 'package:flutter/material.dart';

import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import 'package:timeasy/views/statistics/weeklystatistics_widget.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/tools/date_tools.dart';

class WeeklyView extends StatelessWidget {
  final Project _project;

  WeeklyView(this._project) {}

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: WeeklyViewWidget(_project),
    );
  }
}

class WeeklyViewWidget extends StatefulWidget {
  final Project _project;

  WeeklyViewWidget(this._project, {Key key}) : super(key: key) {}

  @override
  _WeeklyViewState createState() {
    return new _WeeklyViewState(_project);
  }
}

class _WeeklyViewState extends State<WeeklyViewWidget> {
  final Project _project;
  int _calendarWeek = 0;
  int _year = 0;
  int _lastPosition = -1;
  final _dateTools = new DateTools();

  // Need to initialize the first page to such a high value to be able to swipe
  // backwards from the current week:
  final _pageController = new PageController(initialPage: 100000);

  _WeeklyViewState(this._project) {}

  @override
  void initState() {
    super.initState();
    _calendarWeek = _dateTools.getWeekNumber(DateTime.now()) - 1; //need to subtract one because the page is flipped forward once on startup.
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
                if (_calendarWeek > _dateTools.getNumberOfWeeks(_year)) {
                  _calendarWeek = 1;
                  _year++;
                }
              } else if ((position < _lastPosition) && (_calendarWeek > 0)) {
                _calendarWeek--;
                if (_calendarWeek < 1) {
                  _year--;
                  _calendarWeek = _dateTools.getNumberOfWeeks(_year);
                }
              }
              _lastPosition = position;
              return new WeeklyStatisticsWidget(_project, _calendarWeek, _year);
            }));
  }

  String _getTitle() {
    return "${AppLocalizations.of(context).weeklyOverview} (${_project.name})";
  }
}
