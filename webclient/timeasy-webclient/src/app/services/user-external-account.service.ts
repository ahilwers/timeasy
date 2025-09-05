import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { 
  UserExternalAccount, 
  UserExternalAccountRequest, 
  UserExternalAccountResponse,
  ConnectProjectToAccountRequest
} from '../models/user-external-account.model';

@Injectable({
  providedIn: 'root'
})
export class UserExternalAccountService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  // Create a new external account
  createAccount(request: UserExternalAccountRequest): Observable<UserExternalAccount> {
    return this.http.post<UserExternalAccount>(`${this.apiUrl}/user/external-accounts`, request);
  }

  // Update an existing external account
  updateAccount(accountId: string, request: UserExternalAccountRequest): Observable<UserExternalAccount> {
    return this.http.put<UserExternalAccount>(`${this.apiUrl}/user/external-accounts/${accountId}`, request);
  }

  // Delete an external account
  deleteAccount(accountId: string): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(`${this.apiUrl}/user/external-accounts/${accountId}`);
  }

  // Get all external accounts for the user
  getAccounts(): Observable<UserExternalAccountResponse> {
    return this.http.get<UserExternalAccountResponse>(`${this.apiUrl}/user/external-accounts`);
  }

  // Get external accounts for a specific provider
  getAccountsByProvider(provider: string): Observable<UserExternalAccountResponse> {
    return this.http.get<UserExternalAccountResponse>(`${this.apiUrl}/user/external-accounts/provider/${provider}`);
  }

  // Test an external account connection
  testAccount(accountId: string): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/user/external-accounts/${accountId}/test`, {});
  }

  // Connect a project to an external account
  connectProjectToAccount(projectId: string, request: ConnectProjectToAccountRequest): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(`${this.apiUrl}/projects/${projectId}/external/connect`, request);
  }

  // Disconnect a project from external integration
  disconnectProject(projectId: string): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(`${this.apiUrl}/projects/${projectId}/external/disconnect`);
  }
}