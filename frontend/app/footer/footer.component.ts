import { Component } from '@angular/core';
import { DocsService } from '../_services/docs.service';
import { Paths } from '../consts';

@Component({
    selector: 'app-footer',
    templateUrl: './footer.component.html',
    styleUrls: ['./footer.component.scss'],
})
export class FooterComponent {

    readonly Paths = Paths;
    readonly year = `2022–${new Date().getFullYear()}`;

    constructor(
        readonly docsSvc: DocsService,
    ) {}
}
