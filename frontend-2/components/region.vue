<template>
    <li class="menu-item region-menu-item"><a href="#" v-on:click='changeExpandState();selectRegion();'
                                              class="title-link btn btn-outline-secondary">{{ region.title }}</a>
        <ul>
            <river v-bind:key="river.id" v-bind:river="river" v-for="river in rivers"/>
        </ul>
    </li>
</template>

<script>
    module.exports = {
        props: ['region'],
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
                    app.spoteditorstate.visible = false
                    app.rivereditorstate.visible=false;
                    app.regioneditorstate.visible = false;

                    app.regioneditorstate.region = getRegion(this.region.id)
                    app.regioneditorstate.editMode = false;
                    app.regioneditorstate.visible = true
                },
            }
        },
    }
</script>