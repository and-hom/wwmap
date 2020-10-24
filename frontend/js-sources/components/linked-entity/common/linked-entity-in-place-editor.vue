<template>
  <div>
    <div v-if="editMode">
      <div style="padding-top:15px;">
        <div v-if="map" style="margin-bottom: 20px">
          <ya-map-location v-bind:spot="newEntity"
                           width="100%"
                           height="600px"
                           :editable="true"
                           :ya-search="true"/>
        </div>
        <slot name="form" v-bind:entity="newEntity">
          <label style="padding-right: 10px;" for="title_input"><strong>Название:</strong></label><input
            id="title_input" type="text" v-model="newEntity.title" style="margin-top: 10px; width: 80%;"/>
        </slot>
        <div class="btn-toolbar" style="padding-top:15px;">
          <div class="btn-group mr-2" role="group">
            <button type="button"
                    class="btn btn-success"
                    v-on:click.stop="onSaveEntity"
                    :disabled="!newEntity.title && !allowEmptyTitle">{{ saveBtnText() }}
            </button>
            <button type="button" class="btn btn-cancel" v-on:click.stop="editMode = false">Отмена</button>
          </div>
        </div>
      </div>
    </div>
    <div v-else>
      <div class="wwmap-system-hint">Выберите из списка<span v-if="baseUrl && canCreate"> или
        <button v-on:click.stop="editMode=true">создайте</button></span>
      </div>
      <v-select v-model="selectorModel" label="title" :options="entities" :multiple="multiselect">
        <template slot="no-options">
          Начните печатать название
        </template>
        <template slot="option" slot-scope="option">
          {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}
        </template>
        <template slot="selected-option" slot-scope="option">
          {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}<a href="#"
                                                                type="button"
                                                                title="Редактировать"
                                                                aria-label="Редактировать"
                                                                v-on:click.stop="newEntity=option; editMode=true;"
                                                                class="vs__deselect"><img
            src="https://wwmap.ru/img/edit.png"/></a>
        </template>
      </v-select>
    </div>
    <slot v-if="showSelected" name="grid" v-bind:entities="selected">
      <ul>
        <li v-for="entity in selected">{{ entity.id }}&nbsp;&dash;&nbsp;{{ entity.title }}&nbsp;<button
            v-on:click.stop="onUnselectEntity(entity)">[X]
        </button>
        </li>
      </ul>
    </slot>
  </div>
</template>

<script>

import {doGetJson, doPostJson} from "../../../api";
import {valueToSelectModel, selectModelToValue} from "../../../multiselect-utils";

module.exports = {
  props: {
    value: {},
    baseUrl: {
      type: String,
    },
    data: {
      type: Array,
    },
    bindId: { // true - use ids for v-model; false - use objects for v-model
      type: Boolean,
      default: true,
    },
    multiselect: {
      type: Boolean,
      default: false,
    },
    showSelected: {
      type: Boolean,
      default: false,
    },
    authForList: {
      type: Boolean,
      required: false,
    },
    canCreate: {
      type: Boolean,
      default: true,
    },
    map: {
      type: Boolean,
      default: true,
    },
    mapDefaultLocation: {},
    allowEmptyTitle: {
      type: Boolean,
      default: false,
    }
  },
  computed: {
    selected: {
      get() {
        return valueToSelectModel(this.value, this.entities, this.multiselect, this.bindId);
      },
      set(selected) {
        this.$emit('input', selectModelToValue(selected, this.multiselect, this.bindId));
      }
    },
    selectorModel: {
      get() {
        if (this.multiselect) {
          return this.selected;
        } else if (this.selected.length == 0) {
          return null;
        } else {
          return this.selected[0]
        }
      },
      set(value) {
        this.selected = this.multiselect
            ? value
            : [value];
      }
    },
  },
  created() {
    this.loadEntities();
  },
  data: function () {
    return {
      entities: [],
      newEntity: this.newEntityFactory(),
      editMode: false,
    }
  },
  methods: {
    loadEntities: function () {
      return (
          this.baseUrl
              ? doGetJson(this.baseUrl, this.authForList)
              : new Promise((resolve, _) => resolve(this.data))
      ).then(entities => this.entities = entities)
    },

    onSelectEntity: function (entity) {
      if (this.multiselect) {
        this.selected = this.selected.concat(entity);
      } else {
        this.selected = [entity]
      }
    },

    onUnselectEntity: function (entity) {
      this.selected = this.selected.filter(e => e.id != entity.id);
    },

    persist: function () {
      let url = this.newEntity.id
          ? `${this.baseUrl}/${this.newEntity.id}`
          : this.baseUrl;
      return doPostJson(url, this.newEntity, true);
    },

    onSaveEntity: function () {
      if (this.baseUrl) {
        this.persist().then(created => {
          this.loadEntities().then(_ => {
            this.newEntity = {};
            this.onSelectEntity(created);
            this.editMode = false;
          });
        });
      } else {
        this.newEntity = this.newEntityFactory();
        this.onSelectEntity(created);
        this.editMode = false;
      }
    },
    newEntityFactory: function () {
      let pos;
      if (map) {
        if (this.mapDefaultLocation) {
          if (typeof this.mapDefaultLocation == 'function') {
            pos = this.mapDefaultLocation();
          } else {
            pos = this.mapDefaultLocation;
          }
        } else {
          pos = [0, 0]
        }
      } else {
        pos = null;
      }
      return {
        point: pos,
      }
    },
    saveBtnText: function () {
      return this.newEntity.id ? "Сохранить" : "Добавить"
    },
  }
}
</script>