<template>
  <div>
    <title-bar :title-stack="titleStack" />
    <hero-bar>
      Chains
    </hero-bar>
    <section class="section is-main-section">
      <card-component title="Chains" class="has-table has-mobile-sort-spaced">
        <chains-table />
      </card-component>
      <hr />
    </section>
  </div>
</template>

<script>
import ChainsTable from '@/components/ChainsTable'
import CardComponent from '@/components/CardComponent'
import TitleBar from '@/components/TitleBar'
import HeroBar from '@/components/HeroBar'
export default {
  name: 'Chains',
  components: {
    HeroBar,
    TitleBar,
    CardComponent,
    ChainsTable,
  },
  computed: {
    titleStack() {
      return ['Admin', 'Chains']
    },
  },
  head() {
    return {
      title: 'Chains',
    }
  },
}
</script>
