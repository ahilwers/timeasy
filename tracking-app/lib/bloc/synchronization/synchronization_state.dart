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

/// The server requires a paid subscription for sync (402 Payment Required).
class SynchronizationSubscriptionRequired extends SynchronizationState {}
