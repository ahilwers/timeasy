import { Component, signal, effect } from '@angular/core';

@Component({
  selector: 'app-timer',
  standalone: true,
  templateUrl: './timer.component.html',
  styleUrls: ['./timer.component.css']
})
export class TimerComponent {
  private intervalId: any;
  private startTimestamp: number = 0;
  private elapsedMs = signal<number>(0);
  private running = false;

  start(startTime?: Date) {
    if (this.running) return;

    if (startTime) {
      const now = Date.now();
      this.startTimestamp = startTime.getTime();
      this.elapsedMs.set(now - this.startTimestamp);
    } else {
      this.startTimestamp = Date.now() - this.elapsedMs();
    }

    this.running = true;
    this.intervalId = setInterval(() => {
      const now = Date.now();
      this.elapsedMs.set(now - this.startTimestamp);
    }, 1000);
  }

  stop() {
    if (this.running) {
      clearInterval(this.intervalId);
      this.running = false;
    }
  }

  reset() {
    this.stop();
    this.elapsedMs.set(0);
  }

  get formattedTime(): string {
    const totalSeconds = Math.floor(this.elapsedMs() / 1000);
    const hours = Math.floor(totalSeconds / 3600).toString().padStart(2, '0');
    const minutes = Math.floor((totalSeconds % 3600) / 60).toString().padStart(2, '0');
    const seconds = (totalSeconds % 60).toString().padStart(2, '0');
    return `${hours}:${minutes}:${seconds}`;
  }
}
