<template>
  <div class="container">
    <h1>{{ chain.chain_name }}</h1>

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
        </tr>
      </tbody>
    </table>
    <button v-on:click="update()">Save Changes</button>
    <div class="error">{{ errorText }}</div>
    <!-- <h3>Raw data</h3>
    <p>
      {{ JSON.stringify(chain, "\n", 4) }}
    </p> -->
  </div>
</template>

<script>
export default {
  data() {
    return {
      chain: {
        chain_id: "",
        primary_channel: {},
        denoms: []
      },
      errorText: ""
    };
  },
  async created() {
    await this.loadData();
  },
  async mounted() {
    await this.loadData();
  },

  methods: {
    async loadData() {
      let res = await this.$axios.get(
        "http://localhost:9999/chain/" + this.$route.params.id
      );
      console.log(res);
      this.chain = res.data.chain;
    },
    async update() {
      let res = await this.$axios.post("http://localhost:9999/add", this.chain);
      console.log(res);
      if (res.status != 200) {
        this.errorText = res.error;
      } else {
        this.$nuxt.refresh();
      }
    }
  }
};
</script>

<style scoped>
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
