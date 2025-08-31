import { Observable } from 'rxjs';

export interface UserProfile {
  id: string;
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  displayName: string;
  language?: string;
  attributes?: { [key: string]: string[] };
}

export interface PasswordChangeRequest {
  currentPassword: string;
  newPassword: string;
}

export interface AuthProvider {
  getCurrentUser(): Observable<UserProfile>;
  updateUserProfile(profile: Partial<UserProfile>): Observable<UserProfile>;
  changePassword(passwordRequest: PasswordChangeRequest): Observable<void>;
  refreshUserData(): Observable<UserProfile>;
}