import axios from "~/plugins/axios";

export const state = () => ({
  isNavBarVisible: true,

  isFooterBarVisible: true,

  isAsideVisible: true,
  isAsideMobileExpanded: false,

  chains: [],
})

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

  updateChains(state) {
    axios.get("/chains").then(res => { 
      state.chains = res.data.chains; 
      console.log("fetched chains"); 
    }).catch((e) => {
      console.log(e)
      console.log("failed to fetch chains")
    })
  }
}
