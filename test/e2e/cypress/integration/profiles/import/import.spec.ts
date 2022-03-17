import { routes as express } from '../../../helpers/express'
import { gen, KRATOS_ADMIN, website } from '../../../helpers'

context('Import Identities', () => {
  before(() => {
    cy.useConfigProfile('oidc')
    cy.proxy('express')
  })

  beforeEach(() => {
    cy.clearAllCookies()
  })

  const password = gen.password()
  for (const tc of [
    {
      name: 'cleartext',
      config: {
        password
      },
      checkPassword: password
    },
    {
      name: 'pbkdf2',
      config: {
        hashed_password:
          '$pbkdf2-sha256$i=1000,l=128$e8/arsEf4cvQihdNgqj0Nw$5xQQKNTyeTHx2Ld5/JDE7A'
      },
      checkPassword: '123456'
    },
    {
      name: 'bcrypt',
      config: {
        hashed_password:
          '$2a$10$ZsCsoVQ3xfBG/K2z2XpBf.tm90GZmtOqtqWcB5.pYd5Eq8y7RlDyq'
      },
      checkPassword: '123456'
    },
    {
      name: 'argon2id',
      config: {
        hashed_password:
          '$argon2id$v=19$m=16,t=2,p=1$bVI1aE1SaTV6SGQ3bzdXdw$fnjCcZYmEPOUOjYXsT92Cg'
      },
      checkPassword: '123456'
    }
  ]) {
    it(`should be able to sign in using an imported password (${tc.name})`, () => {
      const email = gen.email()
      cy.request('POST', `${KRATOS_ADMIN}/identities`, {
        schema_id: 'default',
        traits: {
          email,
          website
        },
        credentials: {
          password: {
            config: tc.config
          }
        }
      })

      cy.visit(express.login)

      // Try to sign in with an incorrect password
      cy.get('input[name="identifier"]').type(email)
      cy.get('input[name="password"]').type('invalid-password')
      cy.submitPasswordForm()
      cy.get('*[data-testid="ui/message/4000006"]').should(
        'contain.text',
        'credentials are invalid'
      )

      // But with correct password it succeeds
      cy.get('input[name="password"]').type(tc.checkPassword)
      cy.submitPasswordForm()

      cy.location('pathname').should('not.contain', '/login')
      cy.getSession().should((session) => {
        const { identity } = session
        expect(identity.id).to.not.be.empty
        expect(identity.traits.website).to.equal(website)
      })
    })
  }

  it(`should be able to sign in using imported oidc credentials`, () => {
    const email = gen.email()
    const website = 'https://' + gen.password() + '.com'
    cy.request('POST', `${KRATOS_ADMIN}/identities`, {
      schema_id: 'default',
      traits: {
        email,
        website
      },
      credentials: {
        oidc: {
          config: {
            providers: [
              {
                provider: 'hydra',
                subject: email
              }
            ]
          }
        }
      }
    })

    cy.visit(express.login)
    cy.triggerOidc({ url: express.login })

    cy.get('#username').clear().type(email)
    cy.get('#remember').click()
    cy.get('#accept').click()
    cy.get('[name="scope"]').each(($el) => cy.wrap($el).click())
    cy.get('#remember').click()
    cy.get('#accept').click()

    cy.getSession().should((session) => {
      const { identity } = session
      expect(identity.id).to.not.be.empty
      expect(identity.traits.website).to.equal(website)
    })
  })
})
