<template>
  <div class="card">
    <header v-if="title" class="card-header">
      <p class="card-header-title">
        {{ title }}
      </p>
    </header>
    <div class="card-content">
      <slot />
    </div>
  </div>
</template>

<script>
export default {
  name: 'CardComponent',
  props: {
    title: {
      type: String,
      default: null,
    },
  },
}
</script>
