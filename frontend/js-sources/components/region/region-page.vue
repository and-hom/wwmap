<template>
  <div>
    <window-title v-bind:text="region.title"></window-title>
    <ask id="del-region" title="Точно?"
         msg="Совсем удалить? Восстановить будет никак нельзя!"
         :ok-fn="function() { remove(); }"></ask>
    <ask id="del-region-forbidden" title="Нельзя удалить!"
         msg="Нельзя удалить регион, в котором есть реки. Сначала удалите все реки этого региона вручную."
         :okBtn="false" :noBtn="false" :cancelBtn="true"></ask>

    <region-viewer v-if="!editMode"
                   :country="country"
                   :region="region">
      <button type="button" class="btn btn-primary" v-on:click="add_river()">Добавить реку</button>
      <button type="button" class="btn btn-info" v-if="canEditRegion"
              v-on:click="editMode=true; hideError();">Редактирование
      </button>
      <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-region"
              v-if="region.id>0 && !region.has_rivers">
        Удалить
      </button>
      <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-region-forbidden"
              v-if="region.id>0 && region.has_rivers">
        Удалить
      </button>
    </region-viewer>

    <region-editor v-if="canEditRegion && editMode"
                   ref="editor"
                   :country="country"
                   :region="region">
      <button type="button" class="btn btn-secondary" v-on:click="cancelEditing()">Отменить</button>
      <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-region"
              v-if="region.id>0 && !region.has_rivers">
        Удалить
      </button>
      <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-region-forbidden"
              v-if="region.id>0 && region.has_rivers">
        Удалить
      </button>
    </region-editor>
  </div>
</template>

<script>
import {store} from '../../app-state';
import {hasRole, ROLE_ADMIN} from '../../auth';
import {getRegion, removeRegion, setActiveEntityUrlHash} from "../../editor";

module.exports = {
  props: ['initialRegion', 'country'],
  created: function () {
    hasRole(ROLE_ADMIN).then(canEdit => this.canEdit = canEdit);
    hasRole(ROLE_ADMIN).then(canEditRegion => this.canEditRegion = canEditRegion);
    this.resetToInitialIfRequired();
  },
  updated: function () {
    this.resetToInitialIfRequired();
  },
  computed: {
    editMode: {
      get: function () {
        return store.state.regioneditorstate.editMode
      },

      set: function (newVal) {
        store.commit("setRegionEditorEditMode", newVal);
      }
    },
  },
  data: function () {
    return {
      canEdit: false,
      canEditRegion: false,
      region: null,
      previousRegionId: this.initialRegion.id,

      getEditModeButtonTitle: function () {
        return this.editMode ? 'Просмотр' : 'Редактирование';
      },
      // end of editor

      add_river: function () {
        store.commit('newRiver', {
          country: this.country,
          region: this.region
        })
      },
    }
  },
  methods: {
    resetToInitialIfRequired: function () {
      if (this.shouldReInit()) {
        this.previousRegionId = this.initialRegion.id;
        this.region = this.initialRegion;
      }
    },
    shouldReInit: function () {
      return this.region == null ||
          this.previousRegionId !== this.initialRegion.id && this.initialRegion.id > 0
    },
    remove: function () {
      this.hideError();
      removeRegion(this.region.id).then(
          _ => this.closeEditorAndShowParent(),
          err => this.showError("не могу удалить: " + err))
    },
    cancelEditing: function () {
      this.editMode = false;
      if (this.region && this.region.id > 0) {
        this.reload();
      } else {
        this.closeEditorAndShowParent();
      }
    },
    reload: function () {
      getRegion(this.region.id).then(region => {
        this.region = region;
        this.hideError();
      });
    },
    closeEditorAndShowParent: function () {
      setActiveEntityUrlHash(this.country.id);
      store.commit('setTreeSelection', {
        countryId: this.country.id,
        regionId: null,
        riverId: null,
        spotId: null
      });
      store.commit("setRegionEditorEditMode", false);
      store.dispatch('reloadCountrySubentities', this.country.id);
      store.commit('showCountryPage', {country: this.country});
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