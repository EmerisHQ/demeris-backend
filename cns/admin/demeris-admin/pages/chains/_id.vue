<template>
  <div class="container">

    <h1>{{ chain.chain_name }}</h1>

    <h3>Primary Channels</h3>

    <div v-for="ch in Object.keys(chain.primary_channel)">
      <label :for="ch">Primay Channel {{ ch }}</label>
      <input
        type="text"
        :name="ch"
        :id="ch"
        v-model="chain.primary_channel[ch]"
      /><br />
      <button v-on:click="updatePrimaryChannel(ch)">Update</button>

    </div>

    <h3>Denoms</h3>
    <div v-for="denom in chain.denoms">
      {{ denom.name }}

      <label :for="'denomDisplayName' + denom.name">Display Name</label>
      <input
        type="text"
        :name="'denomDisplayName' + denom.name"
        :id="'denomDisplayName' + denom.name"
        v-model="denom.display_name"
      />

      <label :for="'verified' + denom.name">Verified</label>
      <input
        type="checkbox"
        :name="'verified' + denom.name"
        :id="'verified' + denom.name"
        v-model="denom.verified"
      />
    </div>
    <button v-on:click="updateDenoms()">Update</button>

    <h3>Raw data</h3>
    <p>
      {{ JSON.stringify(chain, "\n", 4) }}
    </p>
  </div>
</template>

<script>
export default {
  data() {
    return {
      // chains: []
      chain: {
        chain_id: "",
        primary_channel: {},
        denoms: [] 
      }
    };
  },
  async created() {
    console.log("loaded!");

    await this.loadData()
    // let chain = await this.$axios.get(
    //   "http://localhost:9999/chains/" + this.$route.params.id
    // );
    // console.log(chain);
    // this.chain = chain;
  },
    async mounted() {
    console.log("loaded!");

    await this.loadData()
    // let chain = await this.$axios.get(
    //   "http://localhost:9999/chains/" + this.$route.params.id
    // );
    // console.log(chain);
    // this.chain = chain;
  },

  methods: {
    async updatePrimaryChannel(dest_chain) {
      // console.log(JSON.stringify(this.chain.primary_channel, "\n", 2));
      let request = {
        "chain_name": this.$route.params.id,
        "dest_chain": dest_chain,
        "primary_channel": this.chain.primary_channel[dest_chain]
      }
      console.log(JSON.stringify(request, "\n", 2));
      await this.$axios.post("http://localhost:9999/update_primary_channel", request)
      this.$nuxt.refresh()
    },
    async updateDenoms() {
      // console.log(JSON.stringify(this.chain.denoms, "\n", 2));
      let request = {
        "chain_name": this.$route.params.id,
        "denoms": this.chain.denoms
      }
      console.log(JSON.stringify(request, "\n", 2));
      await this.$axios.post("http://localhost:9999/denoms", request)
      this.$nuxt.refresh()
    },
    async loadData() {
      let res = await this.$axios.get(
        "http://localhost:9999/chain/" + this.$route.params.id
      );
      console.log(res);
      this.chain = res.data.chain;
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
