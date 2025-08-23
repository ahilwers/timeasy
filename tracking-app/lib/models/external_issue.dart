class ExternalIssue {
  final String id;
  final String projectId;
  final String provider; // 'github', 'gitlab', 'jira'
  final String key;
  final String title;
  final String state;
  final String? url;
  final DateTime createdAt;
  final DateTime updatedAt;

  ExternalIssue({
    required this.id,
    required this.projectId,
    required this.provider,
    required this.key,
    required this.title,
    required this.state,
    this.url,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ExternalIssue.fromMap(Map<String, dynamic> map) {
    return ExternalIssue(
      id: map['id'],
      projectId: map['projectId'],
      provider: map['provider'],
      key: map['key'],
      title: map['title'],
      state: map['state'],
      url: map['url'],
      createdAt: DateTime.parse(map['createdAt']),
      updatedAt: DateTime.parse(map['updatedAt']),
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'projectId': projectId,
      'provider': provider,
      'key': key,
      'title': title,
      'state': state,
      'url': url,
      'createdAt': createdAt.toIso8601String(),
      'updatedAt': updatedAt.toIso8601String(),
    };
  }
}

class IssueResolveResult {
  final String? issueId;
  final String key;
  final String? title;
  final String? state;
  final String? url;
  final String status; // 'resolved' or 'pending'

  IssueResolveResult({
    this.issueId,
    required this.key,
    this.title,
    this.state,
    this.url,
    required this.status,
  });

  factory IssueResolveResult.fromMap(Map<String, dynamic> map) {
    return IssueResolveResult(
      issueId: map['issueId'],
      key: map['key'],
      title: map['title'],
      state: map['state'],
      url: map['url'],
      status: map['status'],
    );
  }

  bool get isResolved => status == 'resolved';
  bool get isPending => status == 'pending';
}

class DescriptionSuggestionsResponse {
  final List<String> suggestions;

  DescriptionSuggestionsResponse({required this.suggestions});

  factory DescriptionSuggestionsResponse.fromMap(Map<String, dynamic> map) {
    return DescriptionSuggestionsResponse(
      suggestions: List<String>.from(map['suggestions'] ?? []),
    );
  }
}