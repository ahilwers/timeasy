/// Thrown when the server answers with 429 Too Many Requests.
///
/// Rate limiting is transient: callers should retry later instead of treating
/// the request as failed. [retryAfterSeconds] carries the server's Retry-After
/// hint when it sent one.
class RateLimitedException implements Exception {
  final int? retryAfterSeconds;

  RateLimitedException([this.retryAfterSeconds]);

  @override
  String toString() {
    final retryAfter = retryAfterSeconds;
    return retryAfter == null
        ? 'RateLimitedException'
        : 'RateLimitedException(retryAfter: ${retryAfter}s)';
  }
}
