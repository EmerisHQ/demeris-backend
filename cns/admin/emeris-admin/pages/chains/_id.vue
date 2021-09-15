<template>
  <div>
    <b-modal :active.sync="isModalActive" has-modal-card scroll="keep">
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">Add denom to CNS</p>
        </header>
        <section class="modal-card-body">
          <form>
            <b-field label="Denom" horizontal>
              <b-input
                v-model="newDenom.name"
                custom-class="is-static"
                placeholder="Display Name"
                readonly
              />
            </b-field>
            <b-field label="Display Name" horizontal>
              <b-input
                v-model="newDenom.display_name"
                placeholder="Display Name"
                required
              />
            </b-field>
            <b-field label="Ticker" horizontal>
              <b-input
                v-model="newDenom.ticker"
                placeholder="Ticker"
                required
              />
            </b-field>
            <b-field label="Logo URL" horizontal>
              <b-input v-model="newDenom.logo" placeholder="Logo" required />
            </b-field>
          </form>
        </section>
        <footer class="modal-card-foot">
          <button class="button" type="button" @click="cancel">Cancel</button>
          <button class="button is-danger" @click="add">Add</button>
        </footer>
      </div>
    </b-modal>
    <title-bar :title-stack="titleStack" />
    <hero-bar>
      {{ heroTitle }}
    </hero-bar>
    <section class="section is-main-section">
      <tiles>
        <card-component :title="formCardTitle" class="tile is-child">
          <form>
            <b-field label="Chain Name" horizontal>
              <b-input
                v-model="chain.chain_name"
                custom-class="is-static"
                readonly
              />
            </b-field>
            <b-field label="Genesis Hash" horizontal>
              <b-input
                v-model="chain.genesis_hash"
                custom-class="is-static"
                readonly
              />
            </b-field>
            <b-field label="Display Name" horizontal>
              <b-input
                v-model="chain.display_name"
                placeholder="Chain Name Emeris"
                required
              />
            </b-field>
          </form>
        </card-component>
        <card-component
          v-if="chain.node_info != undefined"
          title="Node Info"
          class="tile is-child"
        >
          <b-field label="endpoint" horizontal>
            <b-input
              :value="chain.node_info.endpoint"
              custom-class="is-static"
              readonly
            />
          </b-field>
          <b-field label="chain_id" horizontal>
            <b-input
              :value="chain.node_info.chain_id"
              custom-class="is-static"
              readonly
            />
          </b-field>
          <b-field label="valid_block_thresh" horizontal>
            <b-input
              :value="chain.valid_block_thresh"
              custom-class="is-static"
              readonly
            />
          </b-field>
          <b-field label="derivation_path" horizontal>
            <b-input
              :value="chain.derivation_path"
              custom-class="is-static"
              readonly
            />
          </b-field>
          <b-field label="bech32 config" horizontal>
            {{ JSON.stringify(chain.node_info.bech32_config, null, 2) }}
          </b-field>
        </card-component>
      </tiles>

      <tiles>
        <card-component title="Primary Channels" class="tile is-child">
          <b-table
            :paginated="true"
            :per-page="10"
            :striped="true"
            :hoverable="true"
            default-sort="name"
            :data="primaryChannels"
          >
            <template slot-scope="props">
              <b-table-column
                label="Counterparty Chain Name"
                field="name"
                sortable
              >
                {{ props.row.name }}
              </b-table-column>

              <b-table-column label="Channel" field="channel" sortable>
                <b-input
                  v-model="chain.primary_channel[props.row.name]"
                  placeholder="channel id"
                  required
                />
              </b-table-column>
            </template>
            <!-- 
            <section slot="empty" class="section">
              <div class="content has-text-grey has-text-centered">
                <template v-if="isLoading">
                  <p>
                    <b-icon icon="dots-horizontal" size="is-large" />
                  </p>
                  <p>Fetching primary channels...</p>
                </template>
                <template v-else>
                  <p>
                    <b-icon icon="emoticon-sad" size="is-large" />
                  </p>
                  <p>No primary channels found</p>
                </template>
              </div>
            </section> -->
          </b-table>
        </card-component>
      </tiles>
      <tiles>
        <card-component title="CNS Denoms" class="tile is-child">
          <b-table
            :paginated="true"
            :per-page="10"
            :striped="true"
            :hoverable="true"
            default-sort="display_name"
            :data="chain.denoms"
          >
            <template slot-scope="props">
              <b-table-column label="base_denom" field="name" sortable>
                {{ props.row.name }}
              </b-table-column>

              <b-table-column
                label="Display Name"
                field="display_name"
                sortable
              >
                <b-input
                  v-model="props.row.display_name"
                  placeholder="Display Name"
                  required
                />
              </b-table-column>
              <b-table-column label="Ticker" field="ticker" sortable>
                <b-input
                  v-model="props.row.ticker"
                  placeholder="Ticker"
                  required
                />
              </b-table-column>
              <b-table-column label="Logo URL" field="logo_url" sortable>
                <b-input v-model="props.row.logo" placeholder="Logo" required />
              </b-table-column>
              <b-table-column label="Low Gas" field="ticker" sortable>
                <b-input
                  v-model="props.row.gas_price_levels.low"
                  placeholder="Low"
                  type="number"
                  required
                />
              </b-table-column>
              <b-table-column label="Avg Gas" field="ticker" sortable>
                <b-input
                  v-model="props.row.gas_price_levels.average"
                  placeholder="Average"
                  type="number"
                  required
                />
              </b-table-column>
              <b-table-column label="High Gas" field="ticker" sortable>
                <b-input
                  v-model="props.row.gas_price_levels.high"
                  placeholder="High"
                  type="number"
                  required
                />
              </b-table-column>

              <b-table-column label="Verified" field="verified" sortable>
                <b-checkbox v-model="props.row.verified" />
              </b-table-column>
            </template>

          </b-table>
        </card-component>
      </tiles>
      <tiles>
        <card-component title="Supply" class="tile is-child">
          <b-table
            :paginated="true"
            :per-page="10"
            :striped="true"
            :hoverable="true"
            default-sort="denom"
            :data="supply"
          >
            <template slot-scope="props">
              <b-table-column label="Denom" field="name" sortable>
                {{ props.row.denom }}
              </b-table-column>

              <b-table-column label="Amount" field="amount" sortable>
                {{ props.row.amount }}
              </b-table-column>

              <b-table-column label="Add to CNS" field="cns_add" sortable>
                <b-button
                  type="is-primary"
                  native-type="submit"
                  v-on:click="setAddingToCNS(props.row.denom)"
                  >Add</b-button
                >
              </b-table-column>
            </template>

            <section slot="empty" class="section">
              <div class="content has-text-grey has-text-centered">
                <template v-if="isLoading">
                  <p>
                    <b-icon icon="dots-horizontal" size="is-large" />
                  </p>
                  <p>Fetching supply from api...</p>
                </template>
                <template v-else>
                  <p>
                    <b-icon icon="emoticon-sad" size="is-large" />
                  </p>
                  <p>No supply found</p>
                </template>
              </div>
            </section>
          </b-table>
        </card-component>
      </tiles>
      <b-button
        type="is-primary"
        :loading="isLoading"
        native-type="submit"
        v-on:click="update()"
        >Save</b-button
      >
    </section>
  </div>
