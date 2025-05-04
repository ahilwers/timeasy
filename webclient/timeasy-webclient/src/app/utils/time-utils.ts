export function formatSecondsToReadableTime(seconds: number, locale: string): string {
  const date = new Date(0);
  date.setSeconds(seconds);
  return date.toLocaleTimeString(locale, {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: 'UTC',
    hour12: false
  });
}
