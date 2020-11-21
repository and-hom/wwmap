<template>
  <v-select v-model="river"
            label="title"
            :multiple="multiselect"
            :filterable="true"
            :options="options">
    <template slot="no-options">
      Начните печатать название реки
    </template>
    <template slot="option" slot-scope="option">
      <div class="d-center">
        {{ option.title }}
      </div>
    </template>
    <template slot="selected-option" slot-scope="option">
      <div class="selected d-center">
        {{ option.title }}
      </div>
    </template>
  </v-select>
</template>

<script>
import {doGetJson} from "../../api";
import {backendApiBase} from "../../config";
import {selectModelToValue, valueToSelectModel} from "../../multiselect-utils";

module.exports = {
  props: {
    value: {
      required: true,
    },
    multiselect: {
      type: Boolean,
      default: false,
    },
    bindId: {
      type: Boolean,
      default: false,
    },
  },
  created() {
    var component = this;
    doGetJson(backendApiBase + '/river').then(json => {
      component.rivers = json;
      component.options = json;
    }, err => {
      console.error(err);
      component.rivers = [];
      component.options = [];
    });
  },
  computed: {
    river: {
      get() {
        return valueToSelectModel(this.value, this.options, this.multiselect, this.bindId);
      },
      set(selected) {
        this.$emit('input', selectModelToValue(selected, this.multiselect, this.bindId));
      }
    }
  },
  data() {
    return {
      options: [],
      rivers: [],
    }
  },
}
</script>
