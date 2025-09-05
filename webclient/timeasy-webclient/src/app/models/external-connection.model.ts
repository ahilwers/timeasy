export interface ExternalConnection {
  id: string;
  projectId: string;
  provider: 'github' | 'gitlab' | 'jira';
  projectRef: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface ConnectProjectRequest {
  provider: 'github' | 'gitlab' | 'jira';
  projectRef: string;
  oauthToken: string;
}

export interface ConnectProjectToAccountRequest {
  userAccountId: string;
  projectRef: string;
}