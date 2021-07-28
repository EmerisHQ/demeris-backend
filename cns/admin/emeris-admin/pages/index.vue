<template>
  <div>
    <title-bar :title-stack="titleStack" />
    <hero-bar> Dashboard </hero-bar>
    <section class="section is-main-section">
      <tiles>
        <card-widget
          class="tile is-child"
          type="is-primary"
          icon="account-multiple"
          :number="2125268"
          label="Chains"
        />
        <card-widget
          class="tile is-child"
          type="is-info"
          icon="cart-outline"
          :number="5112369420"
          prefix="$"
          label="GDEX TVL"
        />
        <card-widget
          class="tile is-child"
          type="is-success"
          icon="chart-timeline-variant"
          :number="256785653"
          suffix=""
          label="IBC transfers"
        />
      </tiles>

      <card-component title="Chains" class="has-table has-mobile-sort-spaced">
        <chains-table />
      </card-component>
    </section>
  </div>
</template>

<script>
import TitleBar from '@/components/TitleBar'
import HeroBar from '@/components/HeroBar'
import Tiles from '@/components/Tiles'
import CardWidget from '@/components/CardWidget'
import CardComponent from '@/components/CardComponent'
import ChainsTable from '@/components/ChainsTable'
export default {
  name: 'Home',
  components: {
    ChainsTable,
    CardComponent,
    CardWidget,
    Tiles,
    HeroBar,
    TitleBar,
  },
  computed: {
    titleStack() {
      return ['Admin', 'Dashboard']
    },
  },
  head() {
    return {
      title: 'Dashboard',
    }
  },
}
</script>
