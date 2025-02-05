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
            'outputs',
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
                ['/reference/collect/database-search', 'database:search'],
                ['/reference/collect/docker-command', 'docker:command'],
                ['/reference/collect/docker-images', 'docker:images'],
                ['/reference/collect/file-lookup', 'file:lookup'],
                ['/reference/collect/file-read', 'file:read'],
                ['/reference/collect/file-read-multiple', 'file:read:multiple'],
                ['/reference/collect/yaml-key', 'yaml:key'],
              ]
            },
            {
              title: 'Analyse',
              path: '/reference/analyse',
              collapsable: false,
              children: [
                ['/reference/analyse/allowed-list', 'allowed:list'],
                ['/reference/analyse/not-empty', 'not:empty'],
                ['/reference/analyse/regex-match', 'regex:match'],
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
