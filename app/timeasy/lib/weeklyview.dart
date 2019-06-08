import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

class WeeklyView extends StatelessWidget {

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: WeeklyViewWidget(title: "timeasy - Wochenübersicht"),
    );
  }

}

class WeeklyViewWidget extends StatefulWidget {

  final String title;

  WeeklyViewWidget({Key key, this.title}) : super(key: key);

  @override
  _WeeklyViewState createState() {
    return new _WeeklyViewState();
  }

}

class _WeeklyViewState extends State<WeeklyViewWidget> {

  int _calendarWeek = 0;
  int _lastPosition = -1;
  // Need to define a page controller with a high initial page because otherwise
  // we could not swipe below the current week
  final _pageController = new PageController(initialPage: 10000);

  @override
  void initState() {
    super.initState();
    _calendarWeek = _weekNumber(DateTime.now())-1; //need to subtract one because the page is flipped forward once on startup.
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text("timeasy - Wochenübersicht"),
      ),
      body: PageView.builder(
        controller: _pageController,
        itemBuilder: (context, position) {
          var oldCalendarWeek = _calendarWeek;
          if (position>_lastPosition) {
            _calendarWeek++;
          } else if (_calendarWeek>1) {
            _calendarWeek--;
          }
          _lastPosition = position;
          if (_calendarWeek!=oldCalendarWeek) {
            return _buildLayout(context);
          }
        }
      )
    );
  }

  _buildLayout(BuildContext context) {

    Locale locale = Localizations.localeOf(context);
    var formatter = new DateFormat.yMd(locale.toString());


    var startDate = formatter.format(_getFirstDayOfWeek(_calendarWeek));
    var endDate = formatter.format(_getLastDayOfWeek(_calendarWeek));

    return Text("Woche "+_calendarWeek.toString()+" "+startDate+" - "+endDate);
  }

  /// Calculates week number from a date as per https://en.wikipedia.org/wiki/ISO_week_date#Calculation
  int _weekNumber(DateTime date) {
    int dayOfYear = int.parse(DateFormat("D").format(date));
    return ((dayOfYear - date.weekday + 10) / 7).floor();
  }

  DateTime _getFirstDayOfWeek(int weekNumber) {
    var daysInYear = (weekNumber-1)*7;
    var firstDayOfYear = new DateTime(2019, 1, 1);
    return firstDayOfYear.add(new Duration(days: daysInYear-1));
  }

  DateTime _getLastDayOfWeek(int weekNumber) {
    var firstDayOfWeek = _getFirstDayOfWeek(weekNumber);
    return firstDayOfWeek.add(new Duration(days: 6));
  }

}