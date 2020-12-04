<template>
  <ul class="ti-tags">
    <li v-for="tag in tags" class="ti-tag">
      <div class="ti-content">
        <div class="ti-tag-center"><span class="">{{ tag }}</span></div>
      </div>
    </li>
  </ul>
</template>

<script>

module.exports = {
  props: {
    tags: {
      type: Array,
      required: true,
    }
  },
  data() {
    return {}
  }
}
</script>
