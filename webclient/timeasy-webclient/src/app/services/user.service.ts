import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject, tap } from 'rxjs';
import { environment } from '../../environments/environment';
import { AuthProvider, UserProfile, PasswordChangeRequest } from '../interfaces/auth-provider.interface';

@Injectable({
  providedIn: 'root'
})
export class UserService implements AuthProvider {
  private readonly apiUrl = `${environment.apiUrl}/user`;
  private readonly http = inject(HttpClient);
  
  private userSubject = new BehaviorSubject<UserProfile | null>(null);
  public user$ = this.userSubject.asObservable();

  getCurrentUser(): Observable<UserProfile> {
    return this.http.get<UserProfile>(`${this.apiUrl}/profile`).pipe(
      tap(user => this.userSubject.next(user))
    );
  }

  updateUserProfile(profile: Partial<UserProfile>): Observable<UserProfile> {
    return this.http.put<UserProfile>(`${this.apiUrl}/profile`, profile).pipe(
      tap(user => this.userSubject.next(user))
    );
  }

  changePassword(passwordRequest: PasswordChangeRequest): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/change-password`, passwordRequest);
  }

  refreshUserData(): Observable<UserProfile> {
    return this.getCurrentUser();
  }

  updateLanguagePreference(language: string): Observable<UserProfile> {
    return this.updateUserProfile({ language });
  }
}