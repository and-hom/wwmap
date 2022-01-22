<template>
    <div>
        <btn-bar ref="btnBar" logObjectType="RIVER" :logObjectId="river.id">
            <button type="button" class="btn btn-success"
                    v-on:click="save()">Сохранить
            </button>
            <slot></slot>
        </btn-bar>
        <div class="spot-editor-panel" style="padding-top:15px;">
            <b-tabs>
                <b-tab title="Главное" active>
                    <label for="river_title" style="font-weight:bold; margin-top:5px">Название:</label>
                    <input v-model.trim="river.title" style="display:block;" id="river_title"/>
                    <dl style="margin-top:10px;">
                        <dt v-if="river.region.id>0">Регион:</dt>
                        <dd v-if="river.region.id>0">
                            <select v-model="river.region.id">
                                <option v-for="region in regions" v-bind:value="region.id">{{region.title}}</option>
                            </select>
                        </dd>
                        <dt>Описание:</dt>
                        <dd>
                          <div class="wwmap-system-hint" style="color: red">Избегайте излишнего форматирования: этот текст показывается в карточке реки на карте.</div>
                          <div style="margin-left: 40px; margin-top: 10px">
                              <editor ref="descEditor"
                                  initialEditType="wysiwyg"
                                  :initialValue="river.description"
                                  :options="editorOptions"
                                  v-on:change="onDescriptionChanged();"/>
                          </div>
                        </dd>
                        <dt>Другие варианты названия для автоматического поиска отчётов:</dt>
                        <dd>
                            <div class="wwmap-system-hint" style="margin-bottom: 7px;">Каждое альтернативное название на
                                новой строке
                            </div>
                            <textarea v-bind:text-content="river.aliases"
                                      v-on:input="river.aliases = parseAliases($event.target.value)"
                                      rows="10" cols="120"
                                      style="resize: none; margin-left:40px;">{{ river.aliases.join('\n') }}</textarea>
                        </dd>
                    </dl>
                </b-tab>
                <b-tab title="Видео" :disabled="river.id<=0">
                  <video-add :spot="spot" v-model="_videos" type="video" :auth="true"></video-add>
                </b-tab>
                <b-tab title="Стоянки" :disabled="river.id<=0">
                  <linked-entity-in-place-editor
                      v-model="river.camps"
                      :multiselect="true"
                      :bind-id="true"
                      :base-url="campApiBase"
                      :auth-for-list="false"
                      :show-selected="false"
                      :map="true"
                      :map-default-location="getCenter"
                      :allow-empty-title="true">
                    <template v-slot:form="slotProps">
                      <camp-form v-model="slotProps.entity"/>
                    </template>
                  </linked-entity-in-place-editor>
                </b-tab>
                <b-tab title="Заброски" :disabled="river.id<=0">
                  <linked-entity-in-place-editor
                      v-model="river.transfers"
                      :multiselect="true"
                      :bind-id="true"
                      :base-url="transferApiBase"
                      :auth-for-list="false"
                      :show-selected="true"
                      :map="false">
                    <template v-slot:form="slotProps">
                      <transfer-form v-model="slotProps.entity"/>
                    </template>
                    <template v-slot:grid="slotProps">
                      <table class="table">
                        <thead>
                        <tr>
                          <th width="200px">Название</th>
                          <th width="250px">Откуда</th>
                          <th>Контакты / Описание</th>
                        </tr>
                        </thead>
                        <tbody>
                        <tr v-for="transfer in slotProps.entities">
                          <td width="200px">{{transfer.title}}</td>
                          <td width="250px"><ul style="display: inline-block; margin-bottom: 0;" class="wwmap-tags">
                            <li v-for="station in transfer.stations" style="margin-bottom: 0;" class="wwmap-tag">{{station}}</li>
                          </ul></td>
                          <td>{{transfer.description}}</td>
                        </tr>
                        </tbody>
                      </table>
                    </template>
                  </linked-entity-in-place-editor>
                </b-tab>
                <b-tab title="Системные параметры">
                    <span class="wwmap-system-hint" style="padding-top: 10px;">Тут собраны настройки разных системных вещей для этой реки</span>
                    <props :p="river.props">
                        <template slot="before">
                            <div class="row">
                                <div class="col-3">
                                    <strong>Подложка общей карты реки при экспорте</strong>
                                </div>
                                <div class="col-9">
                                    <select v-model="river.props.river_export_map_type">
                                        <option :value="null">По-умолчанию</option>
                                        <option value="google#satellite">Спутник Google</option>
                                        <option value="yandex#satellite">Спутник Яндекс</option>
                                        <option value="osm#standard">OSM</option>
                                        <option value="ggc#standard">Топографичсекая карта</option>
                                    </select>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-3">
                                    <strong>Гидропост <a
                                            href="http://gis.vodinfo.ru/informer/">gis.vodinfo.ru/informer</a></strong>
                                </div>
                                <div class="col-9">
                                  <linked-entity-in-place-editor
                                      v-model="river.props.vodinfo_sensors"
                                      :multiselect="true"
                                      :bind-id="true"
                                      :data="sensors"
                                      :show-selected="false"/>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-3">
                                    <strong>Точка отслеживания погоды</strong>
                                    <div class="wwmap-system-hint">Сейчас используется только для формирования ссылки на
                                        прогноз погоды.
                                    </div>
                                </div>
                                <div class="col-9">
                                  <linked-entity-in-place-editor
                                      v-model="river.props.meteo_point"
                                      :multiselect="false"
                                      :bind-id="true"
                                      :base-url="meteoPointApiBase"
                                      :auth-for-list="true"
                                      :show-selected="false"
                                      :map-default-location="getCenter"/>
                                </div>
                            </div>
                            <hr/>
                            <h2>Для отдельных порогов:</h2>
                        </template>
                    </props>
                </b-tab>
            </b-tabs>
        </div>
    </div>
