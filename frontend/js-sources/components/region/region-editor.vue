<template>
  <div>
    <btn-bar ref="btnBar" logObjectType="REGION" :logObjectId="region.id">
      <button type="button" class="btn btn-success"
              v-on:click="save()">Сохранить
      </button>
      <slot></slot>
    </btn-bar>
    <div class="spot-editor-panel" style="padding-top:15px;">
      <label for="region_title" style="font-weight:bold; margin-top:5px">Название:</label>
      <input v-model.trim="region.title" style="display:block;" id="region_title"/>
      <dl style="margin-top:10px;">
        <dt>Страна:</dt>
        <dd>
          <select v-model="region.country_id">
            <option v-for="country in countries" v-bind:value="country.id">{{ country.title }}</option>
          </select>
        </dd>
      </dl>

    </div>
  </div>
</template>

<script>
import {store} from '../../app-state';
import {hasRole, ROLE_ADMIN} from '../../auth';
import {getCountries, saveRegion, setActiveEntityUrlHash} from "../../editor";

var $ = require("jquery");
require("jquery.cookie");

module.exports = {
  props: ['region', 'country'],
  created: function () {
    getCountries().then(countries => this.countries = countries);
  },
  mounted: function () {
    hasRole(ROLE_ADMIN).then(canEdit => this.canEdit = canEdit);
  },
  computed: {
    editMode: {
      get: function () {
        return store.state.spoteditorstate.editMode
      },

      set: function (newVal) {
        store.commit("setSpotEditorEditMode", newVal);
      }
    },
  },
  data: function () {
    return {
      // for editor
      canEdit: false,
      countries: [],
      prevCountryId: this.region.country_id,
      save: function () {
        if (!this.region.title || !this.region.title.replace(/\s/g, '').length) {
          this.showError("Нельзя сохранять регион без названия");
          return
        }

        if (!this.region.country_id) {
          this.showError("Нельзя сохранять регион без страны");
          return
        }

        this.$refs.btnBar.disable()
        saveRegion(this.region).then(updated => {
          this.hideError();

          setActiveEntityUrlHash(updated.country_id, updated.id);
          store.commit('setTreeSelection', {
            countryId: updated.country_id,
            regionId: updated.id,
            riverId: null,
            spotId: null
          });
          store.dispatch('reloadCountrySubentities', updated.country_id);

          this.editMode = false;

        }, _ => {
          this.showError("Не удалось сохранить регион. Возможно, недостаточно прав");
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