<template>
    <div>
        <window-title v-bind:text="river.title + ' / ' + spot.title"></window-title>
        <div style="display:none;" :id="initialSpot.id"></div>
        <ask id="del-spot" title="Точно?" msg="Совсем удалить?" :ok-fn="function() { remove(); }"></ask>
        <ask id="mid-point-about" title="Опорные точки" :no-btn="false" ok-btn-title="Понятно"
             msg='Иногда река в месте порога изгибается. Чтобы контур вокруг порога повторял этот изгиб, нужно добавить несколько опорных точек в середине. Они не несут смысловой нагрузки и нужны только для отрисовки'
             :ok-fn="function() {}"></ask>

        <spot-viewer v-if="!editMode"
                     :spot="spot"
                     :country="country"
                     :region="region"
                     :images="images"
                     :schemas="schemas"
                     :videos="videos"

                     v-on:spotClick="navigateToSpot($event, false)">
            <button type="button" class="btn btn-info" v-if="!editMode" v-on:click="editMode=true; hideError();">
                Редактирование
            </button>
            <button type="button" class="btn btn-danger" v-if="spot.id>0" data-toggle="modal"
                    data-target="#del-spot">Удалить
            </button>
        </spot-viewer>

        <spot-editor v-if="editMode"
                     ref="editor"
                     :spot="spot"
                     :country="country"
                     :region="region"
                     :images="images"
                     :schemas="schemas"
                     :videos="videos"

                     v-on:images="images = $event"
                     v-on:schemas="schemas = $event"
                     v-on:videos="videos = $event"
                     v-on:spot="spot = $event"

                     v-on:spotClick="navigateToSpot($event, true)">
            <button type="button" class="btn btn-secondary" v-if="editMode" v-on:click="cancelEditing()">Отменить
            </button>
            <button type="button" class="btn btn-danger" v-if="spot.id>0" data-toggle="modal"
                    data-target="#del-spot">Удалить
            </button>
        </spot-editor>
    </div>
</template>

<script>
    import {store, navigateToSpot} from '../../app-state';
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth';
    import {getImages, getSpot, nvlReturningId, removeSpot, setActiveEntityUrlHash} from '../../editor';

    module.exports = {
        props: ['initialSpot', 'country', 'region', 'river'],
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
        created: function () {
            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
            this.resetToInitialIfRequired();
        },
        updated: function () {
            this.resetToInitialIfRequired();
        },
        data: function () {
            return {
                spot: null,
                previousSpotId: this.initialSpot.id,
                canEdit: false,
                center: [41, 41],
                bounds: [[-1, -1], [1, 1]],

                images: [],
                schemas: [],
                videos: [],
            }
        },
        methods: {
            resetToInitialIfRequired: function () {
                if (this.shouldReInit()) {
                    this.previousSpotId = this.initialSpot.id;
                    this.spot = this.initialSpot;
                    this.reloadImgs();
                }
            },
            shouldReInit: function () {
                return this.spot == null ||
                    this.previousSpotId !== this.initialSpot.id && this.initialSpot.id > 0
            },
            remove: function () {
                this.hideError();
                removeSpot(this.spot.id).then(
                    _ => this.closeEditorAndShowParent(),
                    err => this.showError("не могу удалить: " + err))
            },
            cancelEditing: function () {
                this.editMode = false;
                if (this.spot && this.spot.id > 0) {
                    this.reload();
                } else {
                    this.closeEditorAndShowParent();
                }
            },
            reload: function () {
                getSpot(this.spot.id).then(spot => {
                    this.spot = spot;
                    this.hideError();
                });
            },
            closeEditorAndShowParent: function () {
                setActiveEntityUrlHash(this.country.id, nvlReturningId(this.region), this.river.id);
                store.commit('setTreeSelection', {
                    countryId: this.country.id,
                    regionId: nvlReturningId(this.region),
                    riverId: this.river.id,
                    spotId: null
                });
                store.commit("setSpotEditorVisible", false);
                store.dispatch('reloadRegionSubentities', {
                    countryId: this.country.id,
                    regionId: nvlReturningId(this.spot.region)
                });
                store.dispatch('reloadRiverSubentities', {
                    countryId: this.country.id,
                    regionId: nvlReturningId(this.spot.region),
                    riverId: this.river.id,
                });
                store.commit('showRiverPage', {
                    country: this.country,
                    regionId: this.spot.region,
                    riverId: this.river.id,
                });
            },
            reloadImgs: function () {
                this.images = [];
                this.schemas = [];
                this.videos = [];

                getImages(this.spot.id, "image").then(images => this.images = images);
                getImages(this.spot.id, "schema").then(schemas => this.schemas = schemas);
                getImages(this.spot.id, "video").then(videos => this.videos = videos);
            },
            showError: function (errMsg) {
                store.commit("setErrMsg", errMsg);
            },
            hideError: function () {
                store.commit("setErrMsg", null);
            },
            navigateToSpot: function(id, edit) {
                navigateToSpot(id, edit);
            },
        }
    }
</script>

<style>
</style>