<template>
    <li class="menu-item country-menu-item"><a
            href="javascript:void(0);" v-on:click='changeExpandState(); return false;' :class="countryClass()">{{
        country.title }}</a>
        <ul class="menu-items">
            <region v-bind:key="region.id" v-bind:region="region" v-bind:country="country" v-for="region of regions"/>
        </ul>
        <ul class="menu-items">
            <river v-bind:key="river.id" v-bind:river="river" v-bind:country="country" v-for="river of rivers"/>
        </ul>
    </li>
</template>

<script type="text/javascript">
    import {store} from '../main'
    import {COUNTRY_ACTIVE_ENTITY_LEVEL, getActiveEntityLevel, isActiveEntity, setActiveEntityUrlHash} from '../editor'

    module.exports = {
        props: ['country'],
        created: function () {
            if (isActiveEntity(this.country.id)) {
                store.commit('showCountrySubentities', this.country.id);
                if (getActiveEntityLevel() == COUNTRY_ACTIVE_ENTITY_LEVEL) {
                    this.selectCountry()
                }
            }
        },
        computed: {
            regions: function () {
                if (store.state.treePath[this.country.id] != null) {
                    return Array.from(store.state.treePath[this.country.id].regions.values())
                }
                return []
            },
            rivers: function () {
                if (store.state.treePath[this.country.id] != null) {
                    return Array.from(store.state.treePath[this.country.id].rivers.values())
                }
                return []
            },
        },
        data: function () {
            return {
                changeExpandState: function () {
                    let t = this;
                    store.commit('onTreeSwitch', function () {
                        if (store.state.treePath[t.country.id]) {
                            store.commit('hideCountrySubentities', t.country.id);
                        } else {
                            store.commit('showCountrySubentities', t.country.id);
                        }
                        t.selectCountry();
                    });
                    return false
                },
                selectCountry: function () {
                    setActiveEntityUrlHash(this.country.id);

                    store.commit('setActiveEntityState', this.country.id);
                    store.commit('hideAll');
                    store.commit('selectCountry', {country: this.country});
                },
                countryClass: function () {
                    if (this.country.id == store.state.selectedCountry) {
                        return "title-link btn btn-outline-danger"
                    } else {
                        return "title-link btn btn-outline-success"
                    }
                }
            }
        }
    }
</script>