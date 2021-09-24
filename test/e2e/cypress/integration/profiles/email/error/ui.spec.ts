import {routes as express} from "../../../../helpers/express";
import {routes as react} from "../../../../helpers/react";

describe('Handling self-service error flows', () => {
  [{
    route: express.base,
    app: 'express',
    profile: 'email'
  }, {
    route: react.base,
    app: 'react',
    profile: 'spa'
  }].forEach(({route, app,profile}) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
      });

      it('should show the error', () => {
        cy.visit(`${route}/error?id=stub:500`, {
          failOnStatusCode: false
        })

        cy.get('code').should('contain.text', 'This is a stub error.')
      })
    })
  })
})
