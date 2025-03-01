enum ChangeType { NEW, CHANGED, DELETED }

class ChangeTypeHelper {
  static ChangeType convertFromString(String value) {
    return ChangeType.values
        .firstWhere((e) => e.toString().split('.').last == value);
  }

  static String convertToString(ChangeType changeType) {
    return changeType.toString().split('.').last;
  }
}