</template>

<script>
import { mapGetters } from "vuex";
import axios from "~/plugins/axios";
import api from "~/plugins/api";
import TitleBar from "@/components/TitleBar";
import HeroBar from "@/components/HeroBar";
import Tiles from "@/components/Tiles";
import CardComponent from "@/components/CardComponent";

export default {
  name: "ChainForm",
  components: {
    CardComponent,
    Tiles,
    HeroBar,
    TitleBar
  },
  data() {
    return {
      id: null,
      isLoading: false,
      chain: this.emptyChain(),
      supply: [],
      isModalActive: false,
      newDenom: {
        denom: "",
        gas_price_levels: {
          low: 0.015,
          average: 0.022,
          high: 0.042
        }
      }
    };
  },
  computed: {
    titleStack() {
      const lastCrumb = this.$route.params.id;

      return ["Admin", "Chains", lastCrumb];
    },
    heroTitle() {
      return this.chain.chain_name;
    },
    formCardTitle() {
      return "Edit Chain";
    },
    primaryChannels() {
      let a = [];
      console.log(this.chain.primary_channel);
      if (this.chain.primary_channel) {
        Object.keys(this.chain.primary_channel).forEach(key =>
          a.push({ name: key, channel: this.chain.primary_channel[key] })
        );
      }

      console.log(a);

      return a;
    }
  },
  async created() {
    this.$store.state.chains.forEach(chain => {
      if (chain.chain_name == this.$route.params.id) {
        this.chain = chain
      }
    })
  },
  async mounted() {
    await this.loadData();
  },
  methods: {
    emptyChain() {
      return {
        chain_name: "",
        denoms: [],
        primaryChannels: {},
        display_name: "",
        node_info: {}
      };
    },
    async loadData() {
      try {
        let res = await axios.get("/chain/" + this.$route.params.id);
        this.chain = res.data.chain;
        let supply = await api.get(
          "/chain/" + this.$route.params.id + "/supply"
        );
        this.supply = supply.data.supply;
      } catch (e) {
        console.log(e);
      }
    },
    async update() {
      this.chain.denoms.forEach(denom => {
        denom.gas_price_levels.low = denom.gas_price_levels.low
          ? parseFloat(denom.gas_price_levels.low)
          : 0.015;
        denom.gas_price_levels.average = denom.gas_price_levels.average
          ? parseFloat(denom.gas_price_levels.average)
          : 0.022;
        denom.gas_price_levels.high = denom.gas_price_levels.high
          ? parseFloat(denom.gas_price_levels.high)
          : 0.042;
      });
      let res = await axios.post("/add", this.chain);
      if (res.status != 200) {
        this.errorText = res.error;
      } else {
        this.$nuxt.refresh();
      }
    },
    setAddingToCNS(denom) {
      this.newDenom.name = denom;
      this.isModalActive = true;
    },
    async add() {
      var overwriting = false;
      this.chain.denoms.forEach(denom => {
        if (denom.name == this.newDenom.name) {
          overwriting = true;
        }
      });
      if (overwriting) {
        // skip overwrite
      } else {
        this.chain.denoms.push(this.newDenom);
        await this.update();
      }
      this.isModalActive = false;
    },
    cancel() {
      this.newDenom = {
        denom: "",
        gas_price_levels: {
          low: 0.015,
          average: 0.022,
          high: 0.042
        }
      };
      this.isModalActive = false;
    }
  },
  head() {
    return {
      title: "Chain"
    };
  }
};
</script>
