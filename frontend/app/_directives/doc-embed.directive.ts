import { Directive, ElementRef, Input, OnChanges, SimpleChanges } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { DocsService } from '../_services/docs.service';
import { ConfigService } from '../_services/config.service';

@Directive({
    // eslint-disable-next-line @angular-eslint/directive-selector
    selector: '[docEmbed]',
})
export class DocEmbedDirective implements OnChanges {

    /**
     * Name of the documentation page to embed.
     */
    @Input() docEmbed?: string;

    constructor(
        private readonly element: ElementRef,
        private readonly http: HttpClient,
        private readonly cfgSvc: ConfigService,
        private readonly docsSvc: DocsService,
    ) {
        // Initially put a placeholder into the directive's element. It'll be replaced with the actual content on load
        // (or with an alert on error)
        element.nativeElement.innerHTML =
            '<div class="placeholder mb-3"></div>' +
            '<div class="placeholder py-5"></div>';
    }

    ngOnChanges(changes: SimpleChanges): void {
        // Do not bother requesting pages during an end-2-end test
        if (changes['docEmbed'] && !this.cfgSvc.isUnderTest && this.docEmbed) {
            const e = this.element.nativeElement;

            // Load the document from the documentation website, bypassing the error handler (since it's a less important resource)
            const url = this.docsSvc.getEmbedPageUrl(this.docEmbed);
            this.http.get(url, {headers: {'X-Bypass-Err-Handler': 'true'}, responseType: 'text'})
                .subscribe({
                    // Update the inner HTML of the element on success
                    next: t => e.innerHTML = t,
                    // Display error on failure
                    error: (err: Error) => e.innerHTML = '<div class="container text-center alert alert-secondary fade-in">' +
                            `Cound not load <a href="${url}" target="_blank" rel="noopener">${this.docEmbed}</a> resource:<br>` +
                            `<span class="small">${err.message}</span>` +
                        '</div>',
                });
        }
    }
}
