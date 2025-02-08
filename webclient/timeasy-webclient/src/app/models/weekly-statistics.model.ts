export interface WeeklyStatistics {
  weekNumber: number;
  year: number;
  firstDay: Date;
  lastDay : Date;
  days: DailyStatistics[];
  sumInSeconds: number;
}

export interface DailyStatistics {
  weekday: string;
  timeInSeconds: number;
}
