<template>
    <div>
        <span v-for="lnk in links">
            <a href="javascript:void(0);" @click="lnk.click">/&nbsp;{{lnk.text}}&nbsp;</a>
        </span>
    </div>
</template>

<script>
    import {setActiveEntityUrlHash} from '../editor'
    import {store} from "../app-state";

    module.exports = {
        props: {
            country: {
                type: Object,
                required: false,
            },
            region: {
                type: Object,
                required: false,
            },
            river: {
                type: Object,
                required: false,
            },
        },
        computed: {
            links: {
                get: function () {
                    let val = [];
                    let t = this;

                    if (t.country) {
                        val.push({
                            text: t.country.title,
                            click: function() {
                                t.navigateToCountry()
                            }
                        });

                        if (t.region && !t.region.fake && t.region.id && t.region.id > 0) {
                            val.push({
                                text: t.region.title,
                                click: function() {
                                    t.navigateToRegion()
                                }
                            });
                        }

                        if (t.river) {
                            val.push({
                                text: t.river.title,
                                click: function() {
                                    t.navigateToRiver()
                                }
                            });
                        }
                    }

                    return val;
                }
            }
        },
        data: function () {
            return {
                navigateToCountry: function () {
                    let t = this;

                    setActiveEntityUrlHash(t.country.id);

                    store.commit('setTreeSelection', {
                        countryId: t.country.id,
                    });

                    store.commit('showCountryPage', {
                        country: t.country,
                    });

                    store.dispatch('showCountrySubentities', {
                        countryId: t.country.id,
                    });
                },
                navigateToRegion: function () {
                    let t = this;

                    setActiveEntityUrlHash(t.country.id, t.region.id)

                    store.commit('setTreeSelection', {
                        countryId: t.country.id,
                        regionId: t.region.id
                    });

                    store.commit('showRegionPage', {
                        country: t.country,
                        regionId: t.region.id,
                    });

                    store.dispatch('showRegionSubentities', {
                        countryId: t.country.id,
                        regionId: t.region.id,
                    });
                },
                navigateToRiver: function () {
                    let t = this;

                    setActiveEntityUrlHash(t.country.id, t.region ? t.region.id : 0, t.river.id)

                    store.commit('setTreeSelection', {
                        countryId: t.country.id,
                        regionId: t.region ? t.region.id : 0,
                        riverId: t.river.id,
                    });

                    store.commit('showRiverPage', {
                        country: t.country,
                        region: t.region,
                        riverId: t.river.id,
                    });

                    store.dispatch('showRiverSubentities', {
                        countryId: t.country.id,
                        regionId: t.region ? t.region.id : 0,
                        riverId: t.river.id,
                    });
                },
            }
        }
    }
</script>