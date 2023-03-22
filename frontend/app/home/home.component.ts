import { Component } from '@angular/core';
import { faExternalLink } from '@fortawesome/free-solid-svg-icons';
import { DocsService } from '../_services/docs.service';

@Component({
    selector: 'app-home',
    templateUrl: './home.component.html',
    styleUrls: ['./home.component.scss'],
})
export class HomeComponent {

    readonly docGetStartedUrl = this.docsSvc.getPageUrl('getting-started/');

    readonly faExternalLink = faExternalLink;

    constructor(
        private readonly docsSvc: DocsService,
    ) {}
}
