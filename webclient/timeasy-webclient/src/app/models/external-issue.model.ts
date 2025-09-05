export interface ExternalIssue {
  id: string;
  projectId: string;
  provider: 'github' | 'gitlab' | 'jira';
  key: string;
  title: string;
  state: string;
  url: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface IssueResolveResult {
  issueId?: string;
  key: string;
  title?: string;
  state?: string;
  url?: string;
  status: 'resolved' | 'pending';
}

export interface DescriptionSuggestionsResponse {
  suggestions: string[];
}