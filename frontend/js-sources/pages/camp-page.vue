<template>
  <page link="camp.htm">
    <entity-grid title="Стоянки"
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
import {backendApiBase} from "../config";
import {createMapParamsStorage} from "wwmap-js-commons/map-settings";

export default {
  data() {
    return {
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
