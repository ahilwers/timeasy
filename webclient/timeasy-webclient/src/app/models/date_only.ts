export class DateOnly {
  constructor(
    public year: number,
    public month: number,
    public day: number
  ) {}

  static fromString(dateStr: string): DateOnly {
    const [y, m, d] = dateStr.split('-').map(Number);
    return new DateOnly(y, m, d);
  }

  static fromDate(date: Date): DateOnly {
    return new DateOnly(date.getFullYear(), date.getMonth() + 1, date.getDate());
  }

  toString(): string {
    if (this, this.year === 0 || this.month === 0 || this.day === 0) {
      return '';
    }
    const m = this.month.toString().padStart(2, '0');
    const d = this.day.toString().padStart(2, '0');
    return `${this.year}-${m}-${d}`;
  }

  toJSON(): string {
    return this.toString();
  }

  toDate(): Date {
    return new Date(this.year, this.month - 1, this.day);
  }
}
