import {WWMapSearchProvider} from "./searchProvider";
import {createLegend} from "./legend";
import {show_map_at_and_highlight_river, highlight_river} from "./main";
import {CACHED_TILES_TEMPLATE, GOOGLE_SAT_TILES, bingSatTiles, getLastPositionAndZoom, setLastPositionZoomType, getWwmapUserInfoForMapControls, isMobileBrowser} from './util';
import {apiBase} from "./config";
import {createMeasurementToolControl} from "./router/control";
import {WWMapMeasurementTool} from "./router/measurement";

export function WWMap(divId, bubbleTemplate, riverList, tutorialPopup, catalogLinkType) {
    this.divId = divId;
    this.bubbleTemplate = bubbleTemplate;

    this.riverList = riverList;

    this.tutorialPopup = tutorialPopup;
    this.catalogLinkType = catalogLinkType;

    this.catFilter = 1;

    addCachedLayer('osm#standard', 'OSM (O)', 'OpenStreetMap contributors, CC-BY-SA', 'osm');
    addLayer('google#satellite', 'Спутник Google (G)', 'Изображения © DigitalGlobe,CNES / Airbus, 2018,Картографические данные © Google, 2018', GOOGLE_SAT_TILES);
    addLayer('bing#satellite', 'Спутник Bing (B)', 'Изображения © Майкрософт (Microsoft), 2019', bingSatTiles);
    addCachedLayer('ggc#standard', 'Топографическая карта (T)', '', 'ggc', 0, 15);
    //      addLayer('marshruty.ru#genshtab', 'Маршруты.ру', 'marshruty.ru', MARSHRUTY_RU_TILES, 8)

    // workaround to change Yandex Satellite map title
    try {
        ymaps.mapType.storage.get('yandex#satellite')._name = 'Спутник Yandex (Y)'
    } catch (err) {
        console.log(err)
    }

    this.selectedRiverTracks = null;

    this.isMobile = isMobileBrowser();
}

WWMap.prototype.loadRivers = function (bounds) {
    if (this.riverList) {
        var riverList = this.riverList;
        $.get(apiBase + "/visible-rivers-light?bbox=" + bounds.join(',') + "&max_cat=" + this.catFilter, function (data) {
            var dataObj = {
                "rivers": JSON.parse(data)
            };
            for (let i in dataObj.rivers) {
                if (dataObj.rivers[i].bounds) {
                    dataObj.rivers[i].bounds = JSON.stringify(dataObj.rivers[i].bounds)
                }
            }
            riverList.update(dataObj)
        });
    }
};

WWMap.prototype.createHelpBtn = function () {
    var helpButton = new ymaps.control.Button({
        data: {
            image: 'http://wwmap.ru/img/help.png'
        },
        options: {
            selectOnClick: false
        }
    });
    var t = this;
    helpButton.events.add('click', function (e) {
        t.tutorialPopup.show()
    });
    return helpButton
};

