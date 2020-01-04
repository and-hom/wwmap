<template>
    <li class="menu-item river-menu-item"><a href="javascript:void(0);" v-on:click='changeExpandState();'
                                             :class="riverClass()"><img v-if="!river.visible" style="margin-right: 6px;"
                                                                        src="img/invisible.png"/>{{ river.title }}</a>
        <ul class="menu-items">
            <li class="menu-item spot-menu-item" v-on:click.stop="selectSpot(spot)"
                v-for="spot in river.spots"><a href="javascript:void(0);"
                                               :class="spotClass(spot)">{{spot.title}}</a>
            </li>
        </ul>
    </li>
</template>

<script>
    import {
        getActiveEntityLevel,
        getActiveId,
        isActiveEntity,
        nvlReturningId,
        RIVER_ACTIVE_ENTITY_LEVEL,
        setActiveEntityUrlHash,
        SPOT_ACTIVE_ENTITY_LEVEL
    } from "../editor";
    import {store} from '../main'

    module.exports = {
        props: ['river', 'region', 'country'],
        created: function () {
            var riverSelected = isActiveEntity(this.country.id, nvlReturningId(this.region), this.river.id);

            if (riverSelected) {
                store.commit('showRiverSubentities', {
                    countryId: this.country.id,
                    regionId: nvlReturningId(this.region),
                    riverId: this.river.id,
                });
                if (getActiveEntityLevel() == RIVER_ACTIVE_ENTITY_LEVEL) {
                    this.selectRiver()
                } else if (getActiveEntityLevel() == SPOT_ACTIVE_ENTITY_LEVEL) {
                    let selectedSpotId = getActiveId(SPOT_ACTIVE_ENTITY_LEVEL);
                    this.selectSpot({id: selectedSpotId})
                }
            }
        },
        data: function () {
            return {
                spots: [],
                collapse: function () {
                    this.spots = []
                },
                changeExpandState: function () {
                    let t = this;
                    store.commit('onTreeSwitch', function () {
                        let idsPath = {
                            countryId: t.country.id,
                            regionId: nvlReturningId(t.region),
                            riverId: t.river.id
                        };
                        store.commit(t.river.spots ? 'hideRiverSubentities' : 'showRiverSubentities', idsPath);
                        t.selectRiver();
                    });
                },
                selectSpot: function (spot) {
                    let t = this;
                    store.commit('onTreeSwitch', function () {
                        setActiveEntityUrlHash(t.country.id, nvlReturningId(t.region), t.river.id, spot.id);
                        store.commit('setActiveEntityState', {
                            countryId: t.country.id,
                            regionId: nvlReturningId(t.region),
                            riverId: t.river.id,
                            spotId: spot.id
                        });

                        store.commit('hideAll');
                        store.commit('selectSpot', {
                            country: t.country,
                            region: t.region,
                            river: t.river,
                            spotId: spot.id,
                        });
                    });

                    return false
                },
                selectRiver: function () {
                    setActiveEntityUrlHash(this.country.id, nvlReturningId(this.region), this.river.id);
                    store.commit('setActiveEntityState', {
                        countryId: this.country.id,
                        regionId: nvlReturningId(this.region),
                        riverId: this.river.id,
                        spotId: null
                    });
                    store.commit('hideAll');
                    store.commit('selectRiver', {
                        country: this.country,
                        region: this.region,
                        riverId: this.river.id
                    });

                    return false
                },
                riverClass: function () {
                    return this.river.id == store.state.selectedRiver
                        ? "title-link btn btn-outline-danger"
                        : "title-link btn btn-outline-info";
                },
                spotClass: function (spot) {
                    return spot.id == store.state.selectedSpot
                        ? "title-link btn btn-outline-danger"
                        : "title-link btn btn-outline-primary";
                }
            }
        },
    }
</script>