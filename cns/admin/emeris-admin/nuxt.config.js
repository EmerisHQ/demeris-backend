export default {
  // Target: https://go.nuxtjs.dev/config-target
  target: 'static',

  generate: {
    exclude: [
      /^\/chain/ // path starts with /chain
    ]
  },

  // Global page headers: https://go.nuxtjs.dev/config-head
  head: {
    title: 'demeris-admin',
    htmlAttrs: {
      lang: 'en'
    },
    meta: [
      { charset: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { hid: 'description', name: 'description', content: '' }
    ],
    link: [
      { rel: 'icon', type: 'image/x-icon', href: '/admin/favicon.ico' },
      { rel: 'dns-prefetch', href: 'https://fonts.gstatic.com' },
      {
        rel: 'stylesheet',
        type: 'text/css',
        href: 'https://fonts.googleapis.com/css?family=Nunito',
      },
      {
        rel: 'stylesheet',
        type: 'text/css',
        href:
          'https://cdn.materialdesignicons.com/4.9.95/css/materialdesignicons.min.css',
      },
    ]
  },

  // Global CSS: https://go.nuxtjs.dev/config-css
  css: ['./assets/scss/main.scss'],

  // Plugins to run before rendering page: https://go.nuxtjs.dev/config-plugins
  plugins: [{ src: '~/plugins/after-each.js', mode: 'client' }],

  // Auto import components: https://go.nuxtjs.dev/config-components
  components: false,

  // Modules for dev and build (recommended): https://go.nuxtjs.dev/config-modules
  buildModules: [
  ],

  // Modules: https://go.nuxtjs.dev/config-modules
  modules: [
    // Doc: https://buefy.github.io/#/documentation
    ['nuxt-buefy', { materialDesignIcons: false }],
    'bootstrap-vue/nuxt',
    '@nuxtjs/axios',
  ],

  // Build Configuration: https://go.nuxtjs.dev/config-build
  build: {
  },

  axios: {
    baseUrl: "http://localhost:8000/v1/cns" || "/v1/cns",
    apiUrl: "http://localhost:8000/v1" || "/v1",
    cnsUrl: "http://localhost:8000/v1/cns" || "/v1/cns",
    // baseUrl: process.env.CNS_URL || "/v1/cns",
    // cnsUrl: process.env.CNS_URL || "/v1/cns",
    // apiUrl: process.env.API_URL || "/v1"
  },

  router: {
    base: process.env.BASE_URL || "/admin"
  }
};
