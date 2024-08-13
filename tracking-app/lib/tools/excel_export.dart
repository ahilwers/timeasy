abstract class ExcelExport {
  Map<String, String> _translations = new Map();

  ExcelExport() {
    addTranslation("date", "Date");
    addTranslation("start", "Start");
    addTranslation("end", "End");
    addTranslation("pause", "Pause");
  }

  Future<void> Export();

  void addTranslation(String key, String value) {
    _translations[key] = value;
  }

  String getTranslation(String key) {
    return _translations[key] ?? key;
  }
}
