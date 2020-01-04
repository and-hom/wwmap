import Vue from 'vue'
import Vuex from 'vuex'
import EditorPage from './editor-page.vue'
import DocsIntegrationPage from './docs-integration-page.vue'
import MapPage from './map-page.vue'
import LogPage from './log-page.vue'
import UsersPage from './users-page.vue'
import TechPage from './tech-page.vue'
import AboutPage from './about-page.vue'
import DocsPage from './docs-page.vue'
import DashboardPage from './dashboard-page.vue'

import vSelect from 'vue-select'
import VueGallery from 'vue-gallery';
import FileUpload from 'vue-upload-component';
import Datepicker from 'vuejs-datepicker';
import {ImageRating} from 'vue-rate-it';

import upperFirst from 'lodash/upperFirst'
import camelCase from 'lodash/camelCase'

import {TabsPlugin} from 'bootstrap-vue'

const moment = require('moment');

import {
    COUNTRY_ACTIVE_ENTITY_LEVEL,
    getActiveId,
    getRegion,
    getRegions,
    getReports,
    getRiver,
    getRiversByCountry,
    getRiversByRegion,
    getSpot,
    getSpots,
    REGION_ACTIVE_ENTITY_LEVEL,
    RIVER_ACTIVE_ENTITY_LEVEL,
    SPOT_ACTIVE_ENTITY_LEVEL
} from './editor'
import {getAuthorizedUserInfoOrNull} from './auth'
import {sensors} from './sensors'
import 'bootstrap/dist/css/bootstrap.min.css';
import 'vue-select/dist/vue-select.css';

import './style/main.css'
import './style/editor.css'

require("bootstrap");


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

export function getRiverFromTree(countryId, regionId, riverId) {
    let river;
    let country = store.state.treePath[countryId];
    if (!country) {
        country = showCountrySubentities(countryId)
    }

    if (regionId && regionId > 0) {
        var region = getById(country.regions, regionId);
        let rivers = region.rivers;
        if (!rivers) {
            rivers = [];//getRiversByRegion(countryId, region.id);
            Vue.set(region, "rivers", rivers);
        }
        river = getById(rivers, riverId)
    } else {
        river = getById(country.rivers, riverId)
    }
    return river;
}

export function getRegionFromTree(countryId, regionId) {
    let country = store.state.treePath[countryId];
    if (!country) {
        //country = showCountrySubentities(countryId)
        return null;
    }

    return regionId && regionId > 0 ? getById(country.regions, regionId) : null;
}


export function getSpotsFromTree(countryId, regionId, riverId) {
    var river;
    if (regionId && regionId > 0) {
        var region = getById(store.state.treePath[countryId].regions, regionId);
        river = getById(region.rivers, riverId)
    } else {
        river = getById(store.state.treePath[countryId].rivers, riverId)
    }
    return river.spots ? river.spots : []
}


