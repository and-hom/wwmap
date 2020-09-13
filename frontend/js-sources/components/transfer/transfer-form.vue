<template>
  <div class="settings">
    <label for="task_title">Название</label><input id="task_title" v-model="transfer.title" v-on:change="$emit('input', transfer)"/>
    <label for="stations">Откуда</label>
    <span style="width:300px">
                            <vue-tags-input
                                id="stations"
                                v-model="currentTag"
                                :tags="current_stations"
                                :autocomplete-items="autocomplete_stations"
                                @tags-changed="newTags => current_stations = newTags"/>
                        </span>
    <label for="description">Описание</label><textarea id="description"
                                                       v-model="transfer.description"
                                                       style="width:300px; height:300px;"></textarea>

  </div>
</template>
<script>

import {doGetJson} from "../../api";
import {backendApiBase} from "../../config";

module.exports = {
  props: {
    value: {
      type: Object,
      required: true,
    },
  },
  created() {
    doGetJson(backendApiBase + "/transfer-full", false).then(transfers => {
      this.stations = Array.from(new Set(transfers.flatMap(t => t.stations)));
    })
  },
  computed: {
    transfer: {
      get() {
        return this.value
      },
      set(t) {
        this.$emit('input', t);
      }
    },
    autocomplete_stations: {
      get: function () {
        return this.stations.filter(s => s).map(s => {
          return {
            "text": s,
          }
        })
      }
    },
    current_stations: {
      get: function () {
        if (this.transfer.stations) {
          return this.transfer.stations.filter(s => s).map(s => {
            return {
              "text": s,
            }
          })
        } else {
          return [];
        }
      },
      set: function (val) {
        this.transfer.stations = val.map(v => v.text);
      },
    },
  },
  data() {
    return {
      stations: [],
      currentTag: '',
    }
  },
}
</script>