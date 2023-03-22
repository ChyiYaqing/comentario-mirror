import { Inject, Injectable, LOCALE_ID } from '@angular/core';
import { ConfigService } from './config.service';

@Injectable({
    providedIn: 'root',
})
export class DocsService {

    constructor(
        @Inject(LOCALE_ID) private readonly locale: string,
        private readonly cfgSvc: ConfigService,
    ) {}

    get urlHome(): string {
        return this.getPageUrl('');
    }

    get urlAbout(): string {
        return this.getPageUrl('about/');
    }

    /**
     * Return the URL of an embeddable with the given name
     * @param pageName Name of the embeddable page.
     */
    getEmbedPageUrl(pageName: string): string {
        return this.getPageUrl(`embed/${pageName}/`);
    }

    /**
     * Return a complete absolute URL for the given page and language.
     * @param path Page path within the language site.
     * @param lang Language to return a URL for. Optional, defaults to the current UI language.
     */
    getPageUrl(path: string, lang?: string): string {
        return `${this.cfgSvc.docsBaseUrl}${lang || this.locale}/${path}`;
    }
}
