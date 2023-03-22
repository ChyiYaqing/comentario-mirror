import { TestBed } from '@angular/core/testing';
import { HttpInterceptorService } from './http-interceptor.service';
import { ToastService } from './toast.service';
import { ToastServiceMock } from '../_testing/mocks.spec';

describe('InterceptorService', () => {

    let service: HttpInterceptorService;

    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [
                HttpInterceptorService,
                {provide: ToastService, useValue: ToastServiceMock},
            ]
        });
        service = TestBed.inject(HttpInterceptorService);
    });

    it('is created', () => {
        expect(service).toBeTruthy();
    });
});
