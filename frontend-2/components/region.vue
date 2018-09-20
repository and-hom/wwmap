<template>
    <li class="menu-item region-menu-item"><a href="javascript:void(0);" v-on:click='changeExpandState();selectRegion();return false;'
                                              class="title-link btn btn-outline-secondary">{{ region.title }}</a>
        <ul>
            <river v-bind:key="river.id" v-bind:river="river" :region="region" :country="country" v-for="river in rivers"/>
        </ul>
    </li>
</template>

<script>
    module.exports = {
        props: ['region', 'country'],
        created: function() {
            if (isActiveEntity(this.country.id, this.region.id)) {
                this.expand()
                if (getActiveEntityLevel()==REGION_ACTIVE_ENTITY_LEVEL) {
                    this.selectRegion()
                }
            } else {
                this.collapse()
            }
        },
        data: function () {
            return {
                rivers: [],
                expand:function () {
                    this.rivers = getRiversByRegion(-1, this.region.id)
                },
                collapse:function () {
                    this.rivers=[]
                },
                changeExpandState:function(){
                    if (this.rivers.length==0) {
                        this.expand();
                    } else {
                        this.collapse();
                    }
                },
                selectRegion:function() {
                    setActiveEntity(this.country.id, this.region.id)

                    app.spoteditorstate.visible = false
                    app.rivereditorstate.visible=false;
                    app.regioneditorstate.visible = false;

                    app.regioneditorstate.region = getRegion(this.region.id)
                    app.regioneditorstate.country = this.country
                    app.regioneditorstate.editMode = false;
                    app.regioneditorstate.visible = true
                },
            }
        },
    }
</script>