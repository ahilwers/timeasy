import 'package:intl/intl.dart';

class DateTools {
  /// Returns week number for a date (https://en.wikipedia.org/wiki/ISO_week_date#Calculation)
  int getWeekNumber(DateTime date) {
    int dayOfYear = int.parse(DateFormat("D").format(date));
    int weekNumber = ((dayOfYear - date.weekday + 10) / 7).floor();
    if (weekNumber < 1) {
      weekNumber = getNumberOfWeeks(date.year - 1);
    } else if (weekNumber > getNumberOfWeeks(date.year)) {
      weekNumber = 1;
    }
    return weekNumber;
  }

  /// Returns number of weeks in a given year (https://en.wikipedia.org/wiki/ISO_week_date#Weeks_per_year)
  int getNumberOfWeeks(int year) {
    final december28 = DateTime(year, 12, 28);
    int dayOfDecember28 = int.parse(DateFormat("D").format(december28));
    return ((dayOfDecember28 - december28.weekday + 10) / 7).floor();
  }

  DateTime getFirstDayOfWeek(int weekNumber, int year) {
    var daysInYear = (weekNumber - 1) * 7;
    var firstDayOfFirstWeek = getFirstDayOfFirstWeek(year);
    return firstDayOfFirstWeek.add(new Duration(days: daysInYear));
  }

  DateTime getFirstDayOfFirstWeek(int year) {
    var firstDay = new DateTime(year, 1, 1);
    final numberOfWeeksInLastYear = getNumberOfWeeks(year - 1);
    if (numberOfWeeksInLastYear == 53) {
      final december28 = new DateTime(year - 1, 12, 28);
      // As devember 28th is always in the last wekk of the last year, a day one week later must be in the first week of the next year:
      firstDay = december28.add(new Duration(days: 7));
    }

    // If this day is not a monday, the first day of the week must be in the last year:
    if (firstDay.weekday != 1) {
      firstDay = firstDay.subtract(new Duration(days: firstDay.weekday - 1));
    }
    return firstDay;
  }

  DateTime getLastDayOfWeek(int weekNumber, int year) {
    var firstDayOfWeek = getFirstDayOfWeek(weekNumber, year);
    return firstDayOfWeek.add(new Duration(days: 6));
  }
}
