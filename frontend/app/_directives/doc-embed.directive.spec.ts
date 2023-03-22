import { Component, DebugElement, LOCALE_ID } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DocEmbedDirective } from './doc-embed.directive';

@Component({
    template: '<div docEmbed="test"></div>',
})
class TestComponent {
}

describe('DocEmbedDirective', () => {

    let httpTestingController: HttpTestingController;
    let fixture: ComponentFixture<TestComponent>;
    let de: DebugElement[];
    let div: HTMLDivElement;

    beforeEach(() => {
        fixture = TestBed.configureTestingModule({
            declarations: [DocEmbedDirective, TestComponent],
            imports: [HttpClientTestingModule],
            providers: [{ provide: LOCALE_ID, useValue: 'zh' }],
        })
        .createComponent(TestComponent);

        httpTestingController = TestBed.inject(HttpTestingController);

        fixture.detectChanges();

        // All elements with an attached directive
        de = fixture.debugElement.queryAll(By.directive(DocEmbedDirective));
        div = de[0].nativeElement as HTMLDivElement;
    });

    it('has one element', () => {
        expect(de.length).toBe(1);
    });

    it('contains a placeholder initially', () => {
        // The element is initially empty
        expect(div.innerHTML).toMatch(/<div class="placeholder.*">/);
        // No classes
        expect(div.classList.value).toBe('');
    });

    it('requests and embeds a doc page', () => {
        // Mock the request
        const req = httpTestingController.expectOne('http://localhost:1313/zh/embed/test/');
        expect(req.request.method).toEqual('GET');
        req.flush('<h1>Super page!</h1>');

        // After the request the HTML is updated
        expect(div.innerHTML).toBe('<h1>Super page!</h1>');

        // Assert there are no more pending requests
        httpTestingController.verify();
    });

    it('displays alert on error', () => {
        // Mock the request
        const req = httpTestingController.expectOne('http://localhost:1313/zh/embed/test/');
        expect(req.request.method).toEqual('GET');
        req.flush(null, {status: 500, statusText: 'Ouch'});

        // After the request the HTML is updated
        expect(div.innerHTML).toContain('Cound not load <a href="http://localhost:1313/zh/embed/test/" target="_blank" rel="noopener">test</a> resource');

        // Assert there are no more pending requests
        httpTestingController.verify();
    });
});
