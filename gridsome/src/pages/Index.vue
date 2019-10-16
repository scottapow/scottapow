<template>
  <Layout :show-logo="false">
    <!-- Author intro -->
    <Author :show-title="false" />

    <!-- List posts -->
    <h1 class="title">Hi and welcome. I'm Scott.</h1>
    <h2 class="subtitle">...a fast and meticulous web dev in Louisville, KY</h2>
    <p class="copy">Honestly, though, we all think we're great. So, <a href="mailto:lfh.scot@gmail.com">get in touch</a> to discern for yourself.</p>
  </Layout>
</template>

<script>
import Author from '~/components/Author.vue';

export default {
  components: {
    Author,
  },
  metaInfo: {
    title: 'Scott Powell'
  }
}
</script>

<style lang="scss" scoped>
  .title { 
    text-align: center; 
  }
  .subtitle {
    text-align: center;
  }
  .copy { 
    max-width: 400px;
    text-align: center; 
    margin: 0 auto 1em;
  }
</style>
