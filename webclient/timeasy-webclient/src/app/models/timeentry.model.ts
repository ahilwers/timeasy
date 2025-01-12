export interface TimeEntry {
  id: string;
  projectId: string;
  description: string;
  startTimeUTCUnix: number;
  endTimeUTCUnix: number;
}
