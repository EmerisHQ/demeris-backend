<template>
  <div>
    <b-table
      :checked-rows.sync="checkedRows"
      :checkable="true"
      :loading="isLoading"
      :per-page="perPage"
      :striped="true"
      :paginated="true"
      :hoverable="true"
      :data="chains"
      default-sort="chain_name"
    >
      <template slot-scope="props">
        <b-table-column class="has-no-head-mobile is-image-cell">
          <div class="image">
            <img
              :src="props.row.logo"
              :alt="props.row.chain_name"
              class="is-rounded"
            />
          </div>
        </b-table-column>
        <b-table-column label="Name" field="name" sortable>
          <a :href="'/chains/' + props.row.chain_name">
            {{ props.row.chain_name }}
          </a>
        </b-table-column>
        <b-table-column label="Display name" field="chain_name" sortable>
          {{ props.row.display_name }}
        </b-table-column>
        <b-table-column label="chain_id" field="chain_id" sortable>
          {{ props.row.node_info.chain_id }} 
        </b-table-column>
        <b-table-column label="enabled">
          <small
            class="has-text-grey is-abbr-like"
            :title="props.row.enabled"
            >{{ props.row.enabled }}</small
          >
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
  </div>
</template>

<script>
import axios from "~/plugins/axios";

export default {
  name: "ChainsTable",
  props: {
    checkable: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      chains: [
        {
          enabled: true,
          chain_name: "cosmos-hub",
          logo: "https://storage.googleapis.com/emeris/logos/atom.svg",
          display_name: "Cosmos Hub Emeris",
          primary_channel: { cn1: "cn1", cn2: "cn2" },
          denoms: [
            {
              name: "uatom",
              display_name: "ATOM",
              logo: "https://storage.googleapis.com/emeris/logos/atom.svg",
              precision: 6,
              verified: true,
              stakable: true,
              ticker: "ATOM",
              fee_token: true,
              gas_price_levels: { low: 0.01, average: 0.022, high: 0.042 },
              fetch_price: true,
              relayer_denom: true,
              minimum_thresh_relayer_balance: 42
            }
          ],
          demeris_addresses: ["feeaddress"],
          genesis_hash: "genesis_hash",
          node_info: {
            endpoint: "cosmos-hub",
            chain_id: "cosmos-hub-testnet",
            bech32_config: {
              main_prefix: "cosmos",
              prefix_account: "cosmos",
              prefix_validator: "val",
              prefix_consensus: "cons",
              prefix_public: "pub",
              prefix_operator: "oper",
              acc_addr: "cosmos",
              acc_pub: "cosmospub",
              val_addr: "cosmosvaloper",
              val_pub: "cosmosvaloperpub",
              cons_addr: "cosmosvalcons",
              cons_pub: "cosmosvalconspub"
            }
          },
          valid_block_thresh: "10s",
          derivation_path: "m/44'/118'/0'/0/0"
        }
      ],
      isLoading: false,
      paginated: false,
      perPage: 10,
      checkedRows: []
    };
  },
  async mounted() {
    let res = await axios.get("/chains");
    this.chains = res.data.chains;
    console.log(res);
  }
};
</script>
