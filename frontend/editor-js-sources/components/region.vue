<template>
    <li class="menu-item region-menu-item"><a href="javascript:void(0);" v-on:click='changeExpandState();return false;'
                                              :class="regionClass()">{{ region.title }}</a>
        <ul>
            <river v-bind:key="river.id" v-bind:river="river" :region="region" :country="country" v-for="river in region.rivers"/>
        </ul>
    </li>
</template>

<script>
    import {getActiveEntityLevel, isActiveEntity, REGION_ACTIVE_ENTITY_LEVEL, setActiveEntityUrlHash,} from "../editor";
    import {store} from "../main";

    module.exports = {
        props: ['region', 'country'],
        created: function() {
            if (isActiveEntity(this.country.id, this.region.id)) {
                store.commit('showRegionSubentities', {
                    countryId: this.country.id,
                    regionId: this.region.id
                });
                if (getActiveEntityLevel()==REGION_ACTIVE_ENTITY_LEVEL) {
                    this.selectRegion()
                }
            }
        },
        data: function () {
            return {
                changeExpandState: function () {
                    let t = this;
                    store.commit('onTreeSwitch', function () {
                        let idsPath = {countryId: t.country.id, regionId: t.region.id};
                        store.commit(t.region.rivers ? 'hideRegionSubentities' : 'showRegionSubentities', idsPath);
                        t.selectRegion();
                    });
                    return false
                },
                selectRegion:function() {
                    setActiveEntityUrlHash(this.country.id, this.region.id);

                    store.commit('setActiveEntityState', {
                        countryId: this.country.id,
                        regionId: this.region.id
                    });
                    store.commit('hideAll');
                    store.commit('selectRegion', {
                        country: this.country,
                        regionId: this.region.id
                    });
                },
                regionClass: function() {
                    if (this.region.id == store.state.selectedRegion) {
                        return "title-link btn btn-outline-danger"
                    } else {
                        return "title-link btn btn-outline-secondary"
                    }
                }
            }
        },
    }
</script>