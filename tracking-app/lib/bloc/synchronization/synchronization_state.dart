abstract class SynchronizationState {}

class SynchronizationInitial extends SynchronizationState {}

class SynchronizationInProgress extends SynchronizationState {}

class SynchronizationSuccess extends SynchronizationState {}

class SynchronizationError extends SynchronizationState {
  final String message;

  SynchronizationError(this.message);
}
