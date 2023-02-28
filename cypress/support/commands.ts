/// <reference types="cypress" />

/** The base URL for the API. */
const apiUrl = Cypress.env('API_URL') || 'http://localhost:8080';

/**
 * Request the backend to reset the database and all the settings to test defaults.
 */
Cypress.Commands.add('backendReset', () =>
    cy.request('POST', `${apiUrl}/api/e2e/reset`).its('status').should('eq', 204));
