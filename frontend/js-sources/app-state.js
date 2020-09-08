import Vue from 'vue'
import Vuex from 'vuex'

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
    getTransfers,
    nvlReturningId,
    REGION_ACTIVE_ENTITY_LEVEL,
    RIVER_ACTIVE_ENTITY_LEVEL,
    setActiveEntityUrlHash,
    SPOT_ACTIVE_ENTITY_LEVEL
} from './editor'
import {getAuthorizedUserInfoOrNull} from "./auth";
import {sensors} from "./sensors";


export function getById(arr, id) {
    if (!arr) {
        return null
    }
    var filtered = arr.filter(function (x) {
        return x.id === id
    });
    if (filtered.length > 0) {
        return filtered[0]
    }
    return null
}

export function getRiverFromTree(countryId, regionId, riverId) {
    let country = getById(store.state.treePath, countryId);
    if (!country) {
        return null
    }

    if (regionId && regionId > 0) {
        var region = getById(country.regions, regionId);
        if (!region) {
            return null
        }

        return getById(region.rivers, riverId)
    } else {
        return getById(country.rivers, riverId)
    }
}

export function getRegionFromTree(countryId, regionId) {
    let country = getById(store.state.treePath, countryId);
    if (!country) {
        return null;
    }

    return regionId && regionId > 0 ? getById(country.regions, regionId) : null;
}


export function getSpotFromTree(countryId, regionId, riverId, spotId) {
    let river = getRiverFromTree(countryId, regionId, riverId);
    return river ? getById(river.spots, spotId) : null;
}


export function getSpotsFromTree(countryId, regionId, riverId) {
    var river;
    if (regionId && regionId > 0) {
        var region = getById(getById(store.state.treePath, countryId).regions, regionId);
        river = getById(region.rivers, riverId)
    } else {
        river = getById(getById(store.state.treePath, countryId).rivers, riverId)
    }
    return river.spots ? river.spots : []
}

export function navigateToSpot(spotId, edit) {
    getSpot(spotId).then(spot => {
        let countryId = spot.river.region.country_id;
        let regionId = nvlReturningId(spot.river.region);
        let riverId = spot.river.id;

        store.commit('onTreeSwitch', function () {
            setActiveEntityUrlHash(countryId, regionId, riverId, spotId);

            store.commit('setTreeSelection', {
                countryId: countryId,
                regionId: regionId,
                riverId: riverId,
                spotId: spotId
            });

            store.commit('showSpotPage', {
                country: {"id": countryId},
                region: spot.river.region,
                river: spot.river,
                spotId: spotId,
                editMode: edit,
            });

            store.dispatch('showRiverSubentities', {
                countryId: countryId,
                regionId: regionId,
                riverId: riverId
            });
        });
    });
}


