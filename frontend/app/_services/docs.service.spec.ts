import { TestBed } from '@angular/core/testing';
import { DocsService } from './docs.service';

describe('DocsService', () => {
    let service: DocsService;

    beforeEach(() => {
        TestBed.configureTestingModule({});
        service = TestBed.inject(DocsService);
    });

    it('is created', () => {
        expect(service).toBeTruthy();
    });
});
