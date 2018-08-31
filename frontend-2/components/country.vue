<template>
    <li class="menu-item country-menu-item"><a
            href="#" v-on:click='changeExpandState()' class="title-link btn btn-outline-success">{{ country.title }}</a>
        <ul>
            <region v-bind:key="region.id" v-bind:region="region" v-for="region in regions"/>
        </ul>
        <ul>
            <li>
                <ul>
                    <river v-bind:key="river.id" v-bind:river="river" v-for="river in rivers"/>
                </ul>
            </li>
        </ul>
    </li>
</template>

<script type="text/javascript">
    module.exports = {
        props: ['country'],
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
                changeExpandState:function(){
                    if (this.rivers.length==0 && this.regions.length==0) {
                        this.expand();
                    } else {
                        this.collapse();
                    }
                }
            }
        }
    }
</script>