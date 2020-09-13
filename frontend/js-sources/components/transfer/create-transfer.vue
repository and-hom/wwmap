<template>
    <div id="add-transfer" class="modal fade" tabindex="-1" role="dialog">
        <div class="modal-dialog modal-dialog-centered" role="document">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Добавить трансфер</h5>
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                </div>
                <div class="modal-body">
                  <transfer-form v-model="transfer"/>
                  <label for="rivers">Реки</label>
                  <div id="rivers">
                    <div v-if="refreshRiverHack">
                      <div v-for="river in transfer.rivers" style="min-height: 32px;">{{ river.title }}
                        <button style="float: right;" v-on:click.stop="delRiver(river.id)">[X]</button>
                      </div>
                    </div>
                    <river-select
                        v-model="river"
                        v-on:input="addRiver($event)"></river-select>
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
    import {backendApiBase} from '../../config'
    import {doPostJson} from '../../api'

    module.exports = {
        props: {
            transfer: {
                type: Object,
                required: true,
            },
            stations: {
                type: Array,
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
                    return this.transfer.rivers ? this.transfer.rivers : [];
                },
                set: function (val) {
                    this.transfer.rivers = val;
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
                doPostJson(backendApiBase + "/transfer", this.transfer, true).then(_ => {
                    this.okFn();
                });
            },
            cancel: function () {
                if (this.cancelFn) {
                    this.cancelFn();
                }
            },
            addRiver: function (river) {
                if (river == null) {
                    return;
                }
                this.refreshRiverHack = false;
                if (this.transfer.rivers==null) {
                    this.transfer.rivers = [];
                }

                for (let i = 0; i < this.transfer.rivers.length; i++) {
                    if (this.transfer.rivers[i].id == river.id) {
                        return
                    }
                }
                this.transfer.rivers.push(river);
                this.transfer.rivers.sort((a, b) => a.id - b.id);
                this.refreshRiverHack = true;
            },
            delRiver: function (id) {
                this.refreshRiverHack = false;
                this.transfer.rivers = this.transfer.rivers.filter(x => x.id!=id)
                this.refreshRiverHack = true;
            }
        }
    }
</script>

<style>
</style>