import { defineConfig } from 'cypress';

export default defineConfig({
    e2e: {
        setupNodeEvents(on, config) {
            // implement node event listeners here
        },
        // Backend's (API and the admin app) base URL
        baseUrl: 'http://localhost:8080',
    },
});
