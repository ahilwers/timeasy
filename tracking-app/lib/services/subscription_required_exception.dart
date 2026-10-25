/// Thrown when the server responds with 402 Payment Required, i.e. the user
/// does not have an active subscription for the hosted service.
class SubscriptionRequiredException implements Exception {
  final String message;

  SubscriptionRequiredException([this.message = 'Subscription required']);

  @override
  String toString() => message;
}
