<template>
    <div class="spot-editor-panel" style="padding-top:15px;">
        <b-tabs>
            <b-tab title="Главное" active>
                <input v-model.trim="river.title" style="display:block"/>
                <dl>
                    <dt>Показывать на карте:</dt>
                    <dd>
                        <span style="padding-left:40px;" v-if="river.visible">Да</span>
                        <span style="padding-left:40px;" v-else>Нет</span>&nbsp;&nbsp;&nbsp;
                        <button type="button" class="btn btn-info"
                                v-if="canEdit && !river.visible"
                                v-on:click="setVisible(true); hideError();">
                            Показывать на карте
                        </button>
                        <button type="button" class="btn btn-info" v-if="canEdit && river.visible"
                                v-on:click="setVisible(false); hideError();">
                            Скрыть на карте
                        </button>
                        <div style="padding-left:40px;" class="wwmap-system-hint">Нужно, когда мы не хотим выставлять
                            наполовину размеченную и описанную реку.
                            Если добавляешь часть порогов, а остальные планируешь на потом, не делай реку видимой на
                            карте.
                        </div>
                    </dd>
                    <dt v-if="river.region.id>0">Регион:</dt>
                    <dd v-if="river.region.id>0">
                        <select v-model="river.region.id">
                            <option v-for="region in regions" v-bind:value="region.id">{{region.title}}</option>
                        </select>
                    </dd>
                    <dt>Описание:</dt>
                    <dd>
                                    <textarea v-bind:text-content="river.aliases"
                                              rows="10" cols="120"
                                              style="resize: none; margin-left:40px;"
                                              v-model="river.description"></textarea>
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
                                <strong>Гидропост <a href="http://gis.vodinfo.ru/informer/">gis.vodinfo.ru/informer</a></strong>
                            </div>
                            <div class="col-9">
                                <ul id="sensors">
                                    <li v-for="sensor in selectedSensors">
                                        {{ sensor }} - {{ sensorsById[sensor] }}
                                        <button v-on:click.stop="removeSensor(sensor)">[x]</button>
                                    </li>
                                </ul>
                                <v-select v-model="activeSensor" label="title" :options="sensors"
                                          @input="onSelectSensor">
                                    <template slot="no-options">
                                        Начните печатать название гидропоста
                                    </template>
                                    <template slot="option" slot-scope="option">
                                        {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}
                                    </template>
                                    <template slot="selected-option" slot-scope="option">
                                        {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}
                                    </template>
                                </v-select>
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
                                <div v-if="meteoPointSelectMode">
                                    <div class="wwmap-system-hint">Выберите из списка или
                                        <button v-on:click.stop="addMeteoPoint">создайте</button>
                                    </div>
                                    <v-select v-model="meteoPoint" label="title" :options="meteoPoints"
                                              @input="onSelectMeteoPoint">
                                        <template slot="no-options">
                                            Начните печатать название точки
                                        </template>
                                        <template slot="option" slot-scope="option">
                                            {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}
                                        </template>
                                        <template slot="selected-option" slot-scope="option">
                                            {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}
                                        </template>
                                    </v-select>
                                </div>
                                <div v-else>
                                    <div style="padding-top:15px;">
                                        <ya-map-location v-bind:spot="meteoPoint" width="100%" height="600px"
                                                         :editable="true" :ya-search="true"/>
                                    </div>
                                    <label style="padding-right: 10px;"
                                           for="meteo_point_title_input"><strong>Название:</strong></label><input
                                        id="meteo_point_title_input" type="text" v-model="meteoPoint.title"
                                        style="margin-top: 10px; width: 80%;"/>
                                    <div class="btn-toolbar" style="padding-top:15px;">
                                        <div class="btn-group mr-2" role="group">
                                            <button type="button" class="btn btn-success"
                                                    v-on:click.stop="onAddMeteoPoint" :disabled="!meteoPoint.title">
                                                Добавить
                                            </button>
                                            <button type="button" class="btn btn-cancel"
                                                    v-on:click.stop="onCancelAddMeteoPoint">Отмена
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <hr/>
                        <h2>Для отдельных порогов:</h2>
                    </template>
                </props>
            </b-tab>
        </b-tabs>
    </div>
