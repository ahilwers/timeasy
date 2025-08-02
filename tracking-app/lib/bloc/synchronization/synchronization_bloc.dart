import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:timeasy/bloc/synchronization/synchronization_event.dart';
import 'package:timeasy/bloc/synchronization/synchronization_state.dart';

class SynchronizationBloc
    extends Bloc<SynchronizationEvent, SynchronizationState> {
  SynchronizationBloc() : super(SynchronizationInitial()) {
    on<SynchonizationStartEvent>(
        (event, emit) => emit(SynchronizationInProgress()));
    on<SynchronizationSuccessEvent>((event, emit) =>
        emit(SynchronizationSuccess(event.retrieveChangesResult)));
    on<SynchronizationErrorEvent>(
        (event, emit) => emit(SynchronizationError(event.message)));
  }
}
