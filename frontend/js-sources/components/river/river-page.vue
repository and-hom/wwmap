<template>
    <div>
        <window-title :text="river.title"></window-title>
        <ask id="del-river" title="Точно?"
             msg="Совсем удалить? Все пороги будут также удалены! Да, совсем! Восстановить будет никак нельзя!"
             :ok-fn="function() { remove(); }"></ask>

        <river-viewer v-if="pageMode=='view'"
                      :river="river"
                      :reports="reports"
                      :transfers="transfers"
                      :country="country"
                      :region="region">
            <button type="button" class="btn btn-primary" v-on:click="add_spot()">Добавить препятствие</button>
            <button type="button" class="btn btn-info"
                    v-on:click="pageMode='edit'; hideError();">Редактирование</button>
            <button type="button" class="btn btn-success"
                    v-on:click="pageMode='batch-edit'; hideError();">Пакетное редактирование и загрузка GPX</button>
            <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-river" v-if="river.id>0">Удалить
            </button>
        </river-viewer>

        <river-editor v-if="canEdit && pageMode=='edit'"
                      ref="editor"
                      :river="river"
                      :reports="reports"
                      :transfers="transfers"
                      :country="country"
                      :region="region"
                      v:sensors="sensors">
            <button type="button" class="btn btn-secondary" v-on:click="cancelEditing()">Отменить</button>
            <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-river" v-if="river.id>0">Удалить
            </button>
        </river-editor>

        <river-batch-editor v-if="canEdit && pageMode=='batch-edit'"
                      ref="batchEditor"
                      :river="river"
                      :reports="reports"
                      :country="country"
                      :region="region"
                      v:sensors="sensors">
            <button type="button" class="btn btn-secondary" v-on:click="cancelEditing()">Отменить</button>
            <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-river" v-if="river.id>0">Удалить
            </button>
        </river-batch-editor>
    </div>
</template>

<script>
    import {store} from '../../app-state';
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth';
    import {getRiver, nvlReturningId, removeRiver, setActiveEntityUrlHash} from '../../editor';
    import {createMapParamsStorage} from 'wwmap-js-commons/map-settings'

    module.exports = {
        props: ['initialRiver', 'reports', 'transfers', 'country', 'region'],
        computed: {
            pageMode: {
                get: function () {
                    return store.state.rivereditorstate.pageMode
                },

                set: function (newVal) {
                    store.commit("setRiverEditorPageMode", newVal);
                }
            },
        },
        created: function () {
            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
            this.resetToInitialIfRequired();
        },
        updated: function () {
            this.resetToInitialIfRequired();
        },
        data: function () {
            return {
                river: null,
                previousRiverId: this.initialRiver.id,
                canEdit: false,
                mapParamsStorage: createMapParamsStorage(),
            }
        },
        methods: {
            resetToInitialIfRequired: function () {
                if (this.shouldReInit()) {
                    this.previousRiverId = this.initialRiver.id;
                    this.river = this.initialRiver;
                }
            },
            shouldReInit: function () {
                return this.river == null ||
                    this.previousRiverId !== this.initialRiver.id && this.initialRiver.id > 0
            },
            add_spot: function () {
                store.commit("setRiverEditorVisible", false);
                store.commit("setSpotEditorVisible", false);

                let lastPositionZoomType = this.mapParamsStorage.getLastPositionZoomType();
                store.commit("setSpotEditorState", {
                    visible: true,
                    editMode: true,
                    spot: {
                        id: 0,
                        river: this.river,
                        order_index: "0",
                        automatic_ordering: true,
                        point: lastPositionZoomType.position,
                        aliases: [],
                        props: {},
                    },
                    country: this.country,
                    region: this.region,
                    river: this.river,
                    zoom: lastPositionZoomType.zoom,
                });
            },
            remove: function () {
                this.hideError();
                removeRiver(this.river.id).then(
                    _ => this.closeEditorAndShowParent(),
                    err => this.showError("не могу удалить: " + err))
            },
            cancelEditing: function () {
                this.pageMode = 'view';
                if (this.river && this.river.id > 0) {
                    this.reload();
                } else {
                    this.closeEditorAndShowParent();
                }
            },
            reload: function () {
                getRiver(this.river.id).then(river => {
                    this.river = river;
                    this.hideError();
                });
            },
            closeEditorAndShowParent: function () {
                setActiveEntityUrlHash(this.country.id, nvlReturningId(this.region));
                store.commit('setTreeSelection', {
                    countryId: this.country.id,
                    regionId: nvlReturningId(this.region),
                    riverId: null,
                    spotId: null
                });
                store.commit("setRiverEditorVisible", false);
                if (this.river.region.fake || this.river.region.id == 0) {
                    store.dispatch('reloadCountrySubentities', this.country.id);
                    store.commit('showCountryPage', {country: this.country});
                } else {
                    store.dispatch('reloadRegionSubentities', {
                        countryId: this.country.id,
                        regionId: nvlReturningId(this.river.region)
                    });
                    store.commit('showRegionPage', {
                        country: this.country,
                        regionId: nvlReturningId(this.river.region)
                    });
                }
            },
            showError: function (errMsg) {
                store.commit("setErrMsg", errMsg);
            },
            hideError: function () {
                store.commit("setErrMsg", null);
            },
        }
    }
</script>

<style>
</style>