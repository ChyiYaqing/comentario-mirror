/**
 * Declarations of mocked services.
 */
import { noop } from 'rxjs';
import { ToastService } from '../_services/toast.service';

// noinspection JSUnusedLocalSymbols
export const ToastServiceMock: Partial<ToastService> = {
    clear:             noop,
    addToast:          (severity, id?, errorCode?, message?) => this as any,
    keepOnRouteChange: () => this as any,
    remove:            noop,
    info:              (id, errorCode?, message?, details?) => this as any,
    success:           (id, errorCode?, message?, details?) => this as any,
    warning:           (id, errorCode?, message?, details?) => this as any,
    error:             (id, errorCode?, message?, details?) => this as any,
};
