import {gen,APP_URL} from "../../../../helpers";

context('Registration', () => {
  beforeEach(() => {
    cy.visit(APP_URL + '/auth/registration')
    cy.deleteMail()
  })

  it('should sign up and receive a verification email', () => {
    const email = gen.email()
    cy.register({email})

    cy.getMail().then((body) => {
      const message = body.mailItems[0]

      expect(message.subject.trim()).to.equal('Please verify your email address')
      expect(message.fromAddress.trim()).to.equal('no-reply@ory.kratos.sh')
      expect(message.toAddresses).to.have.length(1)
      expect(message.toAddresses[0].trim()).to.equal(email)

      const parser = new DOMParser();
      const content = parser.parseFromString( message.body, 'text/html');
      const link = content.querySelector('a')

      expect(link).to.not.be.undefined
      expect(link.href).to.contain('http://127.0.0.1:4455/')

      cy.visit(link.href)
    })
  })
})
