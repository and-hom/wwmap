<template>
  <div id="add" class="modal fade" tabindex="-1" role="dialog">
    <div :class="modalDialogClass" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Добавить</h5>
          <button type="button" class="close" data-dismiss="modal" aria-label="Close">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <div class="settings-container">
            <slot name="form" v-bind:entity="entity">
            </slot>
            <label for="rivers">Реки</label>
            <div id="rivers">
              <river-select
                  v-model="entity.rivers"
                  :bind-id="true"
                  :multiselect="true"/>
            </div>
          </div>
          <div v-if="hasMap" class="map-container">
            <ya-map-location :spot="entity"
                             v-bind:refresh-on-change="entity.point"
                             width="100%"
                             height="600px"
                             :editable="true"
                             :ya-search="true"/>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-primary" data-dismiss="modal" v-on:click="save()">
            Сохранить
          </button>
          <button type="button" class="btn btn-secondary" data-dismiss="modal" v-on:click="cancel()">
            Отмена
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style type="text/css">
.entity-editor-form {
  width: 80%;
  max-width: 80%;
}

div.settings {
  display: grid;
  grid-template-columns: max-content max-content;
  grid-gap: 5px;
}

div.settings label {
  text-align: right;
  margin-right: 15px;
}

div.settings label:after {
  content: ":";
}

.modal-body {
  display: flex;
}

.settings-container {
  width: 550px;
  margin: 25px;
}

.map-container {
  width: 100%;
}
</style>

<script>
import {doPostJson} from '../../../api'

module.exports = {
  props: {
    entity: {
      type: Object,
      required: true,
    },
    hasMap: {
      type: Boolean,
      default: false,
    },
    urlBase: {
      type: String,
      required: true,
    },

    okFn: {
      type: Function,
      required: true,
    },
    cancelFn: {
      type: Function,
      required: false,
    },
    failFn: {
      type: Function,
      required: false,
    },
  },
  computed: {
    modalDialogClass() {
      return this.hasMap
          ? 'modal-dialog modal-dialog-centered entity-editor-form'
          : 'modal-dialog modal-dialog-centered';
    }
  },
  data: function () {
    return {
      refreshRiverHack: true,
    }
  },
  methods: {
    save: function () {
      doPostJson(this.urlBase, this.entity, true).then(_ => {
        this.okFn();
      }, err => {
        if (this.failFn) {
          this.failFn(err)
        }
      });
    },
    cancel: function () {
      if (this.cancelFn) {
        this.cancelFn();
      }
    },
  }
}
</script>

<style>
</style>