import 'package:timeasy/tools/retrieve_changes_result.dart';

abstract class SynchronizationState {}

class SynchronizationInitial extends SynchronizationState {}

class SynchronizationInProgress extends SynchronizationState {}

class SynchronizationSuccess extends SynchronizationState {
  final RetrieveChangesResult retrieveChangesResult;

  SynchronizationSuccess(this.retrieveChangesResult);
}

class SynchronizationError extends SynchronizationState {
  final String message;

  SynchronizationError(this.message);
}
