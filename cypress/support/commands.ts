/// <reference types="cypress" />

/**
 * Request the backend to reset the database and all the settings to test defaults.
 */
Cypress.Commands.add('backendReset', () =>
    cy.request('POST', '/api/e2e/reset').its('status').should('eq', 204));
