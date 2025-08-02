import 'package:timeasy/tools/retrieve_changes_result.dart';

abstract class SynchronizationEvent {}

class SynchronizationSuccessEvent extends SynchronizationEvent {
  final RetrieveChangesResult retrieveChangesResult;

  SynchronizationSuccessEvent(this.retrieveChangesResult);
}

class SynchronizationErrorEvent extends SynchronizationEvent {
  final String message;

  SynchronizationErrorEvent(this.message);
}

class SynchonizationStartEvent extends SynchronizationEvent {}
