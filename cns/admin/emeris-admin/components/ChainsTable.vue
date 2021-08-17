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
          <router-link :to="'/chains/' + props.row.chain_name">
            {{ props.row.chain_name }}
          </router-link>
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
            <p>Fetching chains...</p>
          </template>
          <template v-else>
            <p>
              <b-icon icon="emoticon-sad" size="is-large" />
            </p>
            <p>No chains found&hellip;</p>
          </template>
        </div>
      </section>
    </b-table>
  </div>
</template>

<script>
import axios from "~/plugins/axios";
import { mapGetters, mapMutations } from 'vuex'

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
      isLoading: false,
      paginated: false,
      perPage: 10,
      checkedRows: []
    };
  },
  methods: {
    ...mapMutations([
      'updateChains'
    ])
  },
  computed: {
    chains() { return this.$store.state.chains }
  },
  async mounted() {
    this.isLoading = true;
    this.updateChains()
    this.isLoading = false;
  }
};
</script>
