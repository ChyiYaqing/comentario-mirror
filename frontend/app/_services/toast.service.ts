import { Injectable } from '@angular/core';
import { NavigationStart, Router } from '@angular/router';
import { filter } from 'rxjs/operators';
import { Severity, Toast } from '../_models/toast';

/**
 * Service for showing toasts for the user.
 */
@Injectable({
    providedIn: 'root',
})
export class ToastService {

    readonly toasts: Toast[] = [];
    private _keepOnRouteChange = false;

    constructor(private router: Router) {
        // Remove toasts on route change
        router.events
            .pipe(filter(event => event instanceof NavigationStart))
            .subscribe(() => {
                // If we're to skip this route change
                if (this._keepOnRouteChange) {
                    this._keepOnRouteChange = false;

                // Clear all toasts otherwise
                } else {
                    this.clear();
                }
            });
    }

    /**
     * Remove all toasts.
     */
    clear(): void {
        this.toasts.length = 0;
    }

    /**
     * Add a new toast at the head of the list.
     * @param severity Toast severity.
     * @param id Optional toast ID.
     * @param errorCode Optional toast error code.
     * @param message Optional toast message.
     * @param details Optional message details.
     */
    addToast(severity: Severity, id: string, errorCode?: number, message?: string, details?: string): ToastService {
        // Remove any repeated toasts beforehand
        this.removeToastsOfSeverity(severity);

        // Insert a new toast at the beginning
        this.toasts.splice(0, 0, new Toast(severity, id, errorCode, message, details));
        return this;
    }

    /**
     * Do not delete the toasts on the first upcoming route change.
     */
    keepOnRouteChange(): ToastService {
        this._keepOnRouteChange = true;
        return this;
    }

    /**
     * Remove the given toast from the list.
     * @param toast Toast to remove
     */
    remove(toast: Toast): void {
        const idx = this.toasts.indexOf(toast);
        if (idx > -1) {
            this.toasts.splice(idx, 1);
        }
    }

    /**
     * Add an info toast.
     * @param id Optional toast ID.
     * @param errorCode Optional toast error code.
     * @param message Optional toast message.
     * @param details Optional message details.
     */
    info(id: string, errorCode?: number, message?: string, details?: string): ToastService {
        return this.addToast(Severity.INFO, id, errorCode, message, details);
    }

    /**
     * Add an success toast.
     * @param id Optional toast ID.
     * @param errorCode Optional toast error code.
     * @param message Optional toast message.
     * @param details Optional message details.
     */
    success(id: string, errorCode?: number, message?: string, details?: string): ToastService {
        return this.addToast(Severity.SUCCESS, id, errorCode, message, details);
    }

    /**
     * Add an warning toast.
     * @param id Optional toast ID.
     * @param errorCode Optional toast error code.
     * @param message Optional toast message.
     * @param details Optional message details.
     */
    warning(id: string, errorCode?: number, message?: string, details?: string): ToastService {
        return this.addToast(Severity.WARNING, id, errorCode, message, details);
    }

    /**
     * Add an error toast.
     * @param id Optional toast ID.
     * @param errorCode Optional toast error code.
     * @param message Optional toast message.
     * @param details Optional message details.
     */
    error(id: string, errorCode?: number, message?: string, details?: string): ToastService {
        return this.addToast(Severity.ERROR, id, errorCode, message, details);
    }

    /**
     * Remove all toasts of the specified severity.
     */
    private removeToastsOfSeverity(severity: Severity) {
        // Iterate the toasts, starting from the end
        for (let i = this.toasts.length-1; i >= 0; i--) {
            if (this.toasts[i].severity === severity) {
                this.toasts.splice(i, 1);
            }
        }
    }

}
