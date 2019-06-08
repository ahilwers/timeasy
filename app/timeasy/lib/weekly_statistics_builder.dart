import 'package:timeasy/weekly_statistics.dart';
import 'package:timeasy/timeentry_repository.dart';

class WeeklyStatisticsBuilder {

  final TimeEntryRepository _timeEntryRepository =  new TimeEntryRepository();

  Future<WeeklyStatistics> build(int weekNumber) async {
    var weeklyStatistics = new WeeklyStatistics();

    var startDate = getFirstDayOfWeek(weekNumber);
    var endDate = getLastDayOfWeek(weekNumber);
    var timeEntries = await _timeEntryRepository.getTimeEntries(startDate, endDate);

    var lastDay = 0;
    WeeklyStatisticsEntry statisticsEntry;
    for (var timeEntry in timeEntries) {
      var currentDay = timeEntry.startTime.day;
      // A new day has started and a new statistics entry must be created:
      if ((currentDay>lastDay) || (statisticsEntry==null)) {
        statisticsEntry = new WeeklyStatisticsEntry();
        statisticsEntry.date = timeEntry.startTime;
        weeklyStatistics.addEntryForWeekDay(timeEntry.startTime.weekday, statisticsEntry);
      }
      statisticsEntry.seconds+=timeEntry.getSeconds();
    }
    return weeklyStatistics;
  }

  DateTime getFirstDayOfWeek(int weekNumber) {
    var daysInYear = (weekNumber-1)*7;
    var firstDayOfYear = new DateTime(2019, 1, 1);
    return firstDayOfYear.add(new Duration(days: daysInYear-1));
  }

  DateTime getLastDayOfWeek(int weekNumber) {
    var firstDayOfWeek = getFirstDayOfWeek(weekNumber);
    return firstDayOfWeek.add(new Duration(days: 6));
  }

}