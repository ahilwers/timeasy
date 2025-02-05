interface WeeklyStatistics {
  weekNumber: number;
  year: number;
  days: DailyStatistics[];
}

interface DailyStatistics {
  weekday: string;
  timeInSeconds: number;
}
