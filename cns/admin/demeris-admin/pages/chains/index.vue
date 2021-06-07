<template>
  <div class="container">
      <h1>CNS Chains</h1>
    <div v-for="chain in chains">
      {{ chain.chain_name || "no chains" }}
      <router-link :to="chainlink(chain)"> Edit </router-link>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      chains: []
    };
  },
  async mounted() {
    console.log("loaded!");
    let res = await this.$axios.get("http://localhost:9999/chains");
    this.chains = res.data.chains;
  },
  methods: {
      chainlink(chain) {
          return "/chains/" + chain.chain_name
      }
  }
  //   async asyncData({ $axios }) {
  //     let { data } = await $axios.get("/card_sets/170");
  //     return { incidents: data.data.incidents };
  //   }
  // mounted() {
  //   this.getIncidents();
  // },
};
</script>

<style lang="scss" scoped></style>
