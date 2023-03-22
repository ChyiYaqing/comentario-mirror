import { ChangeDetectorRef, Component } from '@angular/core';
import { Router } from '@angular/router';
import { ToastService } from '../_services/toast.service';
import { Toast } from '../_models/toast';
import { Paths } from '../consts';

@Component({
    selector: 'app-toast',
    templateUrl: './toast.component.html',
    styleUrls: ['./toast.component.scss'],
})
export class ToastComponent {

    autohide = true;

    readonly Paths = Paths;

    constructor(
        private readonly ref: ChangeDetectorRef,
        private readonly router: Router,
        private readonly toastSvc: ToastService,
    ) {}

    get toasts(): Toast[] {
        return this.toastSvc.toasts;
    }

    remove(n: Toast): void {
        this.toastSvc.remove(n);
        // Explicitly poke the change detector on element removal (it doesn't get detected automatically)
        this.ref.detectChanges();
    }

    goLogin() {
        // Redirect to login
        this.router.navigate([Paths.auth.login]);
    }
}
