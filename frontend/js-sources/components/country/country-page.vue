<template>
  <div>
    <window-title v-bind:text="country.title"></window-title>
    <ask id="del-country" title="Точно?"
         msg="Совсем удалить? Восстановить будет никак нельзя!"
         :ok-fn="function() { remove(); }"></ask>
    <ask id="del-country-forbidden" title="Нельзя удалить!"
         msg="Нельзя удалить страну, в которой есть реки. Сначала удалите все реки этой страны вручную."
         :okBtn="false" :noBtn="false" :cancelBtn="true"></ask>

    <country-viewer v-if="!editMode"
                    :country="country">
      <button type="button" class="btn btn-primary" v-on:click="add_river()">Добавить реку без региона</button>
      <button v-if="canAddRegion" type="button" class="btn btn-success" v-on:click="add_region()">Добавить регион
      </button>
      <button type="button" class="btn btn-info" v-if="canEditCountry"
              v-on:click="editMode=true; hideError();">Редактирование
      </button>
      <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-country"
              v-if="country.id>0 && !country.has_rivers">
        Удалить
      </button>
      <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-country-forbidden"
              v-if="country.id>0 && country.has_rivers">
        Удалить
      </button>
    </country-viewer>

    <country-editor v-if="canEditCountry && editMode"
                    ref="editor"
                    :country="country">
      <button type="button" class="btn btn-secondary" v-on:click="cancelEditing()">Отменить</button>
      <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-country"
              v-if="country.id>0 && !country.has_rivers">
        Удалить
      </button>
      <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-country-forbidden"
              v-if="country.id>0 && country.has_rivers">
        Удалить
      </button>
    </country-editor>
  </div>
</template>

<script>
import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth'
import {store} from '../../app-state';
import {getCountry, removeCountry, setActiveEntityUrlHash} from "../../editor";

module.exports = {
  props: ['initialCountry'],
  created: function () {
    hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
    hasRole(ROLE_ADMIN).then(canAddRegion => this.canAddRegion = canAddRegion);
    hasRole(ROLE_ADMIN).then(canEditCountry => this.canEditCountry = canEditCountry);
    this.resetToInitialIfRequired();
  },
  updated: function () {
    this.resetToInitialIfRequired();
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
      canAddRegion: false,
      canEditCountry: false,

      country: null,
      previousCountryId: this.initialCountry.id,

      getEditModeButtonTitle: function () {
        return this.editMode ? 'Просмотр' : 'Редактирование';
      },
      // end of editor
      add_river: function () {
        store.commit('newRiver', {
          country: this.country,
          region: {
            id: 0,
            title: this.country.title,
            country_id: this.country.id,
          }
        })
      },
      add_region: function () {
        store.commit('newRegion', {
          country: this.country,
        })
      },
    }
  },

  methods: {
    resetToInitialIfRequired: function () {
      if (this.shouldReInit()) {
        this.previousCountryId = this.initialCountry.id;
        this.country = this.initialCountry;
      }
    },
    shouldReInit: function () {
      return this.country == null ||
          this.previousCountryId !== this.initialCountry.id && this.initialCountry.id > 0
    },
    remove: function () {
      this.hideError();
      removeCountry(this.country.id).then(
          _ => this.closeEditorAndShowParent(),
          err => this.showError("не могу удалить: " + err.statusText || err))
    },
    cancelEditing: function () {
      this.editMode = false;
      if (this.country && this.country.id > 0) {
        this.reload();
      } else {
        this.closeEditorAndShowParent();
      }
    },
    reload: function () {
      getCountry(this.country.id).then(country => {
        this.country = country;
        this.hideError();
      });
    },
    closeEditorAndShowParent: function () {
      setActiveEntityUrlHash();
      store.commit('setTreeSelection', {
        countryId: null,
        regionId: null,
        riverId: null,
        spotId: null
      });
      store.commit("setCountryEditorEditMode", false);
      store.dispatch('reloadCountries');
      store.commit('showEmptyPage');
    },
    showError: function (errMsg) {
      store.commit("setErrMsg", errMsg);
    },
    hideError: function () {
      store.commit("setErrMsg", null);
    },
  },
}

</script>
