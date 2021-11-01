<template>
  <nav id="navbar-main" class="navbar is-fixed-top">

      <div class="navbar-end">
        <div class="navbar-item has-divider">
          <span>Hi, {{ this.user.email? this.user.email : '' }} </span>
        </div>
        <a
          class="navbar-item"
          title="Log out"
          @click="logout"
        >
          <b-icon icon="logout" custom-size="default" />
          <span>Log out</span>
        </a>
      </div>
  </nav>
</template>

<script>
import { mapGetters } from 'vuex'

export default {
  name: 'NavBar',
  data() {
    return {
      isMenuNavBarActive: false,
    }
  },
  computed: {
    ...mapGetters({
      user: 'getUser'
    }),
  },
  methods: {
    logout() {
      $nuxt.$fire.auth.signOut()
      this.router.push('/')
      // this.$buefy.snackbar.open({
      //   message: 'Log out clicked',
      //   queue: false,
      // })
    },
  },
}
</script>
