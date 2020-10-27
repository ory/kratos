module.exports = {
  projectName: 'ORY Kratos',
  projectSlug: 'kratos',
  projectTagLine:
    'Never build user login, user registration, 2fa, profile management ever again! Works on any operating system, cloud, with any programming language, user interface, and user experience! Written in Go.',
  updateTags: [
    {
      image: 'oryd/kratos',
      files: ['docs/docs/quickstart.mdx']
    },
    {
      replacer: ({ content, next }) =>
        content.replace(
          /git checkout (v[0-9a-zA-Z\\.\\-]+)/gi,
          `git checkout ${next}`
        ),
      files: [
        'docs/docs/guides/zero-trust-iap-proxy-identity-access-proxy.mdx',
        'docs/docs/quickstart.mdx'
      ]
    },
    {
      replacer: ({ content, next }) =>
        content.replace(/(v[0-9a-zA-Z\\.\\-]+)/gi, `${next}`),
      files: ['docs/docs/install.md', 'docs/docs/quickstart.mdx']
    },
    {
      replacer: ({ content, next }) =>
        content.replace(
          /oryd\/kratos:(v[0-9a-zA-Z\\.\\-]+)/gi,
          `oryd/kratos:${next}-sqlite`
        ),
      files: ['quickstart.yml']
    },
    {
      image: 'oryd/kratos',
      files: [
        'quickstart-mysql.yml',
        'quickstart-crdb.yml',
        'quickstart-postgres.yml'
      ]
    }
  ],
  updateConfig: {
    src: '.schema/config.schema.json',
    dst: './docs/docs/reference/configuration.md'
  }
}
