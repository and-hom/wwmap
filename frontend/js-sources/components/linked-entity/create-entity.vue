<template>
    <div id="add" class="modal fade" tabindex="-1" role="dialog">
        <div class="modal-dialog modal-dialog-centered" role="document">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Добавить</h5>
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                </div>
                <div class="modal-body">
                  <slot name="form" v-bind:entity="entity">
                  </slot>
                  <label for="rivers">Реки</label>
                  <div id="rivers">
                    <river-select
                        v-model="entity.rivers"
                        :bind-id="true"
                        :multiselect="true"></river-select>
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
</style>

<script>
import {doPostJson} from '../../api'

module.exports = {
        props: {
            entity: {
                type: Object,
                required: true,
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
        },
        computed: {
            rivers: {
                get: function () {
                    return this.entity.rivers ? this.entity.rivers : [];
                },
                set: function (val) {
                    this.entity.rivers = val;
                },
            }
        },
        data: function () {
            return {
                river: null,
                refreshRiverHack:true,
            }
        },
        methods: {
            save: function () {
                doPostJson(this.urlBase, this.entity, true).then(_ => {
                    this.okFn();
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