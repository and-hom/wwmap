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

<style>
.row label {
  margin-right: 14px;
}
</style>

<script>
import {getAuthorizedUserInfoOrNull, hasRole, ROLE_ADMIN, ROLE_EDITOR} from "../auth";

function loadMapWhenDivIsReady(t) {
        if ($('#wwmap-container').outerWidth()) {
            wwmap.initWWMap("wwmap-container", "wwmap-rivers", {
                catalogLinkType: "wwmap",
                userInfoFunction: getAuthorizedUserInfoOrNull,
            })
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
        data() {
            return {
                map: null,
                userInfo: null,
            }
        },
    }
</script>
