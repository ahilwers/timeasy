class WeeklyStatistics {

  Map<int, WeeklyStatisticsEntry> _entries;

  WeeklyStatisticsEntry getEntryForWeekDay(int weekDay) {
    return _entries[weekDay];
  }

  void addEntryForWeekDay(int weekDay, WeeklyStatisticsEntry entry) {
    _entries[weekDay] = entry;
  }

}

class WeeklyStatisticsEntry {
  DateTime date;
  int seconds;

  double getMinutes() {
    return seconds/60;
  }

  double getHours() {
    return getMinutes()/60;
  }
}