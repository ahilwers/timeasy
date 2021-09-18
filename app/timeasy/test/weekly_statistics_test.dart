import 'dart:math';

import 'package:test/test.dart';
import 'package:timeasy/weekly_statistics.dart';

void main() {
  test('Duration is calculated correctly.', () {
    final weeklyStatistics = new WeeklyStatistics();

    var entry1 = new WeeklyStatisticsEntry();
    entry1.date = new DateTime(2021, 9, 18);
    entry1.seconds = 1800;
    weeklyStatistics.addEntryForWeekDay(6, entry1);

    var entry2 = new WeeklyStatisticsEntry();
    entry2.date = entry1.date;
    entry2.seconds = 1800;
    weeklyStatistics.addEntryForWeekDay(7, entry2);

    expect(weeklyStatistics.getSumInSeconds(), 3600);
    expect(weeklyStatistics.getDuration().inSeconds, 3600);
  });
}
