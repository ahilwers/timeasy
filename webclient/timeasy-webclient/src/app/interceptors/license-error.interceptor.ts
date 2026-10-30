import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';

export const licenseErrorInterceptor: HttpInterceptorFn = (req, next) => {
  const router = inject(Router);

  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 402) {
        // Payment Required - redirect to the upgrade page (RevenueCat checkout)
        router.navigate(['/upgrade']);
      }
      return throwError(() => error);
    })
  );
};