Vue.use(Vuex);
export const store = new Vuex.Store({
    state: {
        spoteditorstate: {
            visible: false,
            editMode: false,
            images: [],
            schemas: [],
            videos: [],
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

        selectCountry(state, payload) {
            state.countryeditorstate.country = payload.country;
            state.countryeditorstate.editMode = false;
            state.countryeditorstate.visible = true
        },
        selectRegion(state, payload) {
            state.regioneditorstate.region = getRegion(payload.regionId);
            state.regioneditorstate.country = payload.country;
            state.regioneditorstate.editMode = false;
            state.regioneditorstate.visible = true
        },
        selectRiver(state, payload) {
            getRiver(payload.riverId).then(river => {
                state.rivereditorstate.river = river;
                state.rivereditorstate.pageMode = 'view';
                state.rivereditorstate.country = payload.country;
                state.rivereditorstate.region = payload.region;
                state.rivereditorstate.visible = true;
                getReports(payload.riverId).then(reports => {
                    state.rivereditorstate.reports = reports;
                });
            });
        },
        selectSpot(state, payload) {
            getSpot(payload.spotId).then(spot => {
                state.spoteditorstate.country = payload.country;
                state.spoteditorstate.region = payload.region;
                state.spoteditorstate.river = payload.river;
                state.spoteditorstate.spot = spot;
                state.spoteditorstate.editMode = false;
                state.spoteditorstate.visible = true;
            });
        },

        setActiveEntityState(state, payload) {
            state.selectedSpot = payload.spotId;
            state.selectedRiver = payload.riverId;
            state.selectedRegion = payload.regionId;
            state.selectedCountry = payload.countryId;
        },

        newRiver(state, payload) {
            state.spoteditorstate.visible = false;
            state.rivereditorstate.visible = false;
            state.regioneditorstate.visible = false;
            state.countryeditorstate.visible = false;

            state.rivereditorstate.visible = true;
            state.rivereditorstate.pageMode = 'edit';
            state.rivereditorstate.river = {
                id: 0,
                region: payload.region,
                aliases: [],
                props: {}
            };
            state.rivereditorstate.country = payload.country;
            state.rivereditorstate.region = payload.region;
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

        showRegionSubentities(state, payload) {
            getRiversByRegion(payload.countryId, payload.regionId)
                .then(rivers => {
                    let region = getRegionFromTree(payload.countryId, payload.regionId);
                    if (region) {
                        Vue.set(region, "rivers", rivers)
                    }
                });
        },

        hideRegionSubentities(state, payload) {
            var region = getRegionFromTree(payload.countryId, payload.regionId);
            if (region) {
                Vue.delete(region, "rivers");
            }
        },

        showRiverSubentities(state, payload) {
            getSpots(payload.riverId).then(spots => {
                let river = getRiverFromTree(payload.countryId, payload.regionId, payload.riverId);
                if (river) {
                    Vue.set(river, "spots", spots)
                }
            });
        },
        hideRiverSubentities(state, payload) {
            var river = getRiverFromTree(payload.countryId, payload.regionId, payload.riverId);
            Vue.delete(river, "spots")
        },

        setRiverEditorPageMode(state, mode) {
            state.rivereditorstate.pageMode = mode;
        },
        setRiverEditorVisible(state, visible) {
            state.rivereditorstate.visible = visible;
        },
        setSpotEditorState(state, payload) {
            state.spoteditorstate = payload;
        },
        setSpotEditorEditMode(state, payload) {
            state.spoteditorstate.editMode = payload;
        },
        setSpotEditorVisible(state, visible) {
            state.spoteditorstate.visible = visible;
        },
        setSpotImages(state, images) {
            state.spoteditorstate.images = images;
        },
        setSpotSchemas(state, schemas) {
            state.spoteditorstate.schemas = schemas;
        },
        setSpotVideos(state, videos) {
            state.spoteditorstate.videos = videos;
        },
        setErrMsg(state, msg) {
            state.errMsg = msg;
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

export function initEditor() {
    return init(EditorPage)
}

export function initDocsIntegration() {
    return init(DocsIntegrationPage)
}

export function initMap() {
    return init(MapPage)
}

export function initLog() {
    return init(LogPage)
}

export function initUsers() {
    return init(UsersPage)
}

export function initTech() {
    return init(TechPage)
}

export function initAbout() {
    return init(AboutPage)
}

export function initDocs() {
    return init(DocsPage)
}

export function initDashboard() {
    return init(DashboardPage)
}

function init(page) {
    Vue.component('v-select', vSelect);
    Vue.component('gallery', VueGallery);
    Vue.component('file-upload', FileUpload);
    Vue.component('datepicker', Datepicker);
    Vue.component('image-rating', ImageRating);
    Vue.use(TabsPlugin);

    Vue.filter('formatDateTimeStr', function(value) {
        if (value) {
            return moment(String(value)).format('YYYY-MM-DD HH:mm:ss')
        }
    });

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
        render: h => h(page),
    });
}

export function getApp() {
    return app
}
