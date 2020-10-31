<template>
  <page link="camp.htm">
    <entity-grid v-model="camps"
                 title="Стоянки"
                 :url-base="backendApiBase + '/camp'"
                 :fields='[
                     {"label": "Название",          "name":"title"},
                     {"label": "Место",             "name":"point",                "type": "location"},
                     {"label": "Описание",          "name":"description"},
                     {"label": "Реки",              "name":"rivers_data",          "type": "rivers"},
                     {"label": "Мест под палатку",  "name":"num_tent_places",      "type": "natural_number"},
                 ]'
                 :blank-entity-factory="createBlankCamp">
      <template v-slot:form="slotProps">
        <camp-form v-model="slotProps.entity"/>
      </template>
    </entity-grid>
  </page>
</template>

<script>
import {backendApiBase, frontendBase} from "../config";
import {createMapParamsStorage} from "wwmap-js-commons/map-settings";

const moment = require('moment');
const VueScrollTo = require('vue-scrollto');

export default {
  data() {
    return {
      camps: [],
      canEdit: false,
      campForEdit: {},
      backendApiBase: backendApiBase,
      mapParamsStorage: createMapParamsStorage(),
    }
  },
  methods: {
    createBlankCamp: function () {
      return {
        title: "",
        description: "",
        point: this.mapParamsStorage.getLastPositionZoomType().position,
        rivers: [],
      }
    },
  }
}
</script>
