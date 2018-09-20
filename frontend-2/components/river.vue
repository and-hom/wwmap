<template>
    <li class="menu-item river-menu-item"><a href="javascript:void(0);" v-on:click='changeExpandState();selectRiver();'
                                             class="title-link btn btn-outline-info">{{ river.title }}</a>
        <ul>
            <li class="menu-item spot-menu-item" v-on:click.stop="selectSpot(spot)"
                v-for="spot in spots"><a href="javascript:void(0);"
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
                this.expand()
                if (getActiveEntityLevel()==RIVER_ACTIVE_ENTITY_LEVEL) {
                    this.selectRiver()
                } else if (getActiveEntityLevel()==SPOT_ACTIVE_ENTITY_LEVEL) {
                    var selectedSpotId = getActiveId(SPOT_ACTIVE_ENTITY_LEVEL)
                    selectedSpot = this.spots.filter(function(x){return x.id==selectedSpotId})
                    if (selectedSpot.length>0) {
                        this.selectSpot(selectedSpot[0])
                    }
                }
            } else {
                this.collapse()
            }
        },
        data: function () {
            return {
                spots: [],
                expand:function() {
                    this.spots = getSpots(this.river.id)
                },
                collapse:function () {
                    this.spots=[]
                },
                changeExpandState:function(){
                    if (this.spots.length==0) {
                        this.expand();
                    } else {
                        this.collapse();
                    }
                },
                selectSpot:function(spot) {
                    setActiveEntity(this.country.id, nvlReturningId(this.region), this.river.id, spot.id)

                    app.spoteditorstate.visible = false
                    app.rivereditorstate.visible=false;
                    app.regioneditorstate.visible = false;

                    app.spoteditorstate.visible=true;
                    app.spoteditorstate.editMode = false;
                    app.spoteditorstate.spot=getSpot(spot.id)

                    this.$forceUpdate()

                    return false
                },
                selectRiver:function() {
                    setActiveEntity(this.country.id, nvlReturningId(this.region), this.river.id)

                    app.spoteditorstate.visible = false
                    app.rivereditorstate.visible=false;
                    app.regioneditorstate.visible = false;

                    app.rivereditorstate.river = getRiver(this.river.id)
                    app.rivereditorstate.editMode = false;
                    app.rivereditorstate.reports=getReports(this.river.id)
                    app.rivereditorstate.visible = true

                    return false
                },
                spotClass: function(spot) {
                    var spotSelected = isActiveEntity(this.country.id, nvlReturningId(this.region), this.river.id, spot.id)
                    if (spotSelected) {
                        return "title-link btn btn-outline-danger"
                    } else {
                        return "title-link btn btn-outline-primary"
                    }
                }
            }
        },
    }
</script>