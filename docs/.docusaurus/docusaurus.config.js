export default {
  "plugins": [
    [
      "@docusaurus/plugin-content-docs",
      {
        "path": "docs",
        "sidebarPath": "/Users/foobar/go/src/github.com/ory/kratos/docs/contrib/sidebar.js",
        "editUrl": "https://github.com/ory/kratos/edit/master/docs",
        "routeBasePath": "",
        "showLastUpdateAuthor": true,
        "showLastUpdateTime": true,
        "remarkPlugins": [
          null
        ]
      }
    ],
    [
      "@docusaurus/plugin-content-pages"
    ],
    [
      "@docusaurus/plugin-google-analytics"
    ],
    [
      "@docusaurus/plugin-sitemap"
    ]
  ],
  "themes": [
    [
      "@docusaurus/theme-classic",
      {
        "customCss": "/Users/foobar/go/src/github.com/ory/kratos/docs/src/css/theme.css"
      }
    ],
    [
      "@docusaurus/theme-search-algolia"
    ]
  ],
  "customFields": {},
  "themeConfig": {
    "googleAnalytics": {
      "trackingID": "UA-71865250-1",
      "anonymizeIP": true
    },
    "algolia": {
      "apiKey": "8463c6ece843b377565726bb4ed325b0",
      "indexName": "ory",
      "algoliaOptions": {
        "facetFilters": [
          "tags:kratos",
          "version:v0.2"
        ]
      }
    },
    "navbar": {
      "logo": {
        "alt": "ORY Kratos",
        "src": "img/logo-kratos.svg",
        "href": "https://www.ory.sh/kratos"
      },
      "links": [
        {
          "to": "index",
          "activeBasePath": "kratos/docs",
          "label": "Docs",
          "position": "left"
        },
        {
          "href": "https://www.ory.sh/docs",
          "label": "Ecosystem",
          "position": "left"
        },
        {
          "href": "https://www.ory.sh/blog",
          "label": "Blog",
          "position": "left"
        },
        {
          "href": "https://community.ory.sh",
          "label": "Forum",
          "position": "left"
        },
        {
          "href": "https://www.ory.sh/chat",
          "label": "Chat",
          "position": "left"
        },
        {
          "href": "https://github.com/ory/kratos",
          "label": "GitHub",
          "position": "left"
        },
        {
          "label": "v0.2",
          "position": "right",
          "to": "versions"
        }
      ]
    },
    "footer": {
      "style": "dark",
      "copyright": "Copyright © 2020 ORY GmbH",
      "links": [
        {
          "title": "Company",
          "items": [
            {
              "label": "Imprint",
              "href": "https://www.ory.sh/imprint"
            },
            {
              "label": "Privacy",
              "href": "https://www.ory.sh/privacy"
            },
            {
              "label": "Terms",
              "href": "https://www.ory.sh/tos"
            }
          ]
        }
      ]
    }
  },
  "title": "ORY Kratos",
  "tagline": "Never build user login, user registration, 2fa, profile management ever again! Works on any operating system, cloud, with any programming language, user interface, and user experience! Written in Go.",
  "url": "https://www.ory.sh/",
  "baseUrl": "/kratos/docs/",
  "favicon": "img/favico.png",
  "organizationName": "ory",
  "projectName": "kratos"
};