import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:timeasy/bloc/selected_project/selected_project_event.dart';
import 'package:timeasy/bloc/selected_project/selected_project_state.dart';

class SelectedProjectBloc
    extends Bloc<SelectedProjectEvent, SelectedProjectState> {
  SelectedProjectBloc() : super(SelectedProjectInitial()) {
    on<SetSelectedProjectEvent>((event, emit) {
      emit(SelectedProjectSet(event.project!));
    });
    on<ClearSelectedProjectEvent>((event, emit) {
      emit(SelectedProjectCleared());
    });
  }
}
