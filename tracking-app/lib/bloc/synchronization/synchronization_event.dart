abstract class SynchronizationEvent {}

class SynchronizationSuccessEvent extends SynchronizationEvent {}

class SynchronizationErrorEvent extends SynchronizationEvent {
  final String message;

  SynchronizationErrorEvent(this.message);
}

class SynchonizationStartEvent extends SynchronizationEvent {}
