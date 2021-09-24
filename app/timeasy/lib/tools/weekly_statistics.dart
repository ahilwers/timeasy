class WeeklyStatistics {
  Map<int, WeeklyStatisticsEntry> _entries = new Map();

  WeeklyStatisticsEntry? getEntryForWeekDay(int weekDay) {
    return _entries[weekDay];
  }

  void addEntryForWeekDay(int weekDay, WeeklyStatisticsEntry entry) {
    _entries[weekDay] = entry;
  }

  int getSumInSeconds() {
    var _sum = 0;
    _entries.forEach((weekday, entry) => _sum += entry.seconds);
    return _sum;
  }
}

class WeeklyStatisticsEntry {
  final DateTime date;
  int seconds = 0;

  WeeklyStatisticsEntry(this.date) {}

  double getMinutes() {
    return seconds / 60;
  }

  double getHours() {
    return getMinutes() / 60;
  }
}
