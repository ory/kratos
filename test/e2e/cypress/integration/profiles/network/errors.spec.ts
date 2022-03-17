import { routes as express } from '../../../helpers/express'
import { gen } from '../../../helpers'

describe('Registration failures with email profile', () => {
  before(() => {
    cy.useConfigProfile('network')
    cy.proxy('express')
  })

  it('should not be able to register if we need a localhost schema', () => {
    cy.setDefaultIdentitySchema('localhost')
    cy.visit(express.registration, { failOnStatusCode: false })
    cy.get('.code-box').should(
      'contain.text',
      'ip 127.0.0.1 is in the 127.0.0.0/8'
    )
  })

  it('should not be able to register if we schema has a local ref', () => {
    cy.setDefaultIdentitySchema('ref')
    cy.visit(express.registration, { failOnStatusCode: false })
    cy.get('.code-box').should(
      'contain.text',
      'ip 192.168.178.1 is in the 192.168.0.0/16 range'
    )
  })

  it('should not be able to login because pre webhook uses local url', () => {
    cy.setDefaultIdentitySchema('working')
    cy.visit(express.login, { failOnStatusCode: false })
    cy.get('.code-box').should(
      'contain.text',
      'ip 192.168.178.2 is in the 192.168.0.0/16 range'
    )
  })

  it('should not be able to verify because post webhook uses local jsonnet', () => {
    cy.setDefaultIdentitySchema('working')
    cy.visit(express.registration, { failOnStatusCode: false })
    cy.get('[data-testid="node/input/traits.email"] input').type(gen.email())
    cy.get('[data-testid="node/input/traits.website"] input').type(
      'https://google.com/'
    )
    cy.get('[data-testid="node/input/password"] input').type(gen.password())
    cy.get('[type="submit"]').click()
    cy.get('.code-box').should(
      'contain.text',
      'ip 192.168.178.3 is in the 192.168.0.0/16 range'
    )
  })
})
