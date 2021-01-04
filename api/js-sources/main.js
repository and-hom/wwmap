import {CATALOG_LINK_TYPES, WWMap} from './map'
import {WWMapPopup} from "./popup";
import {RiverList} from "./riverList";
import {canEdit, getWwmapUserInfoForMapControls} from "./util";
import {apiBase, apiVersion} from './config';
import {loadFragment} from "./template-data";
import {initPresets} from './placemark-preset';
import {initLayoutFilters} from "./layout-template-filter";

import './style/map.css'
import './contrib/lightbox.min.css'
import './contrib/lightbox.min'
import {doGet} from "./api";
import {regiterTemplate7Helpers} from "./template7-helpers";

require('./tube');

var wwMap;

export function show_map_at_and_highlight_river(bounds, riverId) {
    show_map_at(bounds);
    highlight_river(riverId)
}

export function highlight_river(riverId) {
    wwMap.hideSelectedRiverTracks();

    $.get(apiBase + "/river-path-segments?riverId=" + riverId, function (data) {
        wwMap.setSelectedRiverTracks(data);
    });
}

export function show_map_at(bounds) {
    wwMap.setBounds(bounds, {
        checkZoomRange: true,
        duration: 200
    });
}

var riverList;
var spotReportPopup;
var campReportPopup;
var tutorialPopup;
export var userInfoFunction = null;
export const version = apiVersion;

export function initWWMap(mapId, riversListId, options) {
    let optDefined = options && (typeof options == 'object');

    let catalogLinkType = optDefined ? options.catalogLinkType : null;
    let riversTemplateData = optDefined ? options.riversTemplateData : null;
    let countryCode = optDefined ? options.countryCode : null;
    userInfoFunction = optDefined ? options.userInfoFunction : null;

    if (catalogLinkType && CATALOG_LINK_TYPES.indexOf(catalogLinkType) <= -1) {
        throw "Unknown catalog link type. Available are: " + CATALOG_LINK_TYPES
    }

    // initialize popup windows
    spotReportPopup = new WWMapPopup('spot_report_popup', 'spot_report_popup_template', {
        submitUrl: apiBase + "/report",
        okMsg: "Запрос отправлен. Я прочитаю его по мере наличия свободного времени",
        failMsg: "Что-то пошло не так...",
        // To prevent contents lost
        closeOnEscape: false,
        closeOnMouseClickOutside: false
    });
    campReportPopup = new WWMapPopup('camp_report_popup', 'camp_report_popup_template', {
        submitUrl: apiBase + "/report",
        okMsg: "Запрос отправлен. Я прочитаю его по мере наличия свободного времени",
        failMsg: "Что-то пошло не так...",
        // To prevent contents lost
        closeOnEscape: false,
        closeOnMouseClickOutside: false
    });
    tutorialPopup = new WWMapPopup('info_popup', 'info_popup_template');

    // riverList
    if (riversListId) {
        riverList = new RiverList(riversListId, 'rivers_template', riversTemplateData)
    }

    regiterTemplate7Helpers();

    // init and show map
    return ymaps.ready()
        .then(_ => {
            initPresets();
            initLayoutFilters();
        })
        .then(_ => loadFragment('bubble_template'))
        .then(bubbleContent => {
            wwMap = new WWMap(mapId, bubbleContent, riverList, tutorialPopup, catalogLinkType);
        })
        .then(_ => ymaps.modules.require(['overlay.BiPlacemark']))
        .spread(overlay => ymaps.overlay.storage.add("BiPlacemrakOverlay", overlay))
        .then(_ => ymaps.borders.load("001"))
        .then(function (countries) {
            let restrictMapArea = null;
            let geoObject = null;
            let center = null;
            let dMax = 1.0;
            for (let i = 0; i < countries.features.length; i++) {
                if (countries.features[i].properties && countries.features[i].properties.iso3166 != countryCode) {
                    continue
                }

                try {
                    let regionCoords = countries.features[i].geometry.coordinates[0];
                    geoObject = new ymaps.Polygon([
                        [
                            [-80, 179],
                            [80, 179],
                            [80, 0],
                            [80, -179],
                            [-80, -179],
                            [-80, 0],
                            [-80, 179]
                        ],
                        regionCoords,
                    ], {}, {
                        fill: true,
                        fillOpacity: 1,
                        fillColor: '#ffffff',
                    });

                    restrictMapArea = [[90, 180], [-90, -180]];
                    for (let i = 0; i < regionCoords.length; i++) {
                        let point = regionCoords[i];
                        if (restrictMapArea[0][0] > point[0]) {
                            restrictMapArea[0][0] = point[0];
                        }
                        if (restrictMapArea[1][0] < point[0]) {
                            restrictMapArea[1][0] = point[0];
                        }
                        if (restrictMapArea[0][1] > point[1]) {
                            restrictMapArea[0][1] = point[1];
                        }
                        if (restrictMapArea[1][1] < point[1]) {
                            restrictMapArea[1][1] = point[1];
                        }
                    }

                    let dx = Math.abs(restrictMapArea[0][0] - restrictMapArea[1][0])
                    let dy = Math.abs(restrictMapArea[0][1] - restrictMapArea[1][1])
                    dMax = 0.9 * Math.max(dx, dy)

                    let centerX = (restrictMapArea[0][0] + restrictMapArea[1][0]) / 2
                    let centerY = (restrictMapArea[0][1] + restrictMapArea[1][1]) / 2

                    center = [centerX, centerY]
                    restrictMapArea = [[centerX - dMax, centerY - dMax], [centerX + dMax, centerY + dMax]];
                } catch (e) {
                    console.error(e)
                }
            }
            let mapOptions = restrictMapArea ? {
                restrictMapArea: restrictMapArea,
                minZoom: Math.max(Math.floor(Math.log(180 / dMax) / Math.log(2)), 1),
            } : {};

            return {
                mapOptions: mapOptions,
                center: center,
                geoObject: geoObject,
            }
        })
        .then(mapP => countryCode
            ? doGet(`${apiBase}/country/code/${countryCode}`)
                .then(resp => JSON.parse(resp))
                .then(country => {
                    mapP.countryId = country.id;
                    return mapP;
                })
            : mapP
        )
        .then(mapP => {
            wwMap.init({
                countryId: mapP.countryId,
                defaultCenter: mapP.center,
                useHash: mapP.countryId ? false : true,
                mapOptions: mapP.mapOptions,
            })
            if (mapP.geoObject) {
                wwMap.yMap.geoObjects.add(mapP.geoObject);
            }
            return wwMap;
        });
}

export function show_river_info_popup(id) {
    $.get(apiBase + "/river-card/" + id, function (data) {
        var dataObj = JSON.parse(data);
        dataObj.canEdit = canEdit();
        dataObj.apiUrl = apiBase + "/downloads/river";
        dataObj.apiBase = apiBase;
        riverList.riverInfoPopup.show(dataObj);
    });
    return false;
}

export function show_spot_report_popup(id, title, riverTitle) {
    var dataObject = {
        object_id: id,
        object_title: title,
        title: riverTitle
    };
    getWwmapUserInfoForMapControls().then(info => {
        if (info && info.login) {
            dataObject.user = info.login;
        }
        spotReportPopup.show(dataObject)
    }, err => spotReportPopup.show(dataObject));
}

export function show_camp_report_popup(id, title) {
    var dataObject = {
        object_id: id,
        object_title: title,
        title: title
    };
    getWwmapUserInfoForMapControls().then(info => {
        if (info && info.login) {
            dataObject.user = info.login;
        }
        campReportPopup.show(dataObject)
    }, err => campReportPopup.show(dataObject));
}
