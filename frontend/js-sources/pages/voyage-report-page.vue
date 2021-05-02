<template>
  <page link="reports.htm">
    <entity-grid title="Отчёты"
                 :url-base="backendApiBase + '/voyage-report'"
                 :fields='[
                     {"label": "Автор",             "name":"author"},
                     {"label": "Источник",          "name":"source"},
                     {"label": "ID в источнике",    "name":"remote_id"},
                     {"label": "Ссылка",            "name":"url","text":"title","type": "link"},
                     {"label": "Теги",              "name":"tags",              "type": "tags"},
                     {"label": "Дата похода",       "name":"date_of_trip",      "type": "date"},
                     {"label": "Дата публикации",   "name":"date_published",    "type": "date"},
                     {"label": "Реки",              "name":"rivers_data",       "type": "rivers"},
                 ]'
                 removed-field="removed"
                 :blank-entity-factory="createBlankReport">
      <template v-slot:form="slotProps">
        <voyage-report-form v-model="slotProps.entity"/>
      </template>
    </entity-grid>
  </page>
</template>


<script>
import {backendApiBase} from "../config";
import {randomStringId} from "wwmap-js-commons/util";

export default {
  data() {
    return {
      backendApiBase: backendApiBase,
    }
  },
  methods: {
    createBlankReport() {
      return {
        source: "MANUAL",
        remote_id: randomStringId(32),
        tags: [],
        rivers: [],
      };
    },
  }
}
</script>
