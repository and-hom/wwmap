<template>
  <div>
    <btn-bar ref="btnBar" logObjectType="COUNTRY" :logObjectId="country.id">
      <button type="button" class="btn btn-success"
              v-on:click="save()">Сохранить
      </button>
      <slot></slot>
    </btn-bar>
    <div class="spot-editor-panel" style="padding-top:15px;">
      <label for="country_title" style="font-weight:bold; margin-top:5px">Название:</label>
      <input v-model.trim="country.title" style="display:block;" id="country_title"/>
      <label for="country_code" style="font-weight:bold; margin-top:5px">
        Двухбуквенный код <a href="https://ru.wikipedia.org/wiki/%D0%9E%D0%B1%D1%89%D0%B5%D1%80%D0%BE%D1%81%D1%81%D0%B8%D0%B9%D1%81%D0%BA%D0%B8%D0%B9_%D0%BA%D0%BB%D0%B0%D1%81%D1%81%D0%B8%D1%84%D0%B8%D0%BA%D0%B0%D1%82%D0%BE%D1%80_%D1%81%D1%82%D1%80%D0%B0%D0%BD_%D0%BC%D0%B8%D1%80%D0%B0" target="_blank">ОКСМ</a>:
      </label>
      <input v-model.trim="country.code" style="display:block;" id="country_code"/>
    </div>
  </div>
</template>

<script>
import {store} from '../../app-state';
import {hasRole, ROLE_ADMIN} from '../../auth';
import {setActiveEntityUrlHash, saveCountry} from "../../editor";

var $ = require("jquery");
require("jquery.cookie");

module.exports = {
  props: ['country'],
  mounted: function () {
    hasRole(ROLE_ADMIN).then(canEdit => this.canEdit = canEdit);
  },
  computed: {
    editMode: {
      get: function () {
        return store.state.countryeditorstate.editMode
      },

      set: function (newVal) {
        store.commit("setCountryEditorEditMode", newVal);
      }
    },
  },
  data: function () {
    return {
      // for editor
      canEdit: false,
      save: function () {
        if (!this.country.title || !this.country.title.replace(/\s/g, '').length) {
          this.showError("Нельзя сохранять страну без названия");
          return
        }

        this.$refs.btnBar.disable()
        saveCountry(this.country).then(updated => {
          this.hideError();

          setActiveEntityUrlHash(updated.id);
          store.commit('setTreeSelection', {
            countryId: updated.id,
            regionId: null,
            riverId: null,
            spotId: null
          });
          store.dispatch('reloadCountries');

          this.editMode = false;
        }).catch(e => {
          if (e.status == 409) {
            this.showError("Не удалось сохранить страну. Дубликат");
          } else {
            this.showError("Не удалось сохранить страну. Возможно, недостаточно прав");
          }
        }).finally(() => {
          if (this.$refs.btnBar) {
            this.$refs.btnBar.enable()
          }
        });
      },

      showError: function (errMsg) {
        store.commit("setErrMsg", errMsg);
      },
      hideError: function () {
        store.commit("setErrMsg", null);
      },
    }
  }
}

</script>
