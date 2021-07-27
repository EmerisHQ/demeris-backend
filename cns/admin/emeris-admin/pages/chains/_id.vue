<template>
  <div>
    <title-bar :title-stack="titleStack" />
    <hero-bar>
      {{ heroTitle }}
    </hero-bar>
    <section class="section is-main-section">
      <tiles>
        <card-component :title="formCardTitle" class="tile is-child">
          <form @submit.prevent="submit">
            <b-field label="ID" horizontal>
              <b-input
                v-model="chain.chain_name"
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
            <b-field horizontal>
              <b-button
                type="is-primary"
                :loading="isLoading"
                native-type="submit"
                >Save</b-button
              >
            </b-field>
          </form>
        </card-component>
        <card-component title="Chain Info" class="tile is-child">
          <hr />
          <b-field label="Name">
            <b-input
              :value="chain.display_name"
              custom-class="is-static"
              readonly
            />
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

            <section slot="empty" class="section">
              <div class="content has-text-grey has-text-centered">
                <template v-if="isLoading">
                  <p>
                    <b-icon icon="dots-horizontal" size="is-large" />
                  </p>
                  <p>Fetching data...</p>
                </template>
                <template v-else>
                  <p>
                    <b-icon icon="emoticon-sad" size="is-large" />
                  </p>
                  <p>Nothing's here&hellip;</p>
                </template>
              </div>
            </section>
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
            </template>

            <section slot="empty" class="section">
              <div class="content has-text-grey has-text-centered">
                <template v-if="isLoading">
                  <p>
                    <b-icon icon="dots-horizontal" size="is-large" />
                  </p>
                  <p>Fetching data...</p>
                </template>
                <template v-else>
                  <p>
                    <b-icon icon="emoticon-sad" size="is-large" />
                  </p>
                  <p>Nothing's here&hellip;</p>
                </template>
              </div>
            </section>
          </b-table>
        </card-component>
      </tiles>
    </section>
  </div>
</template>

<script>
import axios from "~/plugins/axios";
import api from "~/plugins/api";
import dayjs from "dayjs";
import find from "lodash/find";
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
      supply: [
        {
          denom:
            "ibc/07912C24004932CD561B1751562B22EA787F31F9821568B88F55A8F51D326722",
          amount: "5000"
        },
        {
          denom:
            "ibc/08834A76F4E5AED08690916F61EA12AA71CFD636BBA328062027DF9FA620B7E3",
          amount: "1"
        }
      ]
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
      console.log(this.chain.primary_channel)
      if (this.chain.primary_channel) {
        Object.keys(this.chain.primary_channel).forEach(
          key => (a.push({ name: key, channel: this.chain.primary_channel[key] }))
        );
      }

      console.log(a)

      return a;
    }
  },
  async created() {
    await this.loadData();
  },
  methods: {
    emptyChain() {
      return {
        chain_name: "",
        denoms: [],
        primaryChannels: {},
        display_name: ""
      };
    },
    async loadData() {
      let res = await axios.get("/chain/" + this.$route.params.id);
      this.chain = res.data.chain;
      let supply = await api.get("/chain/" + this.$route.params.id + "/supply");
      this.supply = supply.data.supply;
    },
    async update() {
      let res = await axios.post("/add", this.chain);
      if (res.status != 200) {
        this.errorText = res.error;
      } else {
        this.$nuxt.refresh();
      }
    },
    getData() {
      if (this.$route.params.id) {
        axios
          .get(`${this.$router.options.base}data-sources/chains.json`)
          .then(r => {
            const item = find(
              r.data.chains,
              item => item.chain_name === this.$route.params.id
            );

            if (item) {
              this.chain = item;
              console.log("found!");
            }
          })
          .catch(e => {
            this.$buefy.toast.open({
              message: `Error: ${e.message}`,
              type: "is-danger",
              queue: false
            });
          });
      }
    },
    input(v) {
      this.createdReadable = dayjs(v).format("MMM D, YYYY");
    },
    submit() {
      this.isLoading = true;

      setTimeout(() => {
        this.isLoading = false;

        this.$buefy.snackbar.open({
          message: "saved!",
          queue: false
        });
      }, 500);
    }
  },
  head() {
    return {
      title: "Chain"
    };
  }
};
</script>
