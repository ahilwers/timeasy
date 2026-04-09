class SubscriptionExpiredException implements Exception {
  final String message;
  const SubscriptionExpiredException([
    this.message = 'Your subscription is not active.',
  ]);

  @override
  String toString() => 'SubscriptionExpiredException: $message';
}