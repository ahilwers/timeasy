import {DateOnly} from './date_only';

export interface ExternalConnection {
  id: string;
  projectId: string;
  userAccountId: string;
  provider: 'github' | 'gitlab' | 'jira';
  projectRef: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Project {
  id: string;
  name: string;
  color: string;
  deadline?: DateOnly | null;
  hourlyRate: number;
  timeBudget: number;
  isActive: boolean;
  externalConnection?: ExternalConnection;
}
