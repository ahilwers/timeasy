import 'package:timeasy/models/project.dart';

abstract class SelectedProjectEvent {}

class SetSelectedProjectEvent extends SelectedProjectEvent {
  final Project? project;

  SetSelectedProjectEvent(this.project);
}

class ClearSelectedProjectEvent extends SelectedProjectEvent {}
