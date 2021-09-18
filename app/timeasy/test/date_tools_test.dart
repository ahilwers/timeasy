import 'package:test/test.dart';
import 'package:timeasy/date_tools.dart';

void main() {
  test('Week number is calculated correctly.', () {
    final dateTools = new DateTools();

    var weekNumber = dateTools.getWeekNumber(new DateTime(2021, 9, 14));
    expect(weekNumber, equals(37));

    weekNumber = dateTools.getWeekNumber(new DateTime(2020, 12, 28));
    expect(weekNumber, equals(53));

    weekNumber = dateTools.getWeekNumber(new DateTime(2021, 1, 4));
    expect(weekNumber, equals(1));
  });

  test('First day of week is calculated correctly', () {
    final dateTools = new DateTools();

    var firstDayOfWeek = dateTools.getFirstDayOfWeek(37, 2021);
    expect(firstDayOfWeek.year, 2021);
    expect(firstDayOfWeek.month, 9);
    expect(firstDayOfWeek.day, 13);

    firstDayOfWeek = dateTools.getFirstDayOfWeek(53, 2020);
    expect(firstDayOfWeek.year, 2021);
    expect(firstDayOfWeek.month, 12);
    expect(firstDayOfWeek.day, 28);

  })
}
