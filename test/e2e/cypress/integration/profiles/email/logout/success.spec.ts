import {gen, website} from '../../../../helpers'
import {routes as express} from "../../../../helpers/express";
import {routes as react} from "../../../../helpers/react";

context('Testing logout flows', () => {
  [{
    route: express.login,
    app: 'express', profile: 'email'
  }, {
    route: react.login,
    app: 'react', profile: 'spa'
  }].forEach(({route, profile, app}) => {
    describe(`for app ${app}`, () => {
      const email = gen.email()
      const password = gen.password()

      before(() => {
        cy.useConfigProfile(profile)
        cy.registerApi({email, password, fields: {'traits.website': website}})
        cy.login({email, password, cookieUrl: route})
      })

      beforeEach(() => {
        cy.visit(route)
      })

      it('should sign out and be able to sign in again', () => {
        cy.getSession()
        cy.getCookie('ory_kratos_session').should('not.be.null')
        cy.get('*[data-testid="logout"]').click()
        cy.noSession()
        cy.url().should('include', '/login')
        cy.getCookie('ory_kratos_session').should('be.null')
      })
    })
  })
})
