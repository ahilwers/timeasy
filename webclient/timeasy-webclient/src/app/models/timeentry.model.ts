export interface TimeEntry {
  Id: string;
  projectId: string;
  description: string;
  startTimeUTCUnix: number;
  EndTimeUTCUnix?: number;
}
