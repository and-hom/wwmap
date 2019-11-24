<template>
    <li class="menu-item river-menu-item"><a href="javascript:void(0);" v-on:click='changeExpandState();'
                                             :class="riverClass()"><img v-if="!river.visible" style="margin-right: 6px;" src="img/invisible.png"/>{{ river.title }}</a>
        <ul>
            <li class="menu-item spot-menu-item" v-on:click.stop="selectSpot(spot)"
                v-for="spot in river.spots"><a href="javascript:void(0);"
                :class="spotClass(spot)">{{spot.title}}</a>
            </li>
        </ul>
    </li>
</template>

<script>
    module.exports = {
        props: ['river', 'region', 'country'],
        created: function() {
            var riverSelected = isActiveEntity(this.country.id, nvlReturningId(this.region), this.river.id)

            if (riverSelected) {
                showRiverTree(this.country.id, nvlReturningId(this.region), this.river.id)
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
                    app.onTreeSwitch(function () {
                        if (t.river.spots) {
                            Vue.delete(t.river, "spots")
                        } else {
                            Vue.set(t.river, "spots", getSpots(t.river.id))
                        }
                        t.selectRiver();
                    });
                },
                selectSpot:function(spot) {
                    let t = this;
                    app.onTreeSwitch(function () {
                        setActiveEntity(t.country.id, nvlReturningId(t.region), t.river.id, spot.id);
                        setActiveEntityState(t.country.id, nvlReturningId(t.region), t.river.id, spot.id);

                        app.spoteditorstate.visible = false;
                        app.rivereditorstate.visible=false;
                        app.regioneditorstate.visible = false;
                        app.countryeditorstate.visible = false;

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
                    setActiveEntityState(this.country.id, nvlReturningId(this.region), this.river.id);

                    app.spoteditorstate.visible = false;
                    app.rivereditorstate.visible=false;
                    app.regioneditorstate.visible = false;
                    app.countryeditorstate.visible = false;

                    selectRiver(this.country, this.region, this.river.id);

                    return false
                },
                riverClass: function() {
                    if (this.river.id == app.selectedRiver) {
                        return "title-link btn btn-outline-danger"
                    } else {
                        return "title-link btn btn-outline-info"
                    }
                },
                spotClass: function(spot) {
                    if (spot.id == app.selectedSpot) {
                        return "title-link btn btn-outline-danger"
                    } else {
                        return "title-link btn btn-outline-primary"
                    }
                }
            }
        },
    }
</script>