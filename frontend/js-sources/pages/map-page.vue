<template>
    <page link="map.htm">
        <div class="container-fluid">
            <div class="row">
                <div class="col-10">
                    <div id="wwmap-container" style="width:100%; height: 700px; padding-bottom:17px;"></div>
                    <input v-if="canViewUnpublished" type="checkbox" id="show-unpublished" v-model="showUnpublished"/>
                    <label v-if="canViewUnpublished" for="show-unpublished" :style="unpublishedLabelStyle">Показывать неопубликованное</label>
                    <input v-if="canShowCamps" type="checkbox" id="show-camps" v-model="showCamps"/>
                    <label v-if="canShowCamps" for="show-camps">Показывать стоянки</label>
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
            }).then(map => {
              t.map = map;
              t.map.setOnBoundsChange(t.onBoundsChange)
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
        created: function (){
            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canViewUnpublished => this.canViewUnpublished = canViewUnpublished);
        },
        watch: {
            showUnpublished: function (newValue) {
                this.unpublishedLabelStyle = newValue ? 'color:red; text-decoration: underline;' : '';
                this.map.setShowUnpublished(newValue);
            },
            showCamps: function (newValue) {
              this.map.setShowCamps(newValue);
            },
        },
        data() {
            return {
                showUnpublished: false,
                showCamps: true,
                unpublishedLabelStyle: '',
                canViewUnpublished: false,
                canShowCamps: true,
                map: null,
            }
        },
      methods: {
        onBoundsChange(_, zoom) {
          this.canShowCamps = zoom >= 12;
        },
      }
    }
</script>
