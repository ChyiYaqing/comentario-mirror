import { TestBed } from '@angular/core/testing';
import { ConfigService } from './config.service';

describe('ConfigService', () => {

    let service: ConfigService;

    beforeEach(() => {
        TestBed.configureTestingModule({});
        (window as any).Cypress = undefined;
    });

    it('is created', () => {
        service = TestBed.inject(ConfigService);
        expect(service).toBeTruthy();
    });

    it('sets under-test flag to false when no Cypress is available', () => {
        service = TestBed.inject(ConfigService);
        expect(service.isUnderTest).toBeFalse();
    });

    it('sets under-test flag to true with Cypress available', () => {
        (window as any).Cypress = {};
        service = TestBed.inject(ConfigService);
        expect(service.isUnderTest).toBeTrue();
    });
});
