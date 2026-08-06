const { description } = require('../../package')

module.exports = {
  /**
   * Ref：https://v1.vuepress.vuejs.org/config/#title
   */
  title: 'Shipshape',
  /**
   * Ref：https://v1.vuepress.vuejs.org/config/#description
   */
  description: description,

  /**
   * Extra tags to be injected to the page HTML `<head>`
   *
   * ref：https://v1.vuepress.vuejs.org/config/#head
   */
  head: [
    ['meta', { name: 'theme-color', content: '#3eaf7c' }],
    ['meta', { name: 'apple-mobile-web-app-capable', content: 'yes' }],
    ['meta', { name: 'apple-mobile-web-app-status-bar-style', content: 'black' }]
  ],

  base: "/shipshape/",

  dest: "src/.vuepress/dist/1.x",

  /**
   * Theme configuration, here is the default theme configuration for VuePress.
   *
   * ref：https://v1.vuepress.vuejs.org/theme/default-theme-config.html
   */
  themeConfig: {
    repo: '',
    editLinks: false,
    editLinkText: '',
    docsDir: '',
    lastUpdated: false,
    nav: [
      {
        text: 'Guide',
        link: '/guide/',
      },
      {
        text: 'Reference',
        link: '/reference/',
      },
      {
        text: 'GitHub',
        link: 'https://github.com/salsadigitalauorg/shipshape'
      }
    ],
    sidebar: {
      '/guide/': [
        {
          title: 'Guide',
          collapsable: false,
          children: [
            '',
            'connections',
            'collect',
            'analyse',
            'remediate',
            'outputs',
            'versions',
            '0.x',
            'gaps',
            'roadmap',
          ]
        }
      ],
      '/reference/': [
        {
          title: 'Reference',
          collapsable: false,
          children: [
            '',
            {
              title: 'Connection',
              path: '/reference/connection',
              collapsable: false,
              children: [
                '/reference/connection/mysql',
                ['/reference/connection/docker-exec', 'docker-exec'],
              ]
            },
            {
              title: 'Collect',
              path: '/reference/collect',
              collapsable: false,
              children: [
                ['/reference/collect/command', 'command'],
                ['/reference/collect/database-search', 'database:search'],
                ['/reference/collect/docker-command', 'docker:command'],
                ['/reference/collect/docker-images', 'docker:images'],
                ['/reference/collect/file-drift', 'file:drift'],
                ['/reference/collect/file-fingerprint', 'file:fingerprint'],
                ['/reference/collect/file-lookup', 'file:lookup'],
                ['/reference/collect/file-read', 'file:read'],
                ['/reference/collect/file-read-multiple', 'file:read:multiple'],
                ['/reference/collect/http-crawl', 'http:crawl'],
                ['/reference/collect/http-fetch', 'http:fetch'],
                ['/reference/collect/json-key', 'json:key'],
                ['/reference/collect/static-analysis', 'static-analysis'],
                ['/reference/collect/yaml-key', 'yaml:key'],
              ]
            },
            {
              title: 'Analyse',
              path: '/reference/analyse',
              collapsable: false,
              children: [
                ['/reference/analyse/allowed-list', 'allowed:list'],
                ['/reference/analyse/detected', 'detected'],
                ['/reference/analyse/drift', 'drift'],
                ['/reference/analyse/equals', 'equals'],
                ['/reference/analyse/not-empty', 'not:empty'],
                ['/reference/analyse/not-equals', 'not:equals'],
                ['/reference/analyse/regex-match', 'regex:match'],
                ['/reference/analyse/regex-not-match', 'regex:not-match'],
                ['/reference/analyse/static-analysis', 'static-analysis:breaches'],
              ]
            },
            {
              title: 'Remediate',
              path: '/reference/remediate',
              collapsable: false,
              children: [
                ['/reference/remediate/command', 'command'],
              ]
            },
            {
              title: 'Checks (0.x)',
              path: '/reference/checks',
              collapsable: false,
              children: [
                ['/reference/checks/file', 'file'],
                ['/reference/checks/file-diff', 'filediff'],
                ['/reference/checks/yaml', 'yaml'],
                ['/reference/checks/yaml-lint', 'yamllint'],
                ['/reference/checks/json', 'json'],
                ['/reference/checks/drupal-drush-yaml', 'drush-yaml'],
                ['/reference/checks/drupal-file-module', 'drupal-file-module'],
                ['/reference/checks/drupal-db-module', 'drupal-db-module'],
                ['/reference/checks/drupal-admin-user', 'drupal-admin-user'],
                ['/reference/checks/drupal-db-permissions', 'drupal-db-permissions'],
                ['/reference/checks/drupal-db-user-tfa', 'drupal-db-user-tfa'],
                ['/reference/checks/drupal-user-forbidden', 'drupal-user-forbidden'],
                ['/reference/checks/drupal-role-permissions', 'drupal-role-permissions'],
                ['/reference/checks/drupal-user-role', 'drupal-user-role'],
                ['/reference/checks/drupal-tracking-code', 'drupal-tracking-code'],
                ['/reference/checks/phpstan', 'phpstan'],
                ['/reference/checks/sca-application-type', 'sca:application_type'],
                ['/reference/checks/docker-base-image', 'docker:base_image'],
                ['/reference/checks/crawler', 'crawler'],
              ]
            },
          ]
        }
      ],
    },
    sidebarDepth: 2,
  },

  /**
   * Apply plugins，ref：https://v1.vuepress.vuejs.org/zh/plugin/
   */
  plugins: [
    '@vuepress/plugin-back-to-top',
    '@vuepress/plugin-medium-zoom',
  ]
}
