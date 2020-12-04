<template>
  <page link="transfer.htm">
    <entity-grid title="Заброски"
                 :url-base="backendApiBase + '/transfer'"
                 :fields='[
                     {"label": "Название",          "name":"title"},
                     {"label": "Описание",          "name":"description"},
                     {"label": "Откуда",            "name":"stations",          "type": "tags"},
                     {"label": "Реки",              "name":"rivers_data",       "type": "rivers"},
                 ]'
                 :blank-entity-factory="createBlankTransfer">
      <template v-slot:form="slotProps">
        <transfer-form v-model="slotProps.entity"/>
      </template>
    </entity-grid>
  </page>
</template>


<script>
import {backendApiBase} from "../config";

export default {
  data() {
    return {
      backendApiBase: backendApiBase,
    }
  },
  methods: {
    createBlankTransfer() {
      return {
        title: "",
        stations: [],
        description: "",
        rivers: [],
      };
    },
  }
}
</script>
