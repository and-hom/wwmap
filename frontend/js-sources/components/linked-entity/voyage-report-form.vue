<template>
  <div class="wwmap-settings">
    <label for="title">Заголовок</label><input id="title" v-model="report.title"/>
    <label for="author">Автор</label><input id="author" v-model="report.author"/>
    <label for="source">Источник</label><input id="source" v-model="report.source" readonly="readonly"/>
    <label for="remote_id">ID в источнике</label><input id="remote_id" v-model="report.remote_id"/>
    <label for="url">Ссылка</label><input id="url" v-model="report.url"/>
    <label for="tags">Теги</label>
    <span style="width:300px">
                            <vue-tags-input
                                id="tags"
                                v-model="currentTag"
                                :tags="current_tags"
                                :autocomplete-items="autocomplete_tags"
                                @tags-changed="newTags => current_tags = newTags"/>
                        </span>
    <label for="date_of_trip">Дата похода</label>
    <datepicker id="date_of_trip"
                format="yyyy-MM-dd"
                placeholder="YYYY-MM-DD"
                v-model="report.date_of_trip"
                :clear-button="true"></datepicker>
    <label for="date_published">Дата публикации</label>
    <datepicker id="date_published"
                format="yyyy-MM-dd"
                placeholder="YYYY-MM-DD"
                v-model="report.date_published"
                :clear-button="true"></datepicker>
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
    doGetJson(backendApiBase + "/voyage -report", false).then(reports => {
      this.tags = Array.from(new Set(reports.flatMap(t => t.tags)));
    })
  },
  computed: {
    report: {
      get() {
        return this.value
      },
      set(t) {
        this.$emit('input', t);
      }
    },
    autocomplete_tags: {
      get: function () {
        return this.tags.filter(s => s).map(s => {
          return {
            "text": s,
          }
        })
      }
    },
    current_tags: {
      get: function () {
        if (this.report.tags) {
          return this.report.tags.filter(s => s).map(s => {
            return {
              "text": s,
            }
          })
        } else {
          return [];
        }
      },
      set: function (val) {
        this.report.tags = val.map(v => v.text);
      },
    },
  },
  data() {
    return {
      tags: [],
      currentTag: '',
    }
  },
}
</script>