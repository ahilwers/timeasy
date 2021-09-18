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
}
