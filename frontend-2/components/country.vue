<template>
    <li class="menu-item country-menu-item"><a
            href="javascript:void(0);" v-on:click='changeExpandState()' class="title-link btn btn-outline-success">{{ country.title }}</a>
        <ul>
            <region v-bind:key="region.id" v-bind:region="region" v-bind:country="country" v-for="region in regions"/>
        </ul>
        <ul>
            <li>
                <ul>
                    <river v-bind:key="river.id" v-bind:river="river" v-bind:country="country" v-for="river in rivers"/>
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
                this.expand()
            } else {
                this.collapse()
            }
        },
        data: function() {
            return {
                regions: [],
                rivers: [],
                expand:function () {
                    this.regions = getRegions(this.country.id)
                    this.rivers = getRiversByCountry(this.country.id)
                },
                collapse:function () {
                    this.regions=[]
                    this.rivers=[]
                },
                changeExpandState:function() {
                    setActiveEntity(this.country.id)
                    if (this.rivers.length==0 && this.regions.length==0) {
                        this.expand();
                    } else {
                        this.collapse();
                    }
                    return false
                },
            }
        }
    }
</script>