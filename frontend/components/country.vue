<template>
    <li class="menu-item country-menu-item"><a
            href="javascript:void(0);" v-on:click='changeExpandState()' :class="countryClass()">{{ country.title }}</a>
        <ul>
            <region v-bind:key="region.id" v-bind:region="region" v-bind:country="country" v-for="region of regions"/>
        </ul>
        <ul>
            <li>
                <ul>
                    <river v-bind:key="river.id" v-bind:river="river" v-bind:country="country" v-for="river of rivers"/>
                </ul>
            </li>
        </ul>
    </li>
</template>

<script type="text/javascript">
    module.exports = {
        props: ['country'],
        created: function() {
            if (isActiveEntity(this.country.id)) {
                showCountrySubentities(this.country.id)
            }
        },
        computed: {
            regions: function() {
                if (app.treePath[this.country.id]!=null) {
                    return Array.from(app.treePath[this.country.id].regions.values())
                }
                return []
            },
            rivers: function() {
                if (app.treePath[this.country.id]!=null) {
                    return Array.from(app.treePath[this.country.id].rivers.values())
                }
                return []
            },
        },
        data: function() {
            return {
                changeExpandState:function() {
                    setActiveEntity(this.country.id)
                    setActiveEntityState(this.country.id)

                    if (app.treePath[this.country.id]) {
                        Vue.delete(app.treePath, this.country.id)
                    } else {
                        showCountrySubentities(this.country.id)
                    }
                    return false
                },
                countryClass: function() {
                    if (this.country.id == app.selectedCountry) {
                        return "title-link btn btn-outline-danger"
                    } else {
                        return "title-link btn btn-outline-success"
                    }
                }
            }
        }
    }
</script>