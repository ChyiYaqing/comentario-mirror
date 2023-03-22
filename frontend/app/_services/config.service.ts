import { Injectable } from '@angular/core';
import { NgbConfig, NgbToastConfig } from '@ng-bootstrap/ng-bootstrap';
import { environment } from '../../environments/environment';

@Injectable({
    providedIn: 'root',
})
export class ConfigService {

    /**
     * Toast hiding delay in milliseconds.
     */
    static readonly TOAST_DELAY = 10000;

    /**
     * Whether the system is running under an end-2-end test.
     */
    readonly isUnderTest: boolean = false;

    constructor(
        private readonly ngbConfig: NgbConfig,
        private readonly toastConfig: NgbToastConfig,
    ) {
        // Detect if the e2e-test is active
        this.isUnderTest = !!(window as any).Cypress;

        // Disable animations with e2e to speed up the tests
        ngbConfig.animation = !this.isUnderTest;
        toastConfig.delay = ConfigService.TOAST_DELAY;
    }

    /**
     * Return the base URL for embedded and linked documentation pages.
     */
    get docsBaseUrl(): string {
        return environment.docsBaseUrl;
    }
}
