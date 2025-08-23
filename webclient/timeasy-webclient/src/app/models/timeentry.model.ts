import { ExternalIssue } from './external-issue.model';

export interface TimeEntry {
  id: string;
  projectId: string;
  description: string;
  startTime: Date;
  endTime: Date | undefined;
  externalIssueId?: string;
  externalIssue?: ExternalIssue;
  pendingExternalRef?: string;
}
