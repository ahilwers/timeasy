export interface UserExternalAccount {
  id: string;
  userId: string;
  provider: 'github' | 'gitlab' | 'jira';
  accountName: string;
  baseURL?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface UserExternalAccountRequest {
  provider: 'github' | 'gitlab' | 'jira';
  accountName: string;
  oauthToken: string;
  baseURL?: string;
}

export interface ConnectProjectToAccountRequest {
  userAccountId: string;
  projectRef: string;
}

export interface UserExternalAccountResponse {
  accounts: UserExternalAccount[];
}