<template>
    <div>
        <window-title :text="river.title"></window-title>
        <ask id="del-river" title="Точно?"
             msg="Совсем удалить? Все пороги будут также удалены! Да, совсем! Восстановить будет никак нельзя!"
             :ok-fn="function() { remove(); }"></ask>

        <div v-if="canEdit" class="btn-toolbar justify-content-between">
            <div class="btn-group mr-2" role="group">
                <button v-if="river.id && pageMode == 'view'" type="button" class="btn btn-primary"
                        v-on:click="add_spot()">Добавить препятствие
                </button>
                <button type="button" class="btn btn-info" v-if="pageMode == 'view'"
                        v-on:click="pageMode='edit'; hideError();">
                    Редактирование
                </button>
                <button type="button" class="btn btn-success" v-if="river.id && pageMode == 'view'"
                        v-on:click="pageMode='batch-edit'; hideError();">
                    Пакетное редактирование и загрузка GPX
                </button>
                <button type="button" class="btn btn-success" v-if="pageMode == 'edit'"
                        v-on:click="$refs.editor.save()">Сохранить
                </button>
                <button type="button" class="btn btn-success" v-if="pageMode == 'batch-edit'"
                        v-on:click="$refs.batchEditor.saveSpotsBatch()">Сохранить
                </button>
                <button type="button" class="btn btn-secondary" v-if="pageMode != 'view'"
                        v-on:click="cancelEditing()">Отменить
                </button>
                <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-river" v-if="river.id>0">Удалить
                </button>
            </div>
            <div class="btn-group mr-2">
                <log-dropdown object-type="RIVER" :object-id="river.id"/>
            </div>
        </div>

        <river-viewer v-if="pageMode=='view'"
                      :river="river"
                      :reports="reports"
                      :country="country"
                      :region="region"/>

        <river-editor v-if="pageMode=='edit'"
                      ref="editor"
                      :river="river"
                      :reports="reports"
                      :country="country"
                      :region="region"
                      v:sensors="sensors"/>

        <river-batch-editor v-if="pageMode=='batch-edit'"
                      ref="batchEditor"
                      :river="river"
                      :reports="reports"
                      :country="country"
                      :region="region"
                      v:sensors="sensors"/>
    </div>
</template>

<script>
    import {store} from '../../main';
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth';
    import {getRiver, getRiverBounds, nvlReturningId, removeRiver, setActiveEntityUrlHash} from '../../editor';

    module.exports = {
        props: ['initialRiver', 'reports', 'country', 'region'],
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
                center: [41, 41],
                bounds: [[-1, -1], [1, 1]]
            }
        },
        methods: {
            resetToInitialIfRequired: function () {
                if (this.shouldReInit()) {
                    this.previousRiverId = this.initialRiver.id;
                    this.river = this.initialRiver;
                    getRiverBounds(this.river.id).then(bounds => {
                        this.bounds = bounds;
                        this.center = [(bounds[0][0] + bounds[1][0]) / 2, (bounds[0][1] + bounds[1][1]) / 2];
                    })
                }
            },
            shouldReInit: function () {
                return this.river == null ||
                    this.previousRiverId !== this.initialRiver.id && this.initialRiver.id > 0
            },
            add_spot: function () {
                store.commit("setRiverEditorVisible", false);
                store.commit("setSpotEditorVisible", false);

                store.commit("setSpotEditorState", {
                    visible: true,
                    editMode: true,
                    spot: {
                        id: 0,
                        river: this.river,
                        order_index: "0",
                        automatic_ordering: true,
                        point: this.center,
                        aliases: [],
                        props: {},
                    },
                    country: this.country,
                    region: this.region,
                    river: this.river,
                });
            },
            remove: function () {
                this.hideError();
                removeRiver(this.river.id).then(
                    _ => this.closeEditorAndShowParent(),
                    err => this.showError("не могу удалить: " + err))
            },
            cancelEditing: function () {
                this.pageMode='view';
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
                store.commit('setActiveEntityState', {
                    countryId: this.country.id,
                    regionId: nvlReturningId(this.region),
                    riverId: null,
                    spotId: null
                });
                store.commit("setRiverEditorVisible", false);
                if (this.river.region.fake || this.river.region.id == 0) {
                    store.commit('showCountrySubentities', this.country.id);
                    store.commit('selectCountry', {country: this.country});
                } else {
                    store.commit('showRegionSubentities', {
                        countryId: this.country.id,
                        regionId: nvlReturningId(this.river.region)
                    });
                    store.commit('selectRegion', {
                        country: this.country,
                        countryId: nvlReturningId(this.river.region)
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