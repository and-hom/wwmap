<template>
    <div>
        <window-title v-bind:text="spot.title + ' / ' + spot.title"></window-title>
        <div style="display:none;" :id="initialSpot.id"></div>
        <ask id="del-spot" title="Точно?" msg="Совсем удалить?" :ok-fn="function() { remove(); }"></ask>
        <ask id="mid-point-about" title="Опорные точки" :no-btn="false" ok-btn-title="Понятно"
             msg='Иногда река в месте порога изгибается. Чтобы контур вокруг порога повторял этот изгиб, нужно добавить несколько опорных точек в середине. Они не несут смысловой нагрузки и нужны только для отрисовки'
             :ok-fn="function() {}"></ask>

        <div v-if="canEdit" class="btn-toolbar justify-content-between">
            <div class="btn-group mr-2" role="group" aria-label="First group">
                <button type="button" class="btn btn-info" v-if="!editMode" v-on:click="editMode=true; hideError();">
                    Редактирование
                </button>
                <button type="button" class="btn btn-success" v-if="editMode" v-on:click="$refs.editor.save()">
                    Сохранить
                </button>
                <button type="button" class="btn btn-secondary" v-if="editMode" v-on:click="cancelEditing()">Отменить
                </button>
                <button type="button" class="btn btn-danger" v-if="spot.id>0" data-toggle="modal"
                        data-target="#del-spot">Удалить
                </button>
            </div>
            <div class="btn-group mr-2">
                <log-dropdown object-type="SPOT" :object-id="spot.id"/>
            </div>
        </div>

        <spot-viewer v-if="!editMode"
                     :spot="spot"
                     :country="country"
                     :region="region"
                     :images="images"
                     :schemas="schemas"
                     :videos="videos"/>

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
                     v-on:videos="videos = $event"/>
    </div>
</template>

<script>
    import {store} from '../../main';
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth';
    import {getImages, getSpot, nvlReturningId, removeSpot, setActiveEntityUrlHash} from '../../editor';

    module.exports = {
        props: ['initialSpot', 'country', 'region'],
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
                    _ => this.closeEditorAndShowSpot(),
                    err => this.showError("не могу удалить: " + err))
            },
            cancelEditing: function () {
                this.editMode = false;
                if (this.spot && this.spot.id > 0) {
                    this.reload();
                } else {
                    this.closeEditorAndShowSpot();
                }
            },
            reload: function () {
                getSpot(this.spot.id).then(spot => {
                    this.spot = spot;
                    this.hideError();
                });
            },
            closeEditorAndShowSpot: function () {
                setActiveEntityUrlHash(this.country.id, nvlReturningId(this.region));
                store.commit('setActiveEntityState', {
                    countryId: this.country.id,
                    regionId: nvlReturningId(this.region),
                    spotId: null,
                    spotId: null
                });
                store.commit("setSpotEditorVisible", false);
                if (this.spot.region.fake || this.spot.region.id == 0) {
                    store.commit('showCountrySubentities', this.country.id);
                    store.commit('selectCountry', {country: this.country});
                } else {
                    store.commit('showRegionSubentities', {
                        countryId: this.country.id,
                        regionId: nvlReturningId(this.spot.region)
                    });
                    store.commit('selectRegion', {
                        country: this.country,
                        countryId: nvlReturningId(this.spot.region)
                    });
                }
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
        }
    }
</script>

<style>
</style>