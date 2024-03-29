import {WWMapSearchProvider} from "./searchProvider";
import {createLegend} from "./legend";
import {highlight_river} from "./main";
import {
    CACHED_TILES_TEMPLATE,
    createCountryUrlPart,
    createUrlPart,
    defaultPosition,
    isMobileBrowser,
    LAST_MAP_TYPE_KEY,
    LAST_POS_KEY,
    LAST_TOGGLES_KEY,
    LAST_ZOOM_KEY,
} from './util';
import {bingSatTiles} from './map-urls/bing'
import {googleSatTiles} from './map-urls/google'
import {apiBase, frontendBase} from "./config";
import {createMeasurementToolControl} from "./router/control";
import {WWMapMeasurementTool} from "./router/measurement";
import {FeatureToggles} from "./feature-toggles";
import {getWwmapSessionId} from "wwmap-js-commons/auth";
import {SHOW_CAMPS_MIN_ZOOM, SHOW_SLOPE_MIN_ZOOM} from "wwmap-js-commons/constants";
import {
    DEFAULT_MAP_TYPE,
    DEFAULT_POSITION,
    DEFAULT_TOGGLES,
    DEFAULT_ZOOM,
    DefaultMapParamsStorage,
    EditorHashMapParamsStorage,
    JsonDataMapper,
    KeyValueMapParamsStorage,
    LocalStorageDataStorage,
    MapParams,
    MapParamsStorage
} from "wwmap-js-commons/map-settings";

