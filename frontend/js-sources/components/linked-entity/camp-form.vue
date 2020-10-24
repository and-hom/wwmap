<template>
  <div class="wwmap-settings">
    <label for="title">Название</label><input id="title" v-model="camp.title"
                                              v-on:change="$emit('input', camp)"/>
    <label for="description">Описание</label><textarea id="description"
                                                       v-model="camp.description"
                                                       style="width:300px; height:300px;"></textarea>
    <label for="num_tent_places">Кол-во мест под палатку</label><input id="num_tent_places"
                                                                       type="number"
                                                                       min="0"
                                                                       oninput="validity.valid||(value='');"
                                                                       v-model.number="camp.num_tent_places"/>

  </div>
</template>

<script>

module.exports = {
  props: {
    value: {
      type: Object,
      required: true,
    },
  },
  computed: {
    camp: {
      get() {
        return this.value
      },
      set(t) {
        this.$emit('input', t);
      }
    },
  },
  data() {
    return {}
  },
}
</script>