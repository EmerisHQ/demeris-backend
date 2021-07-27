<template>
  <div class="container">
    <h1>{{ chain.chain_name }}</h1>
    <h5 for="display_name">Display Name</h5>
    <input
      type="text"
      name="display_name"
      id="display_name"
      v-model="chain.display_name"
    />

    <br />
    <h5 for="logo">Logo</h5>
    <div>
      <img :src="chain.logo" class="logo" />
      <input
        type="text"
        name="logo"
        id="logo"
        v-model="chain.logo"
      />
    </div>
    <h3>Primary Channels</h3>
    <table>
      <thead>
        <tr>
          <th>Destination Chain</th>
          <th>Primary Channel</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="ch in Object.keys(chain.primary_channel)">
          <td>{{ ch }}</td>
          <td>
            <input
              type="text"
              :name="ch"
              :id="ch"
              v-model="chain.primary_channel[ch]"
            />
          </td>
        </tr>
      </tbody>
    </table>
    <h3>Denoms</h3>

    <table>
      <thead>
        <tr>
          <th>Name</th>
          <th>Display Name</th>
          <th>Verified</th>
          <th>Ticker</th>
          <th>Logo</th>
          <th>Fee Token</th>
          <th>Low Fee</th>
          <th>Average Fee</th>
          <th>High Fee</th>
        </tr>
      </thead>
      <tbody>
        <tr class="denoms" v-for="denom in chain.denoms">
          <td>{{ denom.name }}</td>
          <td>
            <input
              type="text"
              :name="'denomDisplayName' + denom.name"
              :id="'denomDisplayName' + denom.name"
              v-model="denom.display_name"
            />
          </td>
          <td>
            <input
              type="checkbox"
              :name="'verified' + denom.name"
              :id="'verified' + denom.name"
              v-model="denom.verified"
            />
          </td>
          <td>
            <input
              type="text"
              :name="'ticker' + denom.name"
              :id="'ticker' + denom.name"
              v-model="denom.ticker"
            />
          </td>
          <td>
            <img :src="denom.logo" class="logo-sm" />
            <input
              type="text"
              :name="'logo' + denom.name"
              :id="'logo' + denom.name"
              v-model="denom.logo"
            />
          </td>

          <td>
            <input
              type="checkbox"
              :name="'isFeeToken' + denom.name"
              :id="'isFeeToken' + denom.name"
              v-model="denom.fee_token"
            />
          </td>

          <td>
            <input
              type="text"
              :name="'LowTxFee' + denom.name"
              :id="'LowTxFee' + denom.name"
              v-model="denom.gas_price_levels.low"
            />
          </td>
          <td>
            <input
              type="text"
              :name="'AvgTxFee' + denom.name"
              :id="'AvgTxFee' + denom.name"
              v-model="denom.gas_price_levels.average"
            />
          </td>
          <td>
            <input
              type="text"
              :name="'HighTxFee' + denom.name"
              :id="'HighTxFee' + denom.name"
              v-model="denom.gas_price_levels.high"
            />
          </td>
        </tr>
      </tbody>
    </table>
    <button v-on:click="update()">Save Changes</button>
    <div class="error">{{ errorText }}</div>
  </div>
</template>

<script>
import axios from "~/plugins/axios";
import api from "~/sources/api";

export default {
  data() {
    return {
      chain: {
        chain_id: "",
        display_name: "",
        logo: "",
        primary_channel: {},
        denoms: [{ gas_price_levels: {low:"", average: "", high:""} }]
      },
      supply: [{}],
      errorText: ""
    };
  },
  async mounted() {
    await this.loadData()
  },
  methods: {
    async loadData() {
      let res = await axios.get("/chain/" + this.$route.params.id);
      this.chain = res.data.chain;
      // let d = await api.get(`/chain/${this.$route.params.id}/supply`)
      // this.supply = res.data.supply;

      console.log(this.chain,)
},
    async update() {
      let res = await axios.post("/add", this.chain);
      if (res.status != 200) {
        this.errorText = res.error;
      } else {
        this.$nuxt.refresh();
      }
    },
  },
};
</script>

<style scoped>
.logo {
  height: 128px;
  width: 128px;
}

.logo-sm {
  height: 28px;
  width: 28px;
}
.denoms {
  width: 100%;
}
.denom {
  width: 100%;
  margin: 10px;
  align-items: left;
}

th {
  margin: 6px;
  padding-right: 8px;
}
tr {
  margin: 6px;
  padding-right: 8px;
}
input {
  margin: 6px;
}

.error {
  color: red;
}
</style>
