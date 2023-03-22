import { Injectable } from '@angular/core';
import { HttpErrorResponse, HttpEvent, HttpHandler, HttpInterceptor, HttpRequest } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { ToastService } from './toast.service';

@Injectable({
    providedIn: 'root',
})
export class HttpInterceptorService implements HttpInterceptor {

    constructor(
        private readonly toastSvc: ToastService,
    ) {}

    intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>>{
        // If we see this fake header in the request, it means the error handling logic is implemented on the calling
        // side and we need to bypass it here
        const bypass = req.headers.has('X-Bypass-Err-Handler');
        if (bypass) {
            // We also remove it from the request (by creating a clone request) because otherwise CORS would refuse the
            // request
            req = req.clone({headers: req.headers.delete('X-Bypass-Err-Handler')});
        }

        // Run the original handler(s)
        return next.handle(req)
            .pipe(catchError((error: HttpErrorResponse) => {
                // If we're not to bypass the error handling
                if (!bypass) {
                    const errorId = error.error?.id;
                    const details = error.error?.details;

                    // Client-side error
                    if (error.error instanceof ErrorEvent) {
                        this.toastSvc.error(errorId, -1, error.error?.message, details);

                    // 401 Unauthorized from the backend
                    } else if (error.status === 401) {
                        // Remove the current principal if it's a 401 error, which means the user isn't logged in (anymore)
                        // TODO

                        // Add an info toast that the user has to relogin
                        this.toastSvc.info(errorId, 401, error.message, details);

                    // Any other server-side error
                    } else {
                        this.toastSvc.error(errorId, error.status, `${error.message} (error code: ${error.status})`, details);
                    }
                }

                // Rethrow the error
                return throwError(() => error);
            }));
    }
}
