<template>
    <li class="menu-item river-menu-item"><a href="javascript:void(0);" v-on:click='changeExpandState();selectRiver();'
                                             class="title-link btn btn-outline-info">{{ river.title }}</a>
        <ul>
            <li class="menu-item spot-menu-item" v-on:click.stop="selectSpot(spot)" v-for="spot in spots"><a href="javascript:void(0);"
                                                                                                             class="title-link btn btn-outline-primary">{{spot.title}}</a>
            </li>
        </ul>
    </li>
</template>

<script>
    module.exports = {
        props: ['river', 'region', 'country'],
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
            }
        },
    }
</script>