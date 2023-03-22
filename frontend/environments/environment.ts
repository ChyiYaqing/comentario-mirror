import { languages } from './languages';

export const environment = {
    production:  false,
    apiBaseUrl:  '/api', // Must be a relative or a schema-less URL for Angular's XSRF protection to work
    docsBaseUrl: 'http://localhost:1313/',
    languages,
};

/*
 * For easier debugging in development mode, you can import the following file
 * to ignore zone related error stack frames such as `zone.run`, `zoneDelegate.invokeTask`.
 *
 * This import should be commented out in production mode because it will have a negative impact
 * on performance if an error is thrown.
 */
import 'zone.js/dist/zone-error';  // Included with Angular CLI.
