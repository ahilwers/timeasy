export interface TimeEntry {
  id: string;
  projectId: string;
  description: string;
  startTime: Date;
  endTime: Date | undefined;
  externalIssueId?: string;
  pendingExternalRef?: string;
}
