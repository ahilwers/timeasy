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

/// Emitted when the server responds with 402 Payment Required during sync,
/// i.e. the user has no active subscription for the hosted service.
class SynchronizationSubscriptionRequiredEvent extends SynchronizationEvent {}
