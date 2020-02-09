<template>
    <page link="map.htm">
        <div class="container-fluid">
            <div class="row">
                <div class="col-10">
                    <div id="wwmap-container" style="width:100%; height: 700px; padding-bottom:17px;"></div>
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

    function loadMapWhenDivIsReady() {
        if ($('#wwmap-container').outerWidth()) {
            wwmap.initWWMap("wwmap-container", "wwmap-rivers", {
                catalogLinkType: "wwmap",
                userInfoFunction: getAuthorizedUserInfoOrNull,
            })
        } else {
            console.log("#div-container is not ready yet: has no offsetWidth");
            setTimeout(loadMapWhenDivIsReady, 100);
        }
    }

    export default {
        mounted: loadMapWhenDivIsReady,
        data() {
            return {}
        },
    }
</script>
