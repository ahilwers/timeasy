import 'package:intl/intl.dart';

class DateTools {
  /// Calculates week number from a date as per https://en.wikipedia.org/wiki/ISO_week_date#Calculation
  int getWeekNumber(DateTime date) {
    int dayOfYear = int.parse(DateFormat("D").format(date));
    return ((dayOfYear - date.weekday + 10) / 7).floor();
  }

  DateTime getFirstDayOfWeek(int weekNumber, int year) {
    var daysInYear = (weekNumber) * 7;
    var firstDayOfFirstWeek = getFirstDayOfFirstWeek(year);
    return firstDayOfFirstWeek.add(new Duration(days: daysInYear));
  }

  DateTime getFirstDayOfFirstWeek(int year) {
    var firstDay = new DateTime(year, 1, 1);
    // If this day is not a monday, the first day of the week must be in the last year:
    if (firstDay.weekday != 0) {
      firstDay = firstDay.subtract(new Duration(days: firstDay.weekday - 1));
    }
    return firstDay;
  }

  DateTime getLastDayOfWeek(int weekNumber, int year) {
    var firstDayOfWeek = getFirstDayOfWeek(weekNumber, year);
    return firstDayOfWeek.add(new Duration(days: 6));
  }
}
