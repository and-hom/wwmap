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
                 :custom-filter="sourceFilter"
                 :custom-filter-function="filterBySource"
                 :blank-entity-factory="createBlankReport">
      <template v-slot:form="slotProps">
        <voyage-report-form v-model="slotProps.entity"/>
      </template>
      <template v-slot:filter>
        <v-select
            v-model="sourceFilter"
            label="title"
            :placeholder="reportSourceSelectorPlaceholder"
            :options="reportSources"
            :multiselect="false"
            style="width: 400px; display: inline; float: left;"
        />
      </template>
    </entity-grid>
  </page>
</template>


<script>
import {backendApiBase} from "../config";
import {randomStringId} from "wwmap-js-commons/util";
import {doGetJson} from "../api";

export default {
  data() {
    return {
      backendApiBase: backendApiBase,
      sourceFilter: null,
      reportSources: [],
      reportSourceSelectorPlaceholder: "―",
      filterBySource: function (element, source) {
        return source == null || source.id == null || element.source == source.id;
      }
    }
  },
  created() {
    doGetJson(backendApiBase+'/report-sources').then(s => {
      this.reportSources = s;
      this.reportSources.unshift({
        "id": null,
        "title": this.reportSourceSelectorPlaceholder,
      })
    });
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
