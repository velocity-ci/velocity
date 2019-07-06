module.exports = {
  title: "Velocity CI",
  themeConfig: {
    lastUpdated: 'Last Updated',
    nav: [{
        text: 'Home',
        link: '/'
      },
      {
        text: 'Guide',
        link: '/guide/'
      },
      {
        text: 'Technical Design',
        link: '/technical-design/'
      },
      {
        text: 'GitHub',
        link: 'https://github.com/velocity-ci/velocity'
      },
    ],
    sidebar: {
      '/guide/': [
        '/guide/getting-started',
        '/guide/overview',
        '/guide/examples',
        '/guide/components'
      ]
    }
  }
}