<template>
  <div style="margin-left: 15px; margin-right: 15px;">
    <div v-if="canEdit">
      <h2 style="display: inline;">{{ title }}</h2>
      <button data-toggle="modal" data-target="#add" style="margin-left:30px;"
              v-on:click="entityForEdit=blankEntityFactory()">+
      </button>
      <create-entity :entity="entityForEdit"
                     :urlBase="urlBase"
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
           :filter-function="filterData"
           :page-size="10">
      <template v-slot:filter="slotProps">
        <river-select v-model="riverFilter" :multiselect="true" :bindId="true"/>
      </template>
      <template v-slot:default="slotProps">
        <table class="table">
          <thead>
          <tr>
            <th v-for="field in fields">
              {{ field.label }}
            </th>
            <th v-if="canEdit" class="btn-col"></th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="entity in slotProps.data">
            <td v-for="field in fields" class="fitwidth">
              <river-links v-if="field.type=='rivers'" :rivers="entity[field.name]"/>
              <span v-else>{{ entity[field.name] }}</span>
            </td>
            <td v-if="canEdit" class="btn-col">
              <button v-on:click="entityForEdit={ ...entity }"
                      data-toggle="modal" data-target="#add">Правка
              </button>
              <ask :id="'del-'+entity.id" title="Удалить?"
                   msg="Отменить удаление будет нельзя"
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
</style>

<script>

import {doDelete, doGetJson} from "../../api";
import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from "../../auth";
import {store} from "../../app-state";
import {arrays_intersects} from "wwmap-js-commons/util";

module.exports = {
  props: {
    fields: {
      type: Array,
      required: true,
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
      required: true,
    }
  },
  computed: {
    errMsg() {
      return store.state.errMsg
    },
  },
  created: function () {
    this.refresh();
    hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
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
      doGetJson(`${this.urlBase}?rivers=true`, false).then(entities => {
        this.entities = entities;
      })
    },
    showError: function (errMsg) {
      store.commit("setErrMsg", errMsg);
    },
    hideError: function () {
      store.commit("setErrMsg", null);
    },
    filterData(data, filter) {
      if(data && filter && filter.length > 0) {
        return data.filter(d => d.rivers && arrays_intersects(d.rivers, filter))
      } else {
        return data
      }
    },
  },
}
</script>