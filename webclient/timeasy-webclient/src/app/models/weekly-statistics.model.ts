export interface WeeklyStatistics {
  weekNumber: number;
  year: number;
  days: DailyStatistics[];
}

export interface DailyStatistics {
  weekday: string;
  timeInSeconds: number;
}
