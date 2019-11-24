<template>
    <li class="menu-item region-menu-item"><a href="javascript:void(0);" v-on:click='changeExpandState();return false;'
                                              :class="regionClass()">{{ region.title }}</a>
        <ul>
            <river v-bind:key="river.id" v-bind:river="river" :region="region" :country="country" v-for="river in region.rivers"/>
        </ul>
    </li>
</template>

<script>
    module.exports = {
        props: ['region', 'country'],
        created: function() {
            if (isActiveEntity(this.country.id, this.region.id)) {
                showRegionTree(this.country.id, this.region.id)
                if (getActiveEntityLevel()==REGION_ACTIVE_ENTITY_LEVEL) {
                    this.selectRegion()
                }
            }
        },
        data: function () {
            return {
                changeExpandState: function () {
                    let t = this;
                    app.onTreeSwitch(function () {
                        if (t.region.rivers) {
                            Vue.delete(t.region, "rivers")
                        } else {
                            Vue.set(t.region, "rivers", getRiversByRegion(t.country.id, t.region.id))
                        }
                        t.selectRegion();
                    });
                },
                selectRegion:function() {
                    setActiveEntity(this.country.id, this.region.id)
                    setActiveEntityState(this.country.id, this.region.id)

                    app.spoteditorstate.visible = false
                    app.rivereditorstate.visible=false;
                    app.regioneditorstate.visible = false;
                    app.countryeditorstate.visible = false;

                    selectRegion(this.country, this.region.id);
                },
                regionClass: function() {
                    if (this.region.id == app.selectedRegion) {
                        return "title-link btn btn-outline-danger"
                    } else {
                        return "title-link btn btn-outline-secondary"
                    }
                }
            }
        },
    }
</script>