import { Component } from '@angular/core';
import { DocsService } from '../_services/docs.service';
import { Paths } from '../consts';

@Component({
    selector: 'app-navbar',
    templateUrl: './navbar.component.html',
    styleUrls: ['./navbar.component.scss'],
})
export class NavbarComponent {

    readonly Paths = Paths;
    loggedIn = false;

    constructor(
        readonly docsSvc: DocsService,
    ) {}
}
