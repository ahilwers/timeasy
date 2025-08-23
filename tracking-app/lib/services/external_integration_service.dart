import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/external_issue.dart';

class ExternalIntegrationService {
  static final ExternalIntegrationService _instance = ExternalIntegrationService._internal();
  factory ExternalIntegrationService() => _instance;
  ExternalIntegrationService._internal();

  String? _baseUrl;
  String? _authToken;

  void initialize(String baseUrl, String authToken) {
    _baseUrl = baseUrl;
    _authToken = authToken;
  }

  Map<String, String> get _headers => {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer $_authToken',
  };

  /// Get description suggestions for autocomplete
  Future<DescriptionSuggestionsResponse> getDescriptionSuggestions(
    String projectId, 
    String query, 
    {int limit = 10}
  ) async {
    if (_baseUrl == null || _authToken == null) {
      throw Exception('Service not initialized');
    }

    final uri = Uri.parse('$_baseUrl/api/v1/projects/$projectId/descriptions/suggest')
        .replace(queryParameters: {
      'q': query,
      'limit': limit.toString(),
    });

    final response = await http.get(uri, headers: _headers);

    if (response.statusCode == 200) {
      final Map<String, dynamic> data = json.decode(response.body);
      return DescriptionSuggestionsResponse.fromMap(data);
    } else {
      throw Exception('Failed to get suggestions: ${response.statusCode}');
    }
  }

  /// Resolve an issue reference
  Future<IssueResolveResult> resolveIssue(String projectId, String input) async {
    if (_baseUrl == null || _authToken == null) {
      throw Exception('Service not initialized');
    }

    final uri = Uri.parse('$_baseUrl/api/v1/projects/$projectId/issues/resolve')
        .replace(queryParameters: {'input': input});

    final response = await http.get(uri, headers: _headers);

    if (response.statusCode == 200 || response.statusCode == 202) {
      final Map<String, dynamic> data = json.decode(response.body);
      return IssueResolveResult.fromMap(data);
    } else {
      throw Exception('Failed to resolve issue: ${response.statusCode}');
    }
  }

  /// Resolve pending external references
  Future<void> resolvePendingReferences() async {
    if (_baseUrl == null || _authToken == null) {
      throw Exception('Service not initialized');
    }

    final uri = Uri.parse('$_baseUrl/api/v1/external/resolve-pending');

    final response = await http.post(uri, headers: _headers);

    if (response.statusCode != 200) {
      throw Exception('Failed to resolve pending references: ${response.statusCode}');
    }
  }

  /// Detect if input contains issue patterns
  IssuePattern? detectIssuePattern(String input) {
    // GitHub/GitLab pattern: #123 or owner/repo#123
    final githubPattern = RegExp(r'(?:^|\s)(?:[\w.-]+/[\w.-]+)?#(\d+)(?:\s|$)');
    final githubMatch = githubPattern.firstMatch(input);
    if (githubMatch != null) {
      return IssuePattern(
        key: '#${githubMatch.group(1)}',
        provider: 'github', // Could also be gitlab
      );
    }

    // Jira pattern: ABC-123
    final jiraPattern = RegExp(r'(?:^|\s)([A-Z][A-Z0-9]+-\d+)(?:\s|$)');
    final jiraMatch = jiraPattern.firstMatch(input);
    if (jiraMatch != null) {
      return IssuePattern(
        key: jiraMatch.group(1)!,
        provider: 'jira',
      );
    }

    return null;
  }

  /// Get cached description suggestions for offline mode
  Future<List<String>> getCachedDescriptionSuggestions(String projectId, String query) async {
    // This would implement local caching logic
    // For now, return empty list when offline
    return [];
  }
}

class IssuePattern {
  final String key;
  final String provider;

  IssuePattern({required this.key, required this.provider});
}

/// Offline issue resolution service
class OfflineIssueResolutionService {
  static final OfflineIssueResolutionService _instance = OfflineIssueResolutionService._internal();
  factory OfflineIssueResolutionService() => _instance;
  OfflineIssueResolutionService._internal();

  final List<String> _pendingReferences = [];

  /// Add a pending external reference for later resolution
  void addPendingReference(String reference) {
    if (!_pendingReferences.contains(reference)) {
      _pendingReferences.add(reference);
    }
  }

  /// Get all pending references
  List<String> get pendingReferences => List.unmodifiable(_pendingReferences);

  /// Clear pending references (after successful sync)
  void clearPendingReferences() {
    _pendingReferences.clear();
  }

  /// Remove a specific pending reference
  void removePendingReference(String reference) {
    _pendingReferences.remove(reference);
  }
}