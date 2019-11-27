<template>
    <li class="menu-item river-menu-item"><a href="javascript:void(0);" v-on:click='changeExpandState();'
                                             :class="riverClass()"><img v-if="!river.visible" style="margin-right: 6px;" src="img/invisible.png"/>{{ river.title }}</a>
        <ul class="menu-items">
            <li class="menu-item spot-menu-item" v-on:click.stop="selectSpot(spot)"
                v-for="spot in river.spots"><a href="javascript:void(0);"
                :class="spotClass(spot)">{{spot.title}}</a>
            </li>
        </ul>
    </li>
</template>

<script>
    import {RIVER_ACTIVE_ENTITY_LEVEL, SPOT_ACTIVE_ENTITY_LEVEL, nvlReturningId, isActiveEntity, setActiveEntity, getActiveEntityLevel, getActiveId, getSpots} from "../editor";
    import {store, getSpotsFromTree} from '../main'

    module.exports = {
        props: ['river', 'region', 'country'],
        created: function() {
            var riverSelected = isActiveEntity(this.country.id, nvlReturningId(this.region), this.river.id)

            if (riverSelected) {
                store.commit('showRiverTree', this.country.id, nvlReturningId(this.region), this.river.id)
                if (getActiveEntityLevel()==RIVER_ACTIVE_ENTITY_LEVEL) {
                    this.selectRiver()
                } else if (getActiveEntityLevel()==SPOT_ACTIVE_ENTITY_LEVEL) {
                    var selectedSpotId = getActiveId(SPOT_ACTIVE_ENTITY_LEVEL)
                    var spots = getSpotsFromTree(this.country.id, nvlReturningId(this.region), this.river.id)
                    selectedSpot = spots.filter(function(x){return x.id==selectedSpotId})
                    if (selectedSpot.length>0) {
                        this.selectSpot(selectedSpot[0])
                    }
                }
            }
        },
        data: function () {
            return {
                spots: [],
                collapse:function () {
                    this.spots=[]
                },
                changeExpandState: function () {
                    let t = this;
                    store.commit('onTreeSwitch', function () {
                        if (t.river.spots) {
                            delete t.river["spots"];
                        } else {
                            getSpots(t.river.id).then(spots => t.river.spots = spots);
                        }
                        t.selectRiver();
                    });
                },
                selectSpot:function(spot) {
                    let t = this;
                    store.commit('onTreeSwitch', function () {
                        setActiveEntity(t.country.id, nvlReturningId(t.region), t.river.id, spot.id);
                        store.commit('setActiveEntityState', t.country.id, nvlReturningId(t.region), t.river.id, spot.id);

                        store.commit('hideAll');

                        app.spoteditorstate.editMode = false;
                        app.spoteditorstate.spot=getSpot(spot.id);
                        app.spoteditorstate.country = t.country;
                        app.spoteditorstate.region = t.region;
                        app.spoteditorstate.visible=true;
                    });

                    return false
                },
                selectRiver:function() {
                    setActiveEntity(this.country.id, nvlReturningId(this.region), this.river.id);
                    store.commit('setActiveEntityState', this.country.id, nvlReturningId(this.region), this.river.id);
                    store.commit('hideAll');
                    // store.commit('selectRiver', this.country, this.region, this.river.id);

                    return false
                },
                riverClass: function() {
                    if (this.river.id == store.state.selectedRiver) {
                        return "title-link btn btn-outline-danger"
                    } else {
                        return "title-link btn btn-outline-info"
                    }
                },
                spotClass: function(spot) {
                    if (spot.id == store.state.selectedSpot) {
                        return "title-link btn btn-outline-danger"
                    } else {
                        return "title-link btn btn-outline-primary"
                    }
                }
            }
        },
    }
</script>