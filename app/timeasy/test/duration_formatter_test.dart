import 'package:test/test.dart';
import 'package:timeasy/duration_formatter.dart';

void main() {
  test('Durations are formatted correctly.', () {
    final durationFormatter = new DurationFormatter();
    expect(durationFormatter.formatDuration(new Duration(seconds: 3600)), '01:00');
    expect(durationFormatter.formatDuration(new Duration(seconds: 5400)), '01:30');
    expect(durationFormatter.formatDuration(new Duration(seconds: 5520)), '01:32');
  });
}