Vue.use(Vuex);
export const store = new Vuex.Store({
    state: {
        spoteditorstate: {
            visible: false,
            editMode: false,
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
        treePath: [],
        selectedCountry: getActiveId(COUNTRY_ACTIVE_ENTITY_LEVEL),
        selectedRegion: getActiveId(REGION_ACTIVE_ENTITY_LEVEL),
        selectedRiver: getActiveId(RIVER_ACTIVE_ENTITY_LEVEL),
        selectedSpot: getActiveId(SPOT_ACTIVE_ENTITY_LEVEL),
        sensors: sensors,
    },
    getters: {},
    actions: {
        reloadCountrySubentities(context, id) {
            return Promise.all([getRiversByCountry(id), getRegions(id)]).then(result => {
                context.commit('showCountrySubentities', {
                    id: id,
                    rivers: result[0],
                    regions: result[1],
                })
            })
        },
        showCountrySubentities(context, id) {
            let country = getById(context.state.treePath, id);
            if (country && !country.regions && !country.rivers) {
                return context.dispatch('reloadCountrySubentities', id)
            }
        },

        reloadRegionSubentities(context, payload) {
            getRiversByRegion(payload.countryId, payload.regionId).then(rivers => {
                return context.commit('showRegionSubentities', {
                    countryId: payload.countryId,
                    regionId: payload.regionId,
                    rivers: rivers
                });
            });
        },

        showRegionSubentities(context, payload) {
            return context.dispatch('showCountrySubentities', payload.countryId).then(_ => {
                if (!payload.regionId || payload.regionId <= 0) {
                    return
                }

                let region = getRegionFromTree(payload.countryId, payload.regionId);
                if (!region.rivers) {
                    return context.dispatch('reloadRegionSubentities', payload)
                }
            });
        },

        reloadRiverSubentities(context, payload) {
            getSpots(payload.riverId).then(spots => {
                return context.commit('showRiverSubentities', {
                    countryId: payload.countryId,
                    regionId: payload.regionId,
                    riverId: payload.riverId,
                    spots: spots
                });
            });
        },

        showRiverSubentities(context, payload) {
            context.dispatch('showRegionSubentities', payload).then(_ => {
                let river = getRiverFromTree(payload.countryId, payload.regionId, payload.riverId);
                if (river) {
                    return context.dispatch('reloadRiverSubentities', payload)
                }
            })
        },
    },
    mutations: {
        setTreePath(state, treePath) {
            state.treePath = treePath;
        },
        showCountryPage(state, payload) {
            hideAllEditorPanes(state);

            state.countryeditorstate.country = payload.country;
            state.countryeditorstate.editMode = false;
            state.countryeditorstate.visible = true
        },
        showRegionPage(state, payload) {
            hideAllEditorPanes(state);

            getRegion(payload.regionId).then(region => {
                state.regioneditorstate.region = region;
                state.regioneditorstate.country = payload.country;
                state.regioneditorstate.editMode = false;
                state.regioneditorstate.visible = true;
            });
        },
        showRiverPage(state, payload) {
            hideAllEditorPanes(state);

            Promise.all([getRiver(payload.riverId), getReports(payload.riverId), getTransfers(payload.riverId)])
                .then(riverData => {
                    let river = riverData[0];
                    state.rivereditorstate.river = river;
                    state.rivereditorstate.pageMode = 'view';
                    state.rivereditorstate.country = payload.country;
                    state.rivereditorstate.region = payload.region;

                    state.rivereditorstate.reports = riverData[1];
                    state.rivereditorstate.transfers = riverData[2];

                    state.rivereditorstate.visible = true;
                });
        },
        showSpotPage(state, payload) {
            hideAllEditorPanes(state);

            getSpot(payload.spotId).then(spot => {
                state.spoteditorstate.country = payload.country;
                state.spoteditorstate.region = payload.region;
                state.spoteditorstate.river = payload.river;
                state.spoteditorstate.spot = spot;
                state.spoteditorstate.editMode = payload.editMode ? true : false;
                state.spoteditorstate.visible = true;
            });
        },

        setTreeSelection(state, payload) {
            state.selectedSpot = payload.spotId;
            state.selectedRiver = payload.riverId;
            state.selectedRegion = payload.regionId;
            state.selectedCountry = payload.countryId;
        },

        newRegion(state, payload) {
            state.spoteditorstate.visible = false;
            state.rivereditorstate.visible = false;
            state.regioneditorstate.visible = false;
            state.countryeditorstate.visible = false;

            state.regioneditorstate.visible = true;
            state.regioneditorstate.editMode = true;
            state.regioneditorstate.region = {
                id: 0,
                title: '',
                country_id: payload.country.id,
                fake: false,
                has_rivers: false,
            };
            state.regioneditorstate.country = payload.country;
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
            state.rivereditorstate.reports = [];
            state.rivereditorstate.transfers = [];
        },

        showCountrySubentities(state, payload) {
            let country = getById(store.state.treePath, payload.id);
            Vue.set(country, "rivers", payload.rivers);
            Vue.set(country, "regions", payload.regions);
        },
        hideCountrySubentities(state, id) {
            let country = getById(store.state.treePath, id);
            Vue.delete(country, "rivers");
            Vue.delete(country, "regions");
        },

        showRegionSubentities(state, payload) {
            let region = getRegionFromTree(payload.countryId, payload.regionId);
            if (region) {
                Vue.set(region, "rivers", payload.rivers)
            }
        },

        hideRegionSubentities(state, payload) {
            var region = getRegionFromTree(payload.countryId, payload.regionId);
            if (region) {
                Vue.delete(region, "rivers");
            }
        },

        showRiverSubentities(state, payload) {
            let river = getRiverFromTree(payload.countryId, payload.regionId, payload.riverId);
            if (river) {
                Vue.set(river, "spots", payload.spots)
            }
        },

        hideRiverSubentities(state, payload) {
            let river = getRiverFromTree(payload.countryId, payload.regionId, payload.riverId);
            if (river) {
                Vue.delete(river, "spots")
            }
        },

        setRegionEditorEditMode(state, payload) {
            state.regioneditorstate.editMode = payload;
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

// don't call outside the store mutations
function hideAllEditorPanes(state) {
    state.spoteditorstate.visible = false;
    state.rivereditorstate.visible = false;
    state.regioneditorstate.visible = false;
    state.countryeditorstate.visible = false;
}