</template>
<style type="text/css">

</style>

<script>
    import FileUpload from 'vue-upload-component';
    import {getWwmapSessionId, hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth'
    import {sensors, sensorsById} from '../../sensors'
    import {getRiverFromTree, store} from '../../main'
    import {
        addMeteoPoint,
        emptyBounds,
        getAllRegions,
        getMeteoPoints,
        getRiverBounds,
        nvlReturningId,
        saveRiver,
        setActiveEntityUrlHash,
        setRiverVisible,
    } from '../../editor'
    import {backendApiBase} from '../../config'
    import {addMapLayers, registerMapSwitchLayersHotkeys} from '../../map-common';

    var $ = require("jquery");
    require("jquery.cookie");

    module.exports = {
        props: ['river', 'reports', 'country', 'region'],
        components: {
            FileUpload: FileUpload
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
            this.selectedSensors = this.river.props.vodinfo_sensors;
        },
        mounted: function () {
            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);

            getRiverBounds(this.river.id).then(bounds => {
                this.bounds = bounds;
                this.center = [(bounds[0][0] + bounds[1][0]) / 2, (bounds[0][1] + bounds[1][1]) / 2];

                let hideMap = emptyBounds(this.bounds);
                if (this.map && hideMap) {
                    this.map.destroy();
                    this.map = null;
                } else if (this.map) {
                    this.objectManager.setUrlTemplate(backendApiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z&river=' + this.river.id);
                    this.map.setBounds(this.bounds);
                } else if (!this.map && !hideMap) {
                    this.showMap();
                }

                getMeteoPoints().then(points => this.meteoPoints = points);
                this.meteoPoint = this.getMeteoPointById(this.river.props.meteo_point);
            });
        },
        data: function () {
            return {
                map: null,

                canEdit: false,
                save: function () {
                    if (!this.river.title || !this.river.title.replace(/\s/g, '').length) {
                        this.showError("Нельзя сохранять реку без названия");
                        return
                    }
                    saveRiver(this.river).then(updated => {
                        this.meteoPoint = this.getMeteoPointById(this.river.props.meteo_point);
                        this.selectedSensors = this.river.props.vodinfo_sensors;
                        this.pageMode = 'view';
                        this.hideError();
                        var new_region_id = nvlReturningId(updated.region);
                        setActiveEntityUrlHash(updated.region.country_id, new_region_id, updated.id);
                        store.commit('setActiveEntityState', {
                            countryId: updated.region.country_id,
                            regionId: new_region_id,
                            riverId: updated.id,
                            spotId: null
                        });
                        store.commit('showCountrySubentities', updated.region.country_id);

                        if (new_region_id > 0 && !updated.region.fake) {
                            store.commit('showRegionSubentities', {
                                countryId: updated.region.country_id,
                                regionId: new_region_id
                            });
                        }
                        if (this.prevCountryId > 0 && this.prevRegionId > 0 && !this.prevRegionFake && this.prevRegionId != new_region_id) {
                            store.commit('showRegionSubentities', {
                                countryId: this.prevCountryId,
                                regionId: this.prevRegionId
                            });
                        } else if (this.prevCountryId > 0 && this.prevCountryId != updated.region.country_id) {
                            store.commit('showCountrySubentities', this.prevCountryId);
                        }

                        this.reloadMap();

                        this.pageMode = 'view';

                    }, _ => {
                        this.showError("Не удалось сохранить реку. Возможно, недостаточно прав");
                    });
                },
                setVisible: function (visible) {
                    setRiverVisible(this.river.id, visible).then(river => {
                        this.river = river;
                        this.meteoPoint = this.getMeteoPointById(this.river.props.meteo_point);
                        this.selectedSensors = this.river.props.vodinfo_sensors;
                        // set "visible" property to global storage to set icon in the left-side tree using reactivity of vue.js
                        var regionId = this.river.region.id;
                        if (this.river.region.fake) {
                            regionId = -1
                        }
                        getRiverFromTree(this.river.region.country_id, regionId, this.river.id).visible = this.river.visible;
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
                add_spot: function () {
                    store.commit("setRiverEditorVisible", false);
                    store.commit("setSpotEditorVisible", false);

                    store.commit("setSpotEditorState", {
                        visible: true,
                        editMode: true,
                        spot: {
                            id: 0,
                            river: this.river,
                            order_index: "0",
                            automatic_ordering: true,
                            point: this.center,
                            aliases: [],
                            props: {},
                        },
                        country: this.country,
                        region: this.region,
                    });
                },

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


                meteoPoints: [],
                meteoPoint: null,
                meteoPointSelectMode: true,
                getMeteoPointById: function (id) {
                    if (id) {
                        for (let i = 0; i < this.meteoPoints.length; i++) {
                            if (this.meteoPoints[i].id == id) {
                                return this.meteoPoints[i]
                            }
                        }
                    }
                    return {id: null, title: null, point: this.center}
                },
                addMeteoPoint: function () {
                    this.meteoPointSelectMode = false;
                },
                onCancelAddMeteoPoint: function () {
                    this.meteoPointSelectMode = true;
                },
                onAddMeteoPoint: function () {
                    this.onCancelAddMeteoPoint();
                    addMeteoPoint(this.meteoPoint).then(p => {
                        this.meteoPoint = p;
                        this.river.props.meteo_point = p.id;
                        getMeteoPoints().then(meteoPoints => this.meteoPoints = meteoPoints)
                    });
                },
                getDefaultMap: function () {
                    let defaultMap = $.cookie("default_editor_map");
                    if (defaultMap && ymaps.mapType.storage.get(defaultMap)) {
                        return defaultMap
                    }
                    return "osm#standard"
                },
                showMap: function () {
                    if (emptyBounds(this.bounds)) {
                        return;
                    }
                    let t = this;
                    ymaps.ready(function () {
                        ymaps.modules.require(['overlay.BiPlacemark'], function (BiPlacemarkOverlay) {
                            if (ymaps.overlay.storage.get("BiPlacemrakOverlay") == null) {
                                ymaps.overlay.storage.add("BiPlacemrakOverlay", BiPlacemarkOverlay);
                            }
                            addMapLayers();

                            let mapType = t.getDefaultMap();
                            let map = new ymaps.Map("map", {
                                bounds: t.bounds,
                                type: mapType,
                                controls: ["zoomControl"]
                            });
                            map.controls.add(
                                new ymaps.control.TypeSelector([
                                        'osm#standard',
                                        'ggc#standard',
                                        'topomapper#genshtab',
                                        'marshruty.ru#genshtab',
                                        'yandex#satellite',
                                        'google#satellite',
                                        'bing#satellite',
                                    ]
                                )
                            );
                            registerMapSwitchLayersHotkeys(map);
                            var objectManager = new ymaps.RemoteObjectManager(backendApiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z&river=' + t.river.id, {
                                clusterHasBalloon: false,
                                geoObjectOpenBalloonOnClick: false,
                                geoObjectStrokeWidth: 3,
                                splitRequests: true
                            });
                            map.geoObjects.add(objectManager);
                            t.map = map;
                            t.objectManager = objectManager;
                        });
                    });
                },
                reloadMap: function () {
                    if (this.map) {
                        this.map.destroy();
                    }
                    this.showMap();
                },

                gpxJustUploaded: false,
                gpxUrl: function (transliterate) {
                    return `${backendApiBase}/downloads/river/${this.river.id}/gpx?tr=${transliterate}`;
                },
                csvUrl: function (transliterate) {
                    return `${backendApiBase}/downloads/river/${this.river.id}/csv?tr=${transliterate}`;
                },
            }
        },
        methods: {
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
            onSelectMeteoPoint: function (x) {
                this.river.props.meteo_point = x.id
            }
        }
    }

</script>