<template>
  <page link="transfer.htm">
    <div v-if="canEdit">
      <create-entity :transfer="transferForEdit"
                     :stations="stations"
                     :okFn="editOk"
                     :cancelFn="resetTransferForEdit">
        <template v-slot:form="slotProps">
          <transfer-form v-model="slotProps.entity"/>
        </template>
      </create-entity>
      <h2 style="display: inline;">Трансферы</h2>
      <button data-toggle="modal" data-target="#add-transfer" style="margin-left:30px;">+</button>
    </div>
    <table class="table">
      <thead>
      <tr>
        <th width="200px">Название</th>
        <th width="250px">Откуда</th>
        <th width="250px">Реки</th>
        <th>Контакты / Описание</th>
        <th v-if="canEdit"></th>
      </tr>
      </thead>
      <tbody>
      <transfer-row v-for="transfer in transfers"
                    :key="transfer.id"
                    :transfer="transfer"
                    :can-edit="canEdit"
                    :selected="`${transfer.id}` == selected">
        <td v-if="canEdit">
          <button v-on:click="transferForEdit={ ...transfer }"
                  data-toggle="modal" data-target="#add-transfer">Правка
          </button>
          <ask :id="'del-transfer-'+transfer.id" title="Удалить задачу?"
               msg="Вся история и логи будут также удалены. Если удаление всех связанных данных не требуется, лучше отредактировать задачу, сняв флаг enabled"
               :ok-fn="function() {
                          remove(transfer.id)
                        }"></ask>
          <button data-toggle="modal" :data-target="'#del-transfer-'+transfer.id">Удалить</button>
        </td>
      </transfer-row>
      </tbody>
    </table>
  </page>
</template>


<script>
import {doDelete, doGetJson} from "../api";
import {backendApiBase} from "../config";
import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from "../auth";
import Vue from "vue";

const moment = require('moment');
const VueScrollTo = require('vue-scrollto');

export default {
  created: function () {
    Vue.use(VueScrollTo)
    this.refresh();
    hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
  },
  mounted: function () {
  },
  beforeDestroy: function () {

  },
  updated: function () {
  },
  data() {
    return {
      transfers: [],
      canEdit: false,
      transferForEdit: {},
      stations: [],
      selected: window.location.hash ? window.location.hash.substr(1) : '',
      scrolled: false,
    }
  },
  methods: {
    refresh: function () {
      doGetJson(backendApiBase + "/transfer", false).then(transfers => {
        this.transfers = transfers;
        this.stations = Array.from(new Set(transfers.flatMap(t => t.stations)));
      }).then(this.scrollToActive)

    },
    scrollToActive: function () {
      if (this.selected) {
        let t = this;
        this.$nextTick().then(_ => {
          let id = `row${t.selected}`;
          let elt = document.getElementById(id);
          if (elt) {
            VueScrollTo.scrollTo(`#row${t.selected}`, 1000, {});
          } else {
            console.log("Can't scroll to active")
            setTimeout(t.scrollToActive, 100);
          }
        });
      }
    },

    resetTransferForEdit: function () {
      this.transferForEdit = {
        title: "",
        stations: [],
        description: "",
        rivers: [],
      }
    },
    editOk: function () {
      this.refresh();
      this.resetTransferForEdit();
    },
    remove: function (id) {
      doDelete(backendApiBase + "/transfer/" + id, true).then(this.refresh)
    },
  }
}
</script>
