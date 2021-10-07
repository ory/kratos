import {gen, website} from '../../../../helpers'
import {routes as express} from "../../../../helpers/express";
import {routes as react} from "../../../../helpers/react";

context('Testing logout flows', () => {
  [{
    route: express.login,
    app: 'express' as 'express', profile: 'email'
  }, {
    route: react.login,
    app: 'react' as 'react', profile: 'spa'
  }].forEach(({route, profile, app}) => {
    describe(`for app ${app}`, () => {
      const email = gen.email()
      const password = gen.password()

      before(() => {
        cy.proxy(app)

        cy.useConfigProfile(profile)
        cy.registerApi({email, password, fields: {'traits.website': website}})
        cy.login({email, password, cookieUrl: route})
      })

      beforeEach(() => {
        cy.visit(route)
        cy.ensureCorrectApp(app)
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
