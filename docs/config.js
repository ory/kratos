module.exports = {
  projectName: 'ORY Kratos',
  projectSlug: 'kratos',
  projectTagLine: 'Never build user login, user registration, 2fa, profile management ever again! Works on any operating system, cloud, with any programming language, user interface, and user experience! Written in Go.',
  updateTags: [
    {
      image: 'oryd/kratos',
      files: ['docs/docs/configure-deploy.md']
    }
  ],
  updateConfig: {
    src: '.schema/config.schema.json',
    dst: './docs/docs/reference/configuration.md'
  }
};
