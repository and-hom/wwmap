<template>
    <page link="map.htm">
        <div class="container-fluid">
            <div class="row">
                <div class="col-10">
                    <div id="wwmap-container" style="width:100%; height: 700px; padding-bottom:17px;"></div>
                    <input type="checkbox" id="show-unpublished" v-model="showUnpublished"/>
                    <label for="show-unpublished" :style="unpublishedLabelStyle">Показывать неопубликованное</label>
                </div>
                <div class="col-2">
                    <div id="wwmap-rivers" class="wwmap-river-menu"></div>
                </div>
            </div>
            <div class="row">
                <div class="col-12">
                </div>
            </div>
        </div>
    </page>
</template>

<script>
    import {getAuthorizedUserInfoOrNull} from "../auth";

    function loadMapWhenDivIsReady(t) {
        if ($('#wwmap-container').outerWidth()) {
            wwmap.initWWMap("wwmap-container", "wwmap-rivers", {
                catalogLinkType: "wwmap",
                userInfoFunction: getAuthorizedUserInfoOrNull,
            }).then(map => t.map = map)
                .catch(ex => console.error(ex));
        } else {
            console.log("#div-container is not ready yet: has no offsetWidth");
            setTimeout(() => loadMapWhenDivIsReady(t), 100);
        }
    }

    export default {
        mounted: function () {
            loadMapWhenDivIsReady(this);
        },
        watch: {
            showUnpublished: function (newValue) {
                this.unpublishedLabelStyle = newValue ? 'color:red; text-decoration: underline;' : '';
                this.map.setShowUnpublished(newValue);
            },
        },
        data() {
            return {
                showUnpublished: false,
                unpublishedLabelStyle: '',
                map: null,
            }
        },
    }
</script>
