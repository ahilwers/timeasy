import 'package:uuid/uuid.dart';

class Project {
  static final String tableName = "Projects";
  static final String idColumn = "id";
  static final String nameColumn = "name";
  static final String createdColumn = "created";
  static final String updatedColumn = "updated";
  static final String deletedColumn = "deleted";

  String id;
  String name = "";
  DateTime created = DateTime.now().toUtc();
  DateTime updated = DateTime.now().toUtc();
  bool deleted = false;

  Project() {
    var uuid = new Uuid();
    id = uuid.v1();
  }

  Project.fromMap(Map<String, dynamic> map) {
    id = map[idColumn];
    name = map[nameColumn];
    int createdMillis = map[createdColumn];
    created = new DateTime.fromMillisecondsSinceEpoch(createdMillis, isUtc: true);
    int updatedMillis = map[updatedColumn];
    updated = new DateTime.fromMillisecondsSinceEpoch(updatedMillis, isUtc: true);
    int deletedInt = map[deletedColumn];
    deletedInt == 0 ? deleted = false : deleted = true;
  }

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      idColumn: id,
      nameColumn: name,
      createdColumn: created.millisecondsSinceEpoch,
      updatedColumn: updated.millisecondsSinceEpoch,
      deletedColumn: deleted ? 1 : 0,
    };
  }
}
