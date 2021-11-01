import axios from "~/plugins/axios";

export const state = () => ({
  isNavBarVisible: true,

  isFooterBarVisible: true,

  isAsideVisible: true,
  isAsideMobileExpanded: false,

  chains: [],
})

export const actions = {
  async onAuthStateChangedAction(state, {authUser, claims}) {
    if (!authUser) {
      state.commit('SET_USER', null)

      this.$router.push({
        path: '/auth/signin',
      })
    } else {
      const {uid, email} = authUser
      state.commit('SET_USER', {uid, email})
    }
  }
}

export const mutations = {
  basic(state, payload) {
    state[payload.key] = payload.value
  },

  asideMobileStateToggle(state, payload = null) {
    const htmlClassName = 'has-aside-mobile-expanded'

    let isShow

    if (payload !== null) {
      isShow = payload
    } else {
      isShow = !state.isAsideMobileExpanded
    }

    if (isShow) {
      document.documentElement.classList.add(htmlClassName)
    } else {
      document.documentElement.classList.remove(htmlClassName)
    }

    state.isAsideMobileExpanded = isShow
  },

  SET_USER(state, user) {
    state.user = user
  },

  updateChains(state) {
    axios.get("/chains").then(res => { 
      state.chains = res.data.chains; 
      console.log("fetched chains"); 
    }).catch((e) => {
      console.error(e)
      console.log("failed to fetch chains")
    })
  }
}

export const getters = {
  getUser(state) {
    return state.user || ""
  },
}