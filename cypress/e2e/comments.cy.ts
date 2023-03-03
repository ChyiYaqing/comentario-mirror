/// <reference types="cypress" />

/** The base URL for the test site. */
const baseUrl = Cypress.env('TEST_SITE_URL') || 'http://localhost:8000';

context('Comments', {baseUrl}, () => {

    before(cy.backendReset);

    it('displays comments', () => {
        cy.visit('/');
        cy.get('h1').should('have.text', 'This page has comments');
    });
});