</template>
<style type="text/css">

</style>

<script>
    import FileUpload from 'vue-upload-component';
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth'
    import {sensors, sensorsById} from '../../sensors'
    import {store} from '../../app-state'
    import {
        getAllRegions,
        getAllTransfers,
        nvlReturningId,
        saveRiver,
        setActiveEntityUrlHash,
    } from '../../editor'
    import {backendApiBase} from '../../config'
    import {markdownEditorConfig} from '../../toast-editor-config'
    import {getWwmapSessionId} from "wwmap-js-commons/auth";
    import {Editor} from '@toast-ui/vue-editor';
    import {createMapParamsStorage} from "wwmap-js-commons/map-settings";

    var $ = require("jquery");
    require("jquery.cookie");

    module.exports = {
        props: ['river', 'reports', 'transfers', 'country', 'region', 'bounds'],
        components: {
            FileUpload: FileUpload,
            editor: Editor,
        },
        computed: {
            uploadPath: function () {
                return backendApiBase + "/river/" + this.river.id + "/gpx"
            },
            headers: function () {
                return {
                    Authorization: getWwmapSessionId()
                }
            },
            pageMode: {
                get: function () {
                    return store.state.rivereditorstate.pageMode
                },

                set: function (newVal) {
                    store.commit("setRiverEditorPageMode", newVal);
                }
            },
        },
        created: function () {
            getAllRegions().then(regions => {
                this.regions = regions.map(function (x) {
                    if (x.fake) {
                        return {
                            id: x.id,
                            title: x.country.title
                        }
                    }
                    return {
                        id: x.id,
                        title: x.country.title + " - " + x.title
                    }
                });
            });

            getAllTransfers().then(transfers => this.allTransfers = transfers);
            this.selectedSensors = this.river.props.vodinfo_sensors;
        },
        mounted: function () {
            let t = this;
            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
        },
        data: function () {
            return {
                prevRegionId: nvlReturningId(this.river.region.id),
                prevCountryId: this.river.region.country_id,
                map: null,
                mapParamsStorage: createMapParamsStorage(),
                operationInProgress: false,

                canEdit: false,
                meteoPointApiBase: backendApiBase + '/meteo-point',
                transferApiBase: backendApiBase + '/transfer',
                campApiBase: backendApiBase + '/camp',

                editorOptions: markdownEditorConfig,
                save: function () {
                    if (!this.river.title || !this.river.title.replace(/\s/g, '').length) {
                        this.showError("Нельзя сохранять реку без названия");
                        return
                    }

                    this.$refs.btnBar.disable()
                    saveRiver(this.river).then(updated => {
                        this.selectedSensors = this.river.props.vodinfo_sensors;
                        this.pageMode = 'view';
                        this.hideError();
                        var new_region_id = nvlReturningId(updated.region);
                        setActiveEntityUrlHash(updated.region.country_id, new_region_id, updated.id);
                        store.commit('setTreeSelection', {
                            countryId: updated.region.country_id,
                            regionId: new_region_id,
                            riverId: updated.id,
                            spotId: null
                        });
                        store.dispatch('reloadCountrySubentities', updated.region.country_id);

                        if (new_region_id > 0 && !updated.region.fake) {
                            store.dispatch('reloadRegionSubentities', {
                                countryId: updated.region.country_id,
                                regionId: new_region_id
                            });
                        }
                        if (this.prevCountryId > 0 && this.prevRegionId > 0 && !this.prevRegionFake && this.prevRegionId != new_region_id) {
                            store.dispatch('reloadRegionSubentities', {
                                countryId: this.prevCountryId,
                                regionId: this.prevRegionId
                            });
                        } else if (this.prevCountryId > 0 && this.prevCountryId != updated.region.country_id) {
                            store.dispatch('reloadCountrySubentities', this.prevCountryId);
                        }

                        this.pageMode = 'view';

                    }, _ => {
                        this.showError("Не удалось сохранить реку. Возможно, недостаточно прав");
                    }).finally(() => {
                        if (this.$refs.btnBar) {
                            this.$refs.btnBar.enable()
                        }
                    });
                },

                showError: function (errMsg) {
                    store.commit("setErrMsg", errMsg);
                },
                hideError: function () {
                    store.commit("setErrMsg", null);
                },
                // end of editor

                files: [],


                regions: [],
                parseAliases: function (strVal) {
                    return strVal.split('\n').map(function (x) {
                        return x.trim()
                    }).filter(function (x) {
                        return x.length > 0
                    })
                },
                sensors: sensors,
                selectedSensors: [],
                sensorsById: sensorsById,
                activeSensor: {id: null, title: null},

                gpxJustUploaded: false,
                gpxUrl: function (transliterate) {
                    return `${backendApiBase}/downloads/river/${this.river.id}/gpx?tr=${transliterate}`;
                },
                csvUrl: function (transliterate) {
                    return `${backendApiBase}/downloads/river/${this.river.id}/csv?tr=${transliterate}`;
                },

                allTransfers: [],

              getCenter: function (bounds) {
                return bounds
                    ? [
                      (bounds[0][0] + bounds[1][0])/2,
                      (bounds[0][1] + bounds[1][1])/2,
                    ]
                    : createMapParamsStorage().getLastPositionZoomTypeToggles().position;
              },
            }
        },
        methods: {
            onDescriptionChanged: function (){
              this.river.description = this.$refs.descEditor.invoke('getMarkdown');
            },
            onSelectSensor: function (x) {
                if (!x || !x.id) {
                    return
                }
                if (!this.river.props.vodinfo_sensors) {
                    this.river.props.vodinfo_sensors = []
                }
                if (!this.river.props.vodinfo_sensors.includes(x.id)) {
                    this.river.props.vodinfo_sensors.push(x.id);
                }
                if (this.activeSensor.id != null) {
                    this.activeSensor = {id: null, title: null};
                }
                this.selectedSensors = this.river.props.vodinfo_sensors;
            },
            removeSensor: function (id) {
                if (this.river.props.vodinfo_sensors) {
                    this.river.props.vodinfo_sensors = this.river.props.vodinfo_sensors.filter(function (s) {
                        return s != id;
                    }).slice();
                    this.selectedSensors = this.river.props.vodinfo_sensors;
                }
            },
        }
    }

</script>