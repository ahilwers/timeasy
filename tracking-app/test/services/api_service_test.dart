import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:test/test.dart';
import 'package:timeasy/services/api_service.dart';
import 'package:timeasy/services/rate_limited_exception.dart';
import 'package:timeasy/services/subscription_required_exception.dart';

void main() {
  test('429 responses throw RateLimitedException with the Retry-After value',
      () async {
    final client = MockClient((request) async {
      return http.Response('{"error":"rate_limited"}', 429,
          headers: {'retry-after': '30'});
    });
    final api =
        ApiService(baseUrl: 'https://example.com', token: 't', client: client);

    await expectLater(
      api.get('/sync/changed/0'),
      throwsA(isA<RateLimitedException>()
          .having((e) => e.retryAfterSeconds, 'retryAfterSeconds', 30)),
    );
  });

  test('429 responses without Retry-After still throw RateLimitedException',
      () async {
    final client = MockClient((request) async {
      return http.Response('{"error":"rate_limited"}', 429);
    });
    final api = ApiService(baseUrl: 'https://example.com', client: client);

    await expectLater(
      api.post('/sync/changed'),
      throwsA(isA<RateLimitedException>()
          .having((e) => e.retryAfterSeconds, 'retryAfterSeconds', isNull)),
    );
  });

  test('402 responses still throw SubscriptionRequiredException', () async {
    final client = MockClient((request) async {
      return http.Response('{"error":"subscription_inactive"}', 402);
    });
    final api = ApiService(baseUrl: 'https://example.com', client: client);

    await expectLater(
      api.get('/projects'),
      throwsA(isA<SubscriptionRequiredException>()),
    );
  });

  test('successful responses are returned unchanged', () async {
    final client = MockClient((request) async {
      return http.Response('{"ok":true}', 200);
    });
    final api = ApiService(baseUrl: 'https://example.com', client: client);

    final response = await api.get('/sync/changed/0');
    expect(response.statusCode, 200);
    expect(response.body, '{"ok":true}');
  });
}
