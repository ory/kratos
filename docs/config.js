module.exports = {
  projectName: 'Ory Kratos',
  projectSlug: 'kratos',
  newsletter:
    'https://ory.us10.list-manage.com/subscribe?u=ffb1a878e4ec6c0ed312a3480&id=f605a41b53&group[17097][4]=1',
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
      replacer: ({ content, next, semverRegex }) =>
        content.replace(semverRegex, `${next}`),
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
    src: './driver/config/.schema/config.schema.json',
    dst: './docs/docs/reference/configuration.md'
  },
  enableRedoc: true
}
