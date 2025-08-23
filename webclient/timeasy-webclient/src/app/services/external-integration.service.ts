import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { ConnectProjectRequest, ExternalConnection } from '../models/external-connection.model';
import { DescriptionSuggestionsResponse, IssueResolveResult } from '../models/external-issue.model';

@Injectable({
  providedIn: 'root'
})
export class ExternalIntegrationService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  // Connect a project to an external provider
  connectProject(projectId: string, request: ConnectProjectRequest): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/projects/${projectId}/external/connect`, request);
  }

  // Disconnect a project from external provider
  disconnectProject(projectId: string): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(`${this.apiUrl}/projects/${projectId}/external/disconnect`);
  }

  // Get description suggestions for autocomplete
  getDescriptionSuggestions(projectId: string, query: string, limit: number = 10): Observable<DescriptionSuggestionsResponse> {
    return this.http.get<DescriptionSuggestionsResponse>(
      `${this.apiUrl}/projects/${projectId}/descriptions/suggest?q=${encodeURIComponent(query)}&limit=${limit}`
    );
  }

  // Resolve an issue reference
  resolveIssue(projectId: string, input: string): Observable<IssueResolveResult> {
    return this.http.get<IssueResolveResult>(
      `${this.apiUrl}/projects/${projectId}/issues/resolve?input=${encodeURIComponent(input)}`
    );
  }

  // Trigger manual sync of project issues
  syncProjectIssues(projectId: string): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/projects/${projectId}/external/sync`, {});
  }

  // Resolve pending external references
  resolvePendingReferences(): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/external/resolve-pending`, {});
  }

  // Check if input matches issue patterns
  detectIssuePattern(input: string): { key: string; provider: string } | null {
    // GitHub/GitLab pattern: #123 or owner/repo#123
    const githubPattern = /(?:^|\s)(?:[\w.-]+\/[\w.-]+)?#(\d+)(?:\s|$)/;
    const githubMatch = input.match(githubPattern);
    if (githubMatch) {
      return {
        key: `#${githubMatch[1]}`,
        provider: 'github' // Could also be gitlab
      };
    }

    // Jira pattern: ABC-123
    const jiraPattern = /(?:^|\s)([A-Z][A-Z0-9]+-\d+)(?:\s|$)/;
    const jiraMatch = input.match(jiraPattern);
    if (jiraMatch) {
      return {
        key: jiraMatch[1],
        provider: 'jira'
      };
    }

    return null;
  }
}