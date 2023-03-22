import { TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { ToastService } from './toast.service';

describe('ToastService', () => {

    let service: ToastService;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule],
        });
        service = TestBed.inject(ToastService);
    });

    it('is created', () => {
        expect(service).toBeTruthy();
    });
});
