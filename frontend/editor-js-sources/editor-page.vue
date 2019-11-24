<template>
    <div>
        <ask id="close-spot-editor" title="Отменить редактирование?"
             msg='Открыт редактор порога. Там могут быть ваши изменения. Для сохранения воспользуйтесь кнопкой "Сохранить" сверху. Закрыть редактор и сбросить изменения?'
             :ok-fn="function() { spoteditorstate.editMode = false; if(closeCallback) {closeCallback()}}"></ask>
        <ask id="close-river-editor" title="Отменить редактирование?"
             msg='Открыт редактор реки. Там могут быть ваши изменения. Для сохранения воспользуйтесь кнопкой "Сохранить" сверху. Закрыть редактор и сбросить изменения?'
             :ok-fn="function() { rivereditorstate.pageMode = 'view'; if(closeCallback) {closeCallback()} }"></ask>

        <page link="editor.htm">
            <div class="container-fluid" style="margin-top: 20px;">
                <div class="row">
                    <div class="col-3" id="left-menu">
                        <ul>
                            <country v-bind:key="country.id" v-bind:country="country" v-for="country in countries"/>
                        </ul>
                    </div>
                    <div id="editor-pane" class="col-9" style="bgcolor:red;">
                        <transition name="fade">
                            <div class="alert alert-danger" role="alert" v-if="errMsg">
                                {{errMsg}}
                            </div>
                        </transition>
                        <div>
                            <country-editor v-if="countryeditorstate.visible"
                                            v-bind:country="countryeditorstate.country"/>
                        </div>
                        <div>
                            <region-editor v-if="regioneditorstate.visible" v-bind:region="regioneditorstate.region"
                                           v-bind:country="regioneditorstate.country"/>
                        </div>
                        <div>
                            <river-editor v-if="rivereditorstate.visible" v-bind:initial-river="rivereditorstate.river"
                                          v-bind:reports="rivereditorstate.reports"
                                          v-bind:country="rivereditorstate.country"
                                          v-bind:region="rivereditorstate.region"
                                          v:sensors="sensors"/>
                        </div>
                        <div>
                            <spot-editor v-if="spoteditorstate.visible" v-bind:initial-spot="spoteditorstate.spot"
                                         v-bind:country="spoteditorstate.country"
                                         v-bind:region="spoteditorstate.region"/>
                        </div>
                    </div>
                </div>
            </div>
        </page>
    </div>
</template>

<script>
    import {
        getActiveId,
        getAllRegions,
        getCountries,
        COUNTRY_ACTIVE_ENTITY_LEVEL,
        REGION_ACTIVE_ENTITY_LEVEL,
        RIVER_ACTIVE_ENTITY_LEVEL,
        SPOT_ACTIVE_ENTITY_LEVEL
    } from './editor'
    import {getAuthorizedUserInfoOrNull} from './auth'
    import {sensors} from './sensors'


    export default {
        data() {
            return {
                countries: getCountries(),
                regions: getAllRegions(),
                "spoteditorstate": {
                    "visible": false,
                    "editMode": false,
                    "images": [],
                    "schemas": []
                },
                "rivereditorstate": {
                    "visible": false,
                    "pageMode": 'view'
                },
                "regioneditorstate": {
                    "visible": false,
                    "editMode": false
                },
                "countryeditorstate": {
                    "visible": false,
                    "editMode": false
                },
                userInfo: getAuthorizedUserInfoOrNull(),
                treePath: {},
                selectedCountry: getActiveId(COUNTRY_ACTIVE_ENTITY_LEVEL),
                selectedRegion: getActiveId(REGION_ACTIVE_ENTITY_LEVEL),
                selectedRiver: getActiveId(RIVER_ACTIVE_ENTITY_LEVEL),
                selectedSpot: getActiveId(SPOT_ACTIVE_ENTITY_LEVEL),
                sensors: sensors,
                errMsg: "",
                closeCallback: function () {
                },
                onTreeSwitch: function (callback) {
                    this.errMsg = null;
                    this.closeCallback = callback;
                    let t = this;

                    if (!this.spoteditorstate.editMode && this.rivereditorstate.pageMode != 'edit' && this.rivereditorstate.pageMode != 'batch-edit') {
                        callback();
                        return;
                    }

                    if (this.spoteditorstate.editMode) {
                        let spotEditorCloseDialog = $('#close-spot-editor');
                        spotEditorCloseDialog.on('hidden.bs.modal', function (e) {
                            t.closeCallback = function () {

                            };
                        });
                        spotEditorCloseDialog.modal();
                    }
                    if (this.rivereditorstate.pageMode == 'edit' || this.rivereditorstate.pageMode == 'batch-edit') {
                        let riverEditorCloseDialog = $('#close-river-editor');
                        riverEditorCloseDialog.on('hidden.bs.modal', function (e) {
                            t.closeCallback = function () {

                            };
                        });
                        riverEditorCloseDialog.modal();
                    }
                },
            }
        }
    }
</script>
