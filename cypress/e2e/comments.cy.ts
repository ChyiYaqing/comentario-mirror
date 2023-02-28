/// <reference types="cypress" />

context('Comments', {baseUrl: 'http://localhost:8000/'}, () => {

    before(cy.backendReset);

    it('displays comments', () => {
        cy.visit('/');
        cy.get('h1').should('have.text', 'This page has comments');
    });
});
