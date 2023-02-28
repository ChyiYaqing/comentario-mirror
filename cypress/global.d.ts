/// <reference types="cypress" />

declare namespace Cypress {

    interface Chainable {

        /**
         * Request the backend to reset the database and all the settings to test defaults.
         */
        backendReset(): void;
    }
}
