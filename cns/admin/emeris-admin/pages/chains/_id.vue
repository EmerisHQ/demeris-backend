<template>
  <div class="container">
    <script
      src="https://upload-widget.cloudinary.com/global/all.js"
      type="text/javascript"
    ></script>
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
      <button id="upload_widget" class="cloudinary-button">Upload image</button>
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
            <button
              :id="'upload_widget_' + chain.name + '_' + denom.name"
              class="cloudinary-button"
              v-on:click="openCustomUploadWidget(denom.name)"
            >
              Upload
            </button>
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
              v-model="denom.fee_levels.low"
            />
          </td>
          <td>
            <input
              type="text"
              :name="'AvgTxFee' + denom.name"
              :id="'AvgTxFee' + denom.name"
              v-model="denom.fee_levels.average"
            />
          </td>
          <td>
            <input
              type="text"
              :name="'HighTxFee' + denom.name"
              :id="'HighTxFee' + denom.name"
              v-model="denom.fee_levels.high"
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
import axios from "~/plugins/axios";

export default {
  data() {
    return {
      chain: {
        chain_id: "",
        display_name: "",
        logo: "",
        primary_channel: {},
        denoms: [{ fee_levels: {} }],
      },
      cloudinary:{},
      errorText: ""
    };
  },
  async created() {
    await this.loadData();
    if (process.browser) {
      this.cloudinary = window.cloudinary

      let widget = window.cloudinary.createUploadWidget(
        {
          cloudName: "emeris",
          uploadPreset: "chain-logos",
          tags: [this.$route.params.id]
        },
        async (error, result) => {
          if (!error && result && result.event === "success") {
            console.log("Done! Here is the image info: ", result);

            this.chain.logo = result.info.secure_url;

            await this.update();
          }
        }
      );
      document.getElementById("upload_widget").addEventListener(
        "click",
        function() {
          widget.open();
        },
        false
      );
    }
  },
  async mounted() {
    await this.loadData();
  },

  methods: {
    async loadData() {
      let res = await axios.get("/chain/" + this.$route.params.id);
      this.chain = res.data.chain;
    },
    async update() {
      let res = await axios.post("/add", this.chain);
      if (res.status != 200) {
        this.errorText = res.error;
      } else {
        this.$nuxt.refresh();
      }
    },
    async openCustomUploadWidget(denom) {
      let widget = this.cloudinary.createUploadWidget(
        {
          cloudName: "emeris",
          uploadPreset: "denom-logos",
          tags: [this.$route.params.id]
        },
        async (error, result) => {
          if (!error && result && result.event === "success") {
            console.log("Done! Here is the image info: ", result);

            this.chain.denoms.forEach(d => {
              if (d.name == denom) {
                d.logo = result.info.secure_url;
              }
            });

            await this.update();
          }
        }
      );
      widget.open();
    }
  }
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