WWMap.prototype.init = function () {
    var positionAndZoom = getLastPositionAndZoom();

    var yMap;
    try {
        yMap = new ymaps.Map(this.divId, {
            center: positionAndZoom.position,
            zoom: positionAndZoom.zoom,
            controls: ["zoomControl", "fullscreenControl"],
            type: positionAndZoom.type
        });
    } catch (err) {
        setLastPositionZoomType(defaultPosition(), defaultZoom(), defaultMapType());
        throw err
    }
    this.yMap = yMap;
    let t = this;

    this.yMap.controls.add(
        new ymaps.control.TypeSelector([
                'osm#standard',
                'ggc#standard',
                'yandex#satellite',
                'google#satellite',
                'bing#satellite',
            ]
        )
    );
    $(document).keyup(function (e) {
        if (!e || !e.target || !e.target.tagName || e.target.tagName.toUpperCase() == 'INPUT') {
            return
        }
        switch (e.which) {
            case 71: // G
                t.yMap.setType('google#satellite');
                break;
            case 66: // B
                t.yMap.setType('bing#satellite');
                break;
            case 89: // Y
                t.yMap.setType('yandex#satellite');
                break;
            case 79: // O
                t.yMap.setType('osm#standard');
                break;
            case 84: // T
                t.yMap.setType('ggc#standard');
                break;
            case 27: // Escape
                t.hideSelectedRiverTracks();
                break;
        }
    });

    if (this.tutorialPopup) {
        this.yMap.controls.add(this.createHelpBtn(), {
            float: 'left'
        });
    }

    this.yMap.controls.add(createLegend(t), {
        float: 'left'
    });

    this.yMap.controls.add('rulerControl', {
        scaleLine: true
    });

    this.yMap.events.add('click', function (e) {
        t.yMap.balloon.close()
    });

    this.yMap.events.add('boundschange', function (e) {
        setLastPositionZoomType(t.yMap.getCenter(), t.yMap.getZoom(), t.yMap.getType());
        t.loadRivers(e.get("newBounds"))
    });

    this.yMap.events.add('typechange', function (e) {
        setLastPositionZoomType(t.yMap.getCenter(), t.yMap.getZoom(), t.yMap.getType())
    });

    var objectManager = new ymaps.RemoteObjectManager(apiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z&link_type=' + this.catalogLinkType, {
        clusterHasBalloon: false,
        geoObjectOpenBalloonOnClick: false,
        geoObjectBalloonContentLayout: ymaps.templateLayoutFactory.createClass(this.bubbleTemplate),
        geoObjectStrokeWidth: 3,
        splitRequests: false
    });

    objectManager.setFilter(function (obj) {
        if (obj.properties.category && t.catFilter) {
            var objCategory = parseInt(obj.properties.category[0]);
            return t.catFilter === 1 || objCategory < 0 || objCategory >= t.catFilter;
        }
        return true
    });

    this.yMap.geoObjects.add(objectManager);
    this.objectManager = objectManager;

    objectManager.objects.events.add(['click'], function (e) {
        if (!t.measurementTool || !t.measurementTool.enabled || !t.measurementTool.edit || t.measurementTool.overZoom) {
            objectManager.objects.balloon.open(e.get('objectId'));
        }
    });
    objectManager.clusters.events.add('click', function (e) {
        let objectId = e.get('objectId');
        let cluster = objectManager.clusters.getById(objectId);
        if(cluster && cluster.properties && cluster.properties.id) {
            highlight_river(cluster.properties.id)
        }
    });

    this.loadRivers(this.yMap.getBounds());

    this.measurementTool = new WWMapMeasurementTool(yMap, objectManager, apiBase);
    var info = getWwmapUserInfoForMapControls();
    if (info && info.experimental_features && !this.isMobile) {
        this.yMap.controls.add(createMeasurementToolControl(this.measurementTool), {});
    }


    let searchControl = new ymaps.control.SearchControl({
        options: {
            provider: new WWMapSearchProvider(e => {
                if (t.measurementTool.enabled && t.measurementTool.edit && !t.measurementTool.overZoom) {
                    t.measurementTool.onMouseMoved(e.get('position'), e.get('coords'));
                }
            }, e => {
                if (t.measurementTool.enabled && t.measurementTool.edit && !t.measurementTool.overZoom) {
                    t.measurementTool.multiPath.pushEmptySegment();
                }
            }),
            placeholderContent: 'Река или порог'
        }
    });
    searchControl.events.add('resultselect', function (e) {
        let index = e.get('index');

        searchControl.getResult(index).then(function (value) {
            if(!value) {
                return
            }

            let id = value.properties.get("id");
            let type = value.properties.get("type");
            let bounds = value.properties.get("boundedBy");

            if (id && bounds && type && type == 'river') {
                highlight_river(id)
            }
        }, function (err) {
            console.log('Ошибка: ' + err);
        });
    }, this);
    this.yMap.controls.add(searchControl);
};

WWMap.prototype.setBounds = function (bounds, opts) {
    this.yMap.setBounds(bounds, opts)
};

WWMap.prototype.hideSelectedRiverTracks = function () {
    if (!this.selectedRiverTracks) {
        return;
    }

    for (let i = 0; i < this.selectedRiverTracks.length; i++) {
        this.yMap.geoObjects.remove(this.selectedRiverTracks[i]);
    }

    this.selectedRiverTracks = null;
};

WWMap.prototype.setSelectedRiverTracks = function (tracks) {
    let mapObjects = [];
    for (let i = 0; i < tracks.length; i++) {
        let mapObject = new ymaps.GeoObject({
            geometry: {
                type: "LineString",
                coordinates: tracks[i].path
            },
            properties: {
                hintContent: tracks[i].Id,
                // balloonContent: ""
            }
        }, {
            strokeColor: "#0000FFAA",
            strokeWidth: 3
        });
        mapObjects.push(mapObject);
        this.yMap.geoObjects.add(mapObject);
    }

    this.selectedRiverTracks = mapObjects;
};

function addCachedLayer(key, name, copyright, mapId, lower_scale, upper_scale) {
    return addLayer(key, name, copyright, CACHED_TILES_TEMPLATE.replace('###', mapId), lower_scale, upper_scale)
}

function addLayer(key, name, copyright, tilesUrlTemplate, lower_scale, upper_scale) {
    if (typeof (lower_scale) == "undefined") {
        lower_scale = 0
    }
    if (typeof (upper_scale) == "undefined") {
        upper_scale = 18
    }
    var layer = function () {
        var layer = new ymaps.Layer(tilesUrlTemplate, {
            projection: ymaps.projection.sphericalMercator
        });
        //  Копирайты.
        layer.getCopyrights = function () {
            return ymaps.vow.resolve(copyright);
        };
        layer.getZoomRange = function () {
            return ymaps.vow.resolve([lower_scale, upper_scale]);
        };
        return layer;
    };
    ymaps.layer.storage.add(key, layer);
    ymaps.mapType.storage.add(key, new ymaps.MapType(name, [key]));
}


export const CATALOG_LINK_TYPES = [
    'none', // do not use spot link from bubble
    'from_spot',  // use link from spot properties
    'wwmap', // use link to wwmap.ru catalog
    'huskytm' // use link to huskytm.ru catalog (upload from wwmap.ru)
];



