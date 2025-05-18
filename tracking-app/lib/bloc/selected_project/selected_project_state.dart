import 'package:timeasy/models/project.dart';

abstract class SelectedProjectState {
  final Project? project;

  SelectedProjectState(this.project);
}

class SelectedProjectInitial extends SelectedProjectState {
  SelectedProjectInitial() : super(null);
}

class SelectedProjectSet extends SelectedProjectState {
  SelectedProjectSet(Project project) : super(project);
}

class SelectedProjectCleared extends SelectedProjectState {
  SelectedProjectCleared() : super(null);
}
