import { languages } from './languages';

export const environment = {
    production:  true,
    apiBaseUrl:  '/api', // Must be a relative or a schema-less URL for Angular's XSRF protection to work
    docsBaseUrl: 'https://docs.comentario.app/',
    languages,
};
