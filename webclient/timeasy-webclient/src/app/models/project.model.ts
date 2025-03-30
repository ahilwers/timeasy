import {DateOnly} from './date_only';

export interface Project {
  id: string;
  name: string;
  color: string;
  deadline?: DateOnly | null;
  hourlyRate: number;
  timeBudget: number;
  isActive: boolean;
}
