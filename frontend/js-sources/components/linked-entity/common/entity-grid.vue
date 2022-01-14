<template>
  <div style="margin-left: 15px; margin-right: 15px;">
    <div v-if="canEdit">
      <h2 style="display: inline;">{{ title }}</h2>
      <button data-toggle="modal" data-target="#add" style="margin-left:30px;"
              v-on:click="entityForEdit=blankEntityFactory()">+
      </button>
      <create-entity :entity="entityForEdit"
                     :urlBase="urlBase"
                     :has-map="hasMap"
                     :okFn="editOk"
                     :failFn="editFail"
                     :cancelFn="editCancel">
        <template v-slot:form="slotProps">
          <slot name="form" v-bind:entity="slotProps.entity">
          </slot>
        </template>
      </create-entity>
    </div>
    <transition name="fade">
      <div class="alert alert-danger" role="alert" v-if="errMsg">
        {{ errMsg }}
      </div>
    </transition>
    <pager :data="entities"
           :filter="riverFilter"
           :custom-filter="customFilter"
           :filter-function="filterData"
           :page-size="10">
      <template v-slot:filter="slotProps">
        <slot name="filter" ></slot>
        <river-select
            v-model="riverFilter"
            :multiselect="true"
            :bindId="true"
            style="width: 400px; display: inline; float: left;"
        />
      </template>
      <template v-slot:default="slotProps">
        <table class="table">
          <thead>
          <tr>
            <th v-for="field in fields" :class="colClass(field.type)">
              {{ field.label }}
            </th>
            <th v-if="canEdit" class="btn-col"></th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="entity in slotProps.data">
            <td v-for="field in fields" class="fitwidth">
              <river-links v-if="field.type=='rivers'" :rivers="entity[field.name]" style="width: 200px"/>
              <tags v-else-if="field.type=='tags'" :tags="entity[field.name]"/>
              <location v-else-if="field.type=='location'" :point-or-line="entity[field.name]"/>
              <a v-else-if="field.type=='link'" target="_blank" :href="entity[field.name]">{{ field.text && entity[field.text] ? entity[field.text] : entity[field.name] }}</a>
              <span v-else-if="field.type=='date'">{{ entity[field.name] | formatDateStr | emptyDate00010101 | orElse('&mdash;') | yearOnly0102}}</span>
              <span v-else-if="field.type=='dateTime'">{{ entity[field.name] | formatDateTimeStr | emptyDate00010101 | orElse('&mdash;') | yearOnly0102}}</span>
              <span v-else>{{ entity[field.name] }}</span>
            </td>
            <td v-if="canEdit && removedField && entity[removedField]===true" class="btn-col">
              <ask :id="'undo-del-'+entity.id" title="Отменить удаление?"
                   msg="Отменить удаление"
                   :ok-fn="function() {undoRemove(entity.id)}"></ask>
              <button data-toggle="modal" :data-target="'#undo-del-'+entity.id">Отменить удаление</button>
            </td>
            <td v-else-if="canEdit" class="btn-col">
              <button v-on:click="entityForEdit={ ...entity }"
                      data-toggle="modal" data-target="#add">Правка
              </button>
              <ask :id="'del-'+entity.id" title="Удалить?"
                   msg="Можно будет отменить удаление"
                   :ok-fn="function() {remove(entity.id)}"></ask>
              <button data-toggle="modal" :data-target="'#del-'+entity.id">Удалить</button>
            </td>
          </tr>
          </tbody>
        </table>
      </template>
    </pager>
  </div>
</template>

<style>
td.fitwidth {
  width: 1px;
  /*white-space: nowrap;*/
}

.table {
  width: 100%;
}

.btn-col {
  width: 200px;
}

.rivers-col {
  /*width: 200px;*/
}

.tags-col {

}

.default-col {

}
</style>

<script>

import {doDelete, doGetJson, doPost} from "../../../api";
import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from "../../../auth";
import {store} from "../../../app-state";
import {arrays_intersects} from "wwmap-js-commons/util";

module.exports = {
  props: {
    fields: {
      type: Array,
      required: true,
    },
    editable: {
      type: Boolean,
      default: true,
    },
    urlBase: {
      type: String,
      required: true,
    },
    title: {
      type: String,
      required: true,
    },
    blankEntityFactory: {
      type: Function,
      default: function () {
        return {};
      },
    },
    removedField: {
      type: String,
      required: false,
    },
    customFilterFunction: {
      type: Function,
      required: false,
    },
    customFilter: {
      type: Object,
      required: false,
    },
  },
  computed: {
    errMsg() {
      return store.state.errMsg
    },
    hasMap() {
      return this.fields.filter(f => f.type == 'location' && f.name == 'point').length > 0
    },
  },
  created: function () {
    this.refresh();
    if(this.editable) {
      hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
    }
  },
  data() {
    return {
      entities: [],
      forEdit: {},
      canEdit: false,
      entityForEdit: this.blankEntityFactory(),
      riverFilter: [],
    }
  },
  methods: {
    remove: function (id) {
      this.hideError();
      doDelete(`${this.urlBase}/${id}`, true).then(this.refresh)
          .catch(err => this.showError(err));
    },
    undoRemove: function (id) {
      this.hideError();
      doPost(`${this.urlBase}/${id}/undo-delete`, true, true).then(this.refresh)
          .catch(err => this.showError(err));
    },
    editOk: function () {
      this.hideError();
      this.refresh();
      this.entityForEdit = this.blankEntityFactory();
    },
    editFail: function (err) {
      this.showError(err)
    },
    editCancel: function () {
      this.entityForEdit = this.blankEntityFactory();
    },
    refresh: function () {
      doGetJson(`${this.urlBase}?rivers=true`, true).then(entities => {
        this.entities = entities;
      })
    },
    showError: function (errMsg) {
      store.commit("setErrMsg", errMsg);
    },
    hideError: function () {
      store.commit("setErrMsg", null);
    },
    filterData(data, filter, customFilter) {
      if (data) {
        return data.filter(d =>
            (filter && filter.length > 0 && d.rivers ? arrays_intersects(d.rivers, filter) : true) &&
            (customFilter && this.customFilterFunction ? this.customFilterFunction(d, customFilter) : true)
        )
      } else {
        return data
      }
    },
    colClass(fieldType) {
      switch (fieldType) {
        case 'rivers':
          return 'rivers-col';
        case 'tags':
          return 'tags-col';
        default:
          return 'default-col';
      }
    },
  },
}
</script>
