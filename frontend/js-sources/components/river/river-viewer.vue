<template>
    <div>
        <breadcrumbs :country="country" :region="region"/>
        <h1>{{ river.title }}</h1>
        <div style="float:right;">
            <div style="width: 700px;">
                <img border='0' :src="informerUrl" v-for="informerUrl in informerUrls()">
            </div>
            <div id="map" style="width:650px; height: 450px;padding-left: 30px;"></div>
        </div>
        <dl>
            <dt>Показывать на карте:</dt>
            <dd>
                <span style="padding-left:40px;" v-if="river.visible">Да</span>
                <span style="padding-left:40px;" v-else>Нет</span>&nbsp;&nbsp;&nbsp;
                <button type="button" class="btn btn-info" v-if="canEdit && pageMode != 'edit' && !river.visible"
                        v-on:click="setVisible(true); hideError();">
                    Показывать на карте
                </button>
                <button type="button" class="btn btn-info" v-if="canEdit && pageMode != 'edit' && river.visible"
                        v-on:click="setVisible(false); hideError();">
                    Скрыть на карте
                </button>
                <div v-if="canEdit" style="padding-left:40px;" class="wwmap-system-hint">Нужно, когда мы не хотим
                    выставлять наполовину размеченную и описанную реку.
                    Если добавляешь часть порогов, а остальные планируешь на потом, не делай реку видимой на карте.
                </div>
                <div v-else></div>
            </dd>
            <dt v-if="river.region.id>0">Регион:</dt>
            <dd v-if="river.region.id>0">
                <div style="padding-left:40px;">
                    <div v-if="river.region.fake">{{country.title}}</div>
                    <div v-else>{{river.region.title}}</div>
                </div>
            </dd>
            <dt>Скачать:</dt>
            <dd>
                <div style="padding-left:40px;"><a
                        id="gpx" :href="gpxUrl(false)" style="padding-right:10px;"
                        alt="Скачать GPX с точками порогов">GPX</a>&nbsp;<a
                        id="gpx_en" :href="gpxUrl(true)"
                        alt="Скачать GPX с точками порогов">GPX<sub>en</sub></a><br/><a
                        id="csv" :href="csvUrl(false)" style="padding-right:10px;"
                        alt="Скачать пороги таблицей">CSV</a>&nbsp;<a
                        id="csv_en" :href="csvUrl(true)" alt="Скачать пороги таблицей">CSV<sub>en</sub></a>
                </div>

            </dd>
            <dt>Описание:</dt>
            <dd>
                <div style="padding-left:40px;">
                    {{river.description}}
                </div>

            </dd>
            <dt>Другие варианты названия для автоматического поиска отчётов:</dt>
            <dd>
                <ul>
                    <li v-for="alias in river.aliases">{{alias}}</li>
                </ul>
            </dd>
            <dt>Отчёты:
            </dt>
            <dd>
                <div style="padding-left:40px;" class="wwmap-system-hint" v-if="canEdit">Поиск отчётов происходит
                    раз в сутки ночью. Наберитесь терпения.
                </div>
                <ul>
                    <li v-for="report in reports"><a target="_blank" :href="report.url">{{report.title}}</a></li>
                </ul>
            </dd>
        </dl>
    </div>
</template>

<style type="text/css">

</style>

<script>
    import FileUpload from 'vue-upload-component';
    import {getWwmapSessionId, hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth'
    import {getRiverFromTree, navigateToSpot, store} from '../../app-state'
    import {emptyBounds, getRiverBounds, setRiverVisible,} from '../../editor'
    import {backendApiBase} from '../../config'
    import {addMapLayers, registerMapSwitchLayersHotkeys} from '../../map-common';

    var $ = require("jquery");
    require("jquery.cookie");

    module.exports = {
        props: ['river', 'reports', 'country', 'region'],
        components: {
            FileUpload: FileUpload
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

            });
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
        data: function () {
            return {
                map: null,
                canEdit: false,

                setVisible: function (visible) {
                    setRiverVisible(this.river.id, visible).then(river => {
                        this.river = river;
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

                informerUrls: function () {
                    if (!this.river.props.vodinfo_sensors) {
                        return [];
                    }
                    return this.river.props.vodinfo_sensors.map(function (s) {
                        return "http://gis.vodinfo.ru/informer/draw/v2_" + s + "_300_200_30_ffffff_110_8_7_H_none.png";
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

                            objectManager.objects.events.add(['click'], function (e) {
                                let id = e.get('objectId');
                                navigateToSpot(id, false);
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

                gpxUrl: function (transliterate) {
                    return `${backendApiBase}/downloads/river/${this.river.id}/gpx?tr=${transliterate}`;
                },
                csvUrl: function (transliterate) {
                    return `${backendApiBase}/downloads/river/${this.river.id}/csv?tr=${transliterate}`;
                },
            }
        },
    }

</script>