export function WWMap(divId, bubbleTemplate, riverList, tutorialPopup, catalogLinkType) {
    this.divId = divId;
    this.bubbleTemplate = bubbleTemplate;

    this.riverList = riverList;

    this.tutorialPopup = tutorialPopup;
    this.catalogLinkType = catalogLinkType;

    this.catFilter = 1;

    this.countryId = true;
    this.experimentalFeatures = false;

    this.onBoundsChange = null;

    addCachedLayer('osm#standard', 'OSM (O)', 'OpenStreetMap contributors, CC-BY-SA', 'osm');
    addLayer('google#satellite', 'Спутник Google (G)', 'Изображения © DigitalGlobe,CNES / Airbus, 2018,Картографические данные © Google, 2018', googleSatTiles);
    addLayer('bing#satellite', 'Спутник Bing (B)', 'Изображения © Майкрософт (Microsoft), 2019', bingSatTiles);
    addCachedLayer('ggc#standard', 'Топографическая карта (T)', '', 'ggc', 0, 15);
    addCachedLayer('topomapper#genshtab', 'TopoMapper', 'TopoMapper.com', 'topo-mapper', 0, 13);
    addCachedLayer('marshruty.ru#genshtab', 'Маршруты.ру', 'marshruty.ru', 'marshruty-ru', 8, 13);

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

        let togglesPart = createUrlPart('toggles', this.featureToggles.serialize());
        let authPart = this.featureToggles.getNeedsAuth()
            ? createUrlPart('session_id', getWwmapSessionId())
            : '';

        let url = `${apiBase}/visible-rivers-lite?bbox=${bounds.join(',')}&max_cat=${this.catFilter}${togglesPart}${authPart}`;
        $.get(url, function (data) {
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
    let helpButton = new ymaps.control.Button({
        data: {
            image: `${frontendBase}/img/help.png`,
            title: 'Справка',
        },
        options: {
            selectOnClick: false
        }
    });
    var t = this;
    helpButton.events.add('press', function (e) {
        t.tutorialPopup.show()
    });
    return helpButton
};

WWMap.prototype.initToolBtn = function (image, title, selected, onpress) {
    let btn = new ymaps.control.Button({
        data: {
            image: image,
            title: title,
        },
        options: {
            selectOnClick: false,
        }
    });
    let t = this;
    btn.state.set('selected', selected);
    btn.events.add('press', e => {
        onpress();

        t.mapParamStorage.setLastPositionZoomTypeToggles(
            t.yMap.getCenter(),
            t.yMap.getZoom(),
            t.yMap.getType(),
            t.featureToggles.serialize(),
            t.countryId
        );
    });
    this.yMap.controls.add(btn, {
        float: 'right'
    });
    return btn;
};

WWMap.prototype.reloadMapObjects = function () {
    this.objectManager.setUrlTemplate(this.createObjectsUrlTemplate())
    this.objectManager.reloadData();

    this.loadRivers(this.yMap.getBounds());
}

WWMap.prototype.createObjectsUrlTemplate = function () {
    let togglesPart = createUrlPart('toggles', this.featureToggles.serialize());
    let authPart = this.featureToggles.getNeedsAuth()
        ? createUrlPart('session_id', getWwmapSessionId())
        : '';
    let countryPart = createCountryUrlPart(this.countryId)
    return `${apiBase}/ymaps-tile-ww?bbox=%b&zoom=%z&link_type=${this.catalogLinkType}${countryPart}${togglesPart}${authPart}`;
};

WWMap.prototype.setShowUnpublished = function (showUnpublished) {
    this.featureToggles.setShowUnpublished(showUnpublished);
    this.reloadMapObjects();
};

WWMap.prototype.setShowCamps = function (showCamps) {
    this.featureToggles.setShowCamps(showCamps);
    this.reloadMapObjects();
};

WWMap.prototype.setShowSlope = function (showSlope) {
    this.featureToggles.setShowSlope(showSlope);
    this.reloadMapObjects();
};

WWMap.prototype.setShowAltitudeCoverage = function (showSlope) {
    this.featureToggles.setShowAltitudeCoverage(showSlope);
    this.reloadMapObjects();
};

WWMap.prototype.init = function (opts) {
    this.countryId = opts ? opts.countryId : null;
    let defaultPositionValue = opts ? opts.defaultCenter : defaultPosition();
    let useHash = opts ? opts.useHash : true;
    let options = opts ? opts.mapOptions : {};
    let showHideButtonsOnMap = opts ? opts.showHideButtonsOnMap : true;
    this.experimentalFeatures = opts ? opts.experimentalFeatures : false;
    let canShowUnpublished = opts ? opts.canShowUnpublished : false;

    this.mapParamStorage = useHash
        ? new MapParamsStorage(
            new EditorHashMapParamsStorage(0),
            new KeyValueMapParamsStorage(LAST_POS_KEY, LAST_ZOOM_KEY, LAST_MAP_TYPE_KEY, LAST_TOGGLES_KEY, new JsonDataMapper(), new LocalStorageDataStorage()),
            new DefaultMapParamsStorage(new MapParams(DEFAULT_POSITION, DEFAULT_ZOOM, DEFAULT_MAP_TYPE, DEFAULT_TOGGLES))
        )
        : new MapParamsStorage(
            new KeyValueMapParamsStorage(LAST_POS_KEY, LAST_ZOOM_KEY, LAST_MAP_TYPE_KEY, LAST_TOGGLES_KEY, new JsonDataMapper(), new LocalStorageDataStorage()),
            new DefaultMapParamsStorage(new MapParams(DEFAULT_POSITION, DEFAULT_ZOOM, DEFAULT_MAP_TYPE, DEFAULT_TOGGLES))
        );

    let positionZoomTypeToggles = this.mapParamStorage.getLastPositionZoomTypeToggles(this.countryId);

    let yMap;
    try {
        yMap = new ymaps.Map(this.divId, {
            center: positionZoomTypeToggles.position,
            zoom: positionZoomTypeToggles.zoom,
            controls: ["zoomControl", "fullscreenControl"],
            type: positionZoomTypeToggles.type
        }, options);
    } catch (err) {
        this.mapParamStorage.setLastPositionZoomTypeToggles(
            defaultPositionValue,
            DEFAULT_ZOOM,
            DEFAULT_MAP_TYPE,
            DEFAULT_TOGGLES,
            this.countryId
        )
        throw err
    }
    this.yMap = yMap;
    let t = this;

    this.yMap.controls.add(
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
    $(document).keyup(function (e) {
        if (!e || !e.target || !e.target.tagName ||
            e.target.tagName.toUpperCase() == 'INPUT' ||
            e.target.tagName.toUpperCase() == 'TEXTAREA') {
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

    let showCampsButton;
    let showUnpublishedButton;
    let showSlopeButton;
    let showAltitudeCoverageButton;
    if (showHideButtonsOnMap) {
        showCampsButton = this.initToolBtn(
            `${frontendBase}/img/camp.svg`,
            'Показывать стоянки',
            false,
            () => this.setShowCamps(!this.featureToggles.getShowCamps())
        );
        if (canShowUnpublished) {
            showUnpublishedButton = this.initToolBtn(
                `${frontendBase}/img/invisible.png`,
                'Показывать неопубликованное',
                false,
                () => this.setShowUnpublished(!this.featureToggles.getShowUnpublished())
            );
        }
        if (this.experimentalFeatures) {
            showSlopeButton = this.initToolBtn(
                `${frontendBase}/img/slope.png`,
                'Показывать уклон рек',
                false,
                () => this.setShowSlope(!this.featureToggles.getShowSlope())
            );
            showAltitudeCoverageButton = this.initToolBtn(
                `${frontendBase}/img/alt-coverage.png`,
                'Покрытие данных о высоте',
                false,
                () => this.setShowAltitudeCoverage(!this.featureToggles.getShowAltitudeCoverage())
            );
        }
    }
    this.featureToggles = new FeatureToggles(
        [true, false, true, false,],
        [true, canShowUnpublished, this.experimentalFeatures, this.experimentalFeatures,],
        [showCampsButton, showUnpublishedButton, showSlopeButton, showAltitudeCoverageButton],
        [SHOW_CAMPS_MIN_ZOOM, 0, SHOW_SLOPE_MIN_ZOOM, 0],
        [false, true, true, true],
        positionZoomTypeToggles.zoom
    )
    this.featureToggles.parse(positionZoomTypeToggles.toggles);

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
        let center = t.yMap.getCenter();
        let zoom = t.yMap.getZoom();
        let needsReloadGeoObjects = t.featureToggles.setZoom(zoom);

        t.mapParamStorage.setLastPositionZoomTypeToggles(
            center,
            zoom,
            t.yMap.getType(),
            t.featureToggles.serialize(),
            t.countryId
        );
        t.loadRivers(e.get("newBounds"))
        if (t.onBoundsChange) {
            t.onBoundsChange(center, zoom)
        }

        if (needsReloadGeoObjects) {
            t.reloadMapObjects();
        }
    });

    this.yMap.events.add('typechange', function (e) {
        t.mapParamStorage.setLastPositionZoomTypeToggles(
            t.yMap.getCenter(),
            t.yMap.getZoom(),
            t.yMap.getType(),
            t.featureToggles.serialize(),
            t.countryId
        );
    });

    let objectManager = new ymaps.RemoteObjectManager(this.createObjectsUrlTemplate(), {
        clusterHasBalloon: false,
        geoObjectOpenBalloonOnClick: false,
        geoObjectBalloonContentLayout: ymaps.templateLayoutFactory.createClass(this.bubbleTemplate),
        geoObjectStrokeWidth: 3,
        splitRequests: false
    });

    objectManager.setFilter(function (obj) {
        if (t.catFilter <= 1 || obj.properties.object_type != 'spot') {
            return true;
        }

        if (t.catFilter) {
            var objCategory = obj.properties.category
                ? parseInt(obj.properties.category[0])
                : 0;
            var riverCategory = obj.properties.river_category
                ? parseInt(obj.properties.river_category)
                : objCategory;

            if (objCategory <= 0) {
                objCategory = riverCategory;
            }
            return objCategory >= t.catFilter;
        }

        return true
    });

    this.yMap.geoObjects.add(objectManager);
    this.objectManager = objectManager;

    objectManager.objects.events.add(['click'], function (e) {
        if (!t.measurementTool || !t.measurementTool.canEditPath()) {
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
    if (!this.isMobile) {
        this.yMap.controls.add(createMeasurementToolControl(this.measurementTool), {});
    }

    this.wwMapSearchProvider = new WWMapSearchProvider(
        e => {
            if (t.measurementTool && t.measurementTool.canEditPath()) {
                t.measurementTool.onMouseMoved(e.get('position'), e.get('coords'));
            }
        },
        e => {
            if (t.measurementTool && t.measurementTool.canEditPath()) {
                t.measurementTool.multiPath.pushEmptySegment();
            }
        },
        this.countryId,
        this.featureToggles
    );

    let searchControl = new ymaps.control.SearchControl({
        options: {
            provider: this.wwMapSearchProvider,
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

WWMap.prototype.getZoom = function () {
    return this.yMap.getZoom();
};

WWMap.prototype.setOnBoundsChange = function (onBoundsChange) {
    this.onBoundsChange = onBoundsChange;
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



