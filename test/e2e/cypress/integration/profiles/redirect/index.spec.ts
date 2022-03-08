import { KRATOS_ADMIN, KRATOS_PUBLIC } from '../../../helpers'

describe('HTTP API', () => {
  it('should redirect from public to admin', () => {
    cy.request(KRATOS_PUBLIC + '/identities').should((response) => {
      expect(response.status).to.eq(200)
      expect(Array.isArray(response.body)).to.be.true
    })

    cy.request(KRATOS_PUBLIC + '/admin/identities').should((response) => {
      expect(response.status).to.eq(200)
      expect(Array.isArray(response.body)).to.be.true
    })

    cy.request(KRATOS_PUBLIC + '/schemas').should((response) => {
      expect(response.status).to.eq(200)
      expect(Array.isArray(response.body)).to.be.true
    })
  })

  it('should redirect from admin to public', () => {
    cy.request(KRATOS_ADMIN + '/identities').should((response) => {
      expect(response.status).to.eq(200)
      expect(Array.isArray(response.body)).to.be.true
    })

    cy.request(KRATOS_ADMIN + '/admin/identities').should((response) => {
      expect(response.status).to.eq(200)
      expect(Array.isArray(response.body)).to.be.true
    })

    cy.request(KRATOS_ADMIN + '/admin/schemas/default').should((response) => {
      expect(response.status).to.eq(200)
      expect(response.body['$schema']).to.eq(
        'http://json-schema.org/draft-07/schema#'
      )
    })
  })
})
