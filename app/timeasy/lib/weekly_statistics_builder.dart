import 'package:intl/intl.dart';
import 'package:timeasy/weekly_statistics.dart';
import 'package:timeasy/timeentry_repository.dart';
import 'package:timeasy/timeentry.dart';
import 'package:timeasy/project.dart';

class WeeklyStatisticsBuilder {

  final TimeEntryRepository _timeEntryRepository =  new TimeEntryRepository();

  Future<WeeklyStatistics> build(Project project, int weekNumber, int year) async {
    var weeklyStatistics = new WeeklyStatistics();

    var startDate = getFirstDayOfWeek(weekNumber, year);
    var endDate = getLastDayOfWeek(weekNumber, year);
    var timeEntries = await _timeEntryRepository.getTimeEntries(project.id, startDate, endDate);

    var lastDay = 0;
    for (var timeEntry in timeEntries) {
      if (!_isTimeEntryValid(timeEntry, endDate))
        continue;
      var currentDay = timeEntry.startTime.day;
      // A new day has started and a new statistics entry must be created:
      var statisticsEntry = weeklyStatistics.getEntryForWeekDay(timeEntry.startTime.weekday);
      if ((currentDay>lastDay) || (statisticsEntry==null)) {
        statisticsEntry = new WeeklyStatisticsEntry();
        statisticsEntry.date = timeEntry.startTime;
        weeklyStatistics.addEntryForWeekDay(timeEntry.startTime.weekday, statisticsEntry);
      }
      statisticsEntry.seconds+=timeEntry.getSeconds();
      lastDay = currentDay;
    }
    return weeklyStatistics;
  }

  bool _isTimeEntryValid(TimeEntry timeEntry, DateTime endDate) {
    // The query may return time entries without end date although they don't
    // fit into this time period. We must check this here:
    var nextDate = endDate.add(Duration(days: 1));
    return timeEntry.startTime.isBefore(nextDate);
  }

  DateTime getFirstDayOfWeek(int weekNumber, int year) {
    var daysInYear = (weekNumber-1)*7;
    var firstDayOfYear = new DateTime(year, 1, 1);
    return firstDayOfYear.add(new Duration(days: daysInYear-1));
  }

  DateTime getLastDayOfWeek(int weekNumber, int year) {
    var firstDayOfWeek = getFirstDayOfWeek(weekNumber, year);
    return firstDayOfWeek.add(new Duration(days: 6));
  }

}