import Vue from 'vue'
import Vuex from 'vuex'
import EditorPage from './editor-page.vue'

import VueSelect from 'vue-select'
import VueGallery from 'vue-gallery';
import FileUpload from 'vue-upload-component';

import upperFirst from 'lodash/upperFirst'
import camelCase from 'lodash/camelCase'

import {
    COUNTRY_ACTIVE_ENTITY_LEVEL,
    getActiveId,
    getRegions,
    getRiversByCountry,
    getSpots,
    REGION_ACTIVE_ENTITY_LEVEL,
    RIVER_ACTIVE_ENTITY_LEVEL,
    SPOT_ACTIVE_ENTITY_LEVEL
} from './editor'
import {getAuthorizedUserInfoOrNull} from './auth'
import {sensors} from './sensors'


import 'bootstrap/dist/css/bootstrap.min.css';
import '../css/editor.css'

export var app;


function getById(arr, id) {
    var filtered = arr.filter(function (x) {
        return x.id === id
    });
    if (filtered.length > 0) {
        return filtered[0]
    }
    return null
}

export function getRiverFromTree(countryId, regionId, id) {
    let river;
    let country = store.state.treePath[countryId];
    if (!country) {
        country = showCountrySubentities(countryId)
    }

    if (regionId && regionId > 0) {
        var region = getById(country.regions, regionId);
        let rivers = region.rivers;
        if (!rivers) {
            rivers = getRiversByRegion(countryId, region.id);
            Vue.set(region, "rivers", rivers);
        }
        river = getById(rivers, id)
    } else {
        river = getById(country.rivers, id)
    }
    return river;
}


export function getSpotsFromTree(countryId, regionId, riverId) {
    var river;
    if (regionId && regionId > 0) {
        var region = getById(store.state.treePath[countryId].regions, regionId);
        river = getById(region.rivers, riverId)
    } else {
        river = getById(store.state.treePath[countryId].rivers, riverId)
    }
    return river.spots
}


Vue.use(Vuex);
export const store = new Vuex.Store({
    state: {
        spoteditorstate: {
            visible: false,
            editMode: false,
            images: [],
            schemas: []
        },
        rivereditorstate: {
            visible: false,
            pageMode: 'view'
        },
        regioneditorstate: {
            visible: false,
            editMode: false
        },
        countryeditorstate: {
            visible: false,
            editMode: false
        },

        errMsg: "",
        closeCallback: function () {
            // do nothing
        },


        userInfo: getAuthorizedUserInfoOrNull(),
        treePath: {},
        selectedCountry: getActiveId(COUNTRY_ACTIVE_ENTITY_LEVEL),
        selectedRegion: getActiveId(REGION_ACTIVE_ENTITY_LEVEL),
        selectedRiver: getActiveId(RIVER_ACTIVE_ENTITY_LEVEL),
        selectedSpot: getActiveId(SPOT_ACTIVE_ENTITY_LEVEL),
        sensors: sensors,
    },
    getters: {},
    mutations: {
        hideAll(state) {
            state.spoteditorstate.visible = false;
            state.rivereditorstate.visible = false;
            state.regioneditorstate.visible = false;
            state.countryeditorstate.visible = false;
        },

        selectCountry(state, country) {
            state.countryeditorstate.country = country;
            state.countryeditorstate.editMode = false;
            state.countryeditorstate.visible = true
        },
        selectRegion(state, country, id) {
            state.regioneditorstate.region = getRegion(id);
            state.regioneditorstate.country = country;
            state.regioneditorstate.editMode = false;
            state.regioneditorstate.visible = true
        },
        selectRiver(state, country, region, id) {
            state.rivereditorstate.river = getRiver(id);
            state.rivereditorstate.pageMode = 'view';
            state.rivereditorstate.reports = getReports(id);
            state.rivereditorstate.country = country;
            state.rivereditorstate.region = region;
            state.rivereditorstate.visible = true;
        },

        setActiveEntityState(state, countryId, regionId, riverId, spotId) {
            state.selectedSpot = spotId;
            state.selectedRiver = riverId;
            state.selectedRegion = regionId;
            state.selectedCountry = countryId;
        },

        newRiver(state, country, region) {
            state.spoteditorstate.visible = false;
            state.rivereditorstate.visible = false;
            state.regioneditorstate.visible = false;
            state.countryeditorstate.visible = false;

            state.rivereditorstate.visible = true;
            state.rivereditorstate.pageMode = 'edit';
            state.rivereditorstate.river = {
                id: 0,
                region: region,
                aliases: [],
                props: {}
            };
            state.rivereditorstate.country = country;
            state.rivereditorstate.region = region;
        },

        showCountrySubentities(state, id) {
            Promise.all([getRiversByCountry(id), getRegions(id)]).then(result => {
                let country = {
                    rivers: result[0],
                    regions: result[1],
                };
                Vue.set(store.state.treePath, id, country);
                return country;
            })
        },
        hideCountrySubentities(state, id) {
            Vue.delete(store.state.treePath, id)
        },

        showRiverSubentities(state, id) {
            if (store.state.treePath[t.country.id]) {
                getSpots(id).then(spots => Vue.set(t.river, "spots", spots));
            }
        },
        hideRiverSubentities(state, id) {
            Vue.delete(store.state.treePath, id)
        },

        showRegionTree(state, countryId, id) {
            var region = getById(state.treePath[countryId].regions, id);
            Vue.set(region, "rivers", getRiversByRegion(countryId, id));
        },

        showRiverTree(state, countryId, regionId, id) {
            var river = getRiverFromTree(countryId, regionId, id);
            getSpots(id).then(spots => Vue.set(river, "spots", spots))
        },
        hideRiverTree(state, countryId, regionId, id) {
            var river = getRiverFromTree(countryId, regionId, id);
            Vue.delete(river, "spots")
        },

        onTreeSwitch(state, callback) {
            state.errMsg = null;
            state.closeCallback = callback;

            if (!state.spoteditorstate.editMode && state.rivereditorstate.pageMode != 'edit' && state.rivereditorstate.pageMode != 'batch-edit') {
                callback();
                return;
            }

            if (state.spoteditorstate.editMode) {
                let spotEditorCloseDialog = $('#close-spot-editor');
                spotEditorCloseDialog.on('hidden.bs.modal', e => {
                    state.closeCallback = function () {

                    };
                });
                spotEditorCloseDialog.modal();
            }
            if (state.rivereditorstate.pageMode == 'edit' || state.rivereditorstate.pageMode == 'batch-edit') {
                let riverEditorCloseDialog = $('#close-river-editor');
                riverEditorCloseDialog.on('hidden.bs.modal', e => {
                    // state.closeCallback = function () {
                    //
                    // };
                });
                riverEditorCloseDialog.modal();
            }
        },
    }
});

export function init() {
    Vue.component('v-select', VueSelect.VueSelect);
    Vue.component('gallery', VueGallery);
    Vue.component('file-upload', FileUpload);

    const requireComponent = require.context('./components', true, /.+?\.(vue|js)$/);

    requireComponent.keys().forEach(fileName => {
        const componentConfig = requireComponent(fileName);
        const componentName = upperFirst(
            camelCase(fileName
                .split('/')
                .pop()
                .replace(/\.\w+$/, '')
            )
        );

        Vue.component(componentName, componentConfig.default || componentConfig)
    });

    app = new Vue({
        el: '#vue-app',
        render: h => h(EditorPage),
    });
}

export function getApp() {
    return app
}
