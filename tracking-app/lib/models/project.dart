import 'package:uuid/uuid.dart';

class Project {
  static final String tableName = "Projects";
  static final String idColumn = "id";
  static final String nameColumn = "name";
  static final String createdColumn = "created";
  static final String updatedColumn = "updated";
  static final String deletedColumn = "deleted";
  static final String colorColumn = "color";
  static final String isActiveColumn = "isActive";

  late String id;
  String name = "";
  String color = "#1E90FF"; // Default blue color
  DateTime created = DateTime.now().toUtc();
  DateTime updated = DateTime.now().toUtc();
  bool deleted = false;
  bool isActive = true;

  Project() {
    var uuid = new Uuid();
    id = uuid.v4();
  }

  Project.fromMap(Map<String, dynamic> map) {
    id = map[idColumn];
    name = map[nameColumn];
    color = map[colorColumn] ?? "#1E90FF"; // Default to blue if not set
    int createdMillis = map[createdColumn];
    created =
        new DateTime.fromMillisecondsSinceEpoch(createdMillis, isUtc: true);
    int updatedMillis = map[updatedColumn];
    updated =
        new DateTime.fromMillisecondsSinceEpoch(updatedMillis, isUtc: true);
    int deletedInt = map[deletedColumn];
    deletedInt == 0 ? deleted = false : deleted = true;
    int isActiveInt = map[isActiveColumn] ?? 1;
    isActiveInt == 0 ? isActive = false : isActive = true;
  }

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      idColumn: id,
      nameColumn: name,
      colorColumn: color,
      createdColumn: created.millisecondsSinceEpoch,
      updatedColumn: updated.millisecondsSinceEpoch,
      deletedColumn: deleted ? 1 : 0,
      isActiveColumn: isActive ? 1 : 0,
    };
  }
}
