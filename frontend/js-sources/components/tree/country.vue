<template>
    <li class="menu-item country-menu-item"><a
            href="javascript:void(0);" v-on:click='changeExpandState(); return false;' :class="countryClass()">{{
        country.title }}</a>
        <ul class="menu-items">
            <region v-bind:key="region.id" v-bind:region="region" v-bind:country="country" v-for="region of country.regions"/>
        </ul>
        <ul class="menu-items">
            <river v-bind:key="river.id" v-bind:river="river" v-bind:country="country" v-for="river of country.rivers"/>
        </ul>
    </li>
</template>

<script type="text/javascript">
    import {store, getById} from '../../app-state'
    import {COUNTRY_ACTIVE_ENTITY_LEVEL, getActiveEntityLevel, isActiveEntity, setActiveEntityUrlHash} from '../../editor'

    module.exports = {
        props: ['country'],
        created: function () {
            if (isActiveEntity(this.country.id)) {
                store.dispatch('reloadCountrySubentities', this.country.id);
                if (getActiveEntityLevel() == COUNTRY_ACTIVE_ENTITY_LEVEL) {
                    this.onSelectCountry()
                }
            }
        },
        data: function () {
            return {
                changeExpandState: function () {
                    let t = this;
                    store.commit('onTreeSwitch', function () {
                        let country = getById(store.state.treePath, t.country.id);
                        if (country && (country.regions || country.rivers)) {
                            store.commit('hideCountrySubentities', t.country.id);
                        } else {
                            store.dispatch('reloadCountrySubentities', t.country.id);
                        }
                        t.onSelectCountry();
                    });
                    return false
                },
                onSelectCountry: function () {
                    setActiveEntityUrlHash(this.country.id);

                    store.commit('setTreeSelection', {
                        countryId: this.country.id
                    });
                    store.commit('showCountryPage', {countryId: this.country.id});
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