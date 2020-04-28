// FIXME we need to implement account takeover measures
// FIXME and test for that. See: https://github.com/ory/kratos/issues/292
import {APP_URL, identity} from "../../../../helpers"

context('Login', () => {
  beforeEach(() => {
    cy.visit(APP_URL + '/settings')
  })
  })
