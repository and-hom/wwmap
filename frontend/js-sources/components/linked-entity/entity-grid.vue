<template>
  <div>
    <div v-if="canEdit">
      <h2 style="display: inline;">{{ title }}</h2>
      <button data-toggle="modal" data-target="#add" style="margin-left:30px;">+</button>
      <create-entity :entity="forEdit"
                     :urlBase="urlBase"
                     :okFn="editOk"
                     :cancelFn="resetEntityForEdit">
        <template v-slot:form="slotProps">
          <slot name="form" v-bind:entity="slotProps.entity">
          </slot>
        </template>
      </create-entity>
    </div>
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
      <tr v-for="entity in entities">
        <td v-for="field in fields" class="fitwidth">
          <river-links v-if="field.type=='rivers'" :rivers="entity[field.name]"/>
          <span v-else>{{ entity[field.name] }}</span>
        </td>
        <td v-if="canEdit" class="btn-col">
          <button v-on:click="forEdit={ ...entity }"
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
  created: function () {
    this.refresh();
    hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
  },
  data() {
    return {
      entities: [],
      forEdit: {},
      canEdit: false,
    }
  },
  methods: {
    remove: function (id) {
      doDelete(urlBase + id, true).then(this.refresh)
    },
    editOk: function () {
      this.refresh();
      this.forEdit = this.blankEntityFactory();
    },
    resetEntityForEdit: function () {
      this.forEdit = this.blankEntityFactory();
    },
    refresh: function () {
      doGetJson(this.urlBase + "?rivers=true", false).then(entities => {
        this.entities = entities;
      })
    },
  },
}
</script>