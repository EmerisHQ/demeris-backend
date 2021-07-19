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
  methods: {
    chainlink(chain) {
      return "/chains/" + chain.chain_name;
    }
  },
  async asyncData({ $axios, $config }) {
    const { chains } = await $axios.$get("chains");
    return { chains };
  }
};
</script>

<style lang="scss" scoped></style>
