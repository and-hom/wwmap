import {CATALOG_LINK_TYPES, WWMap} from './map'
import {WWMapPopup} from "./popup";
import {RiverList} from "./riverList";
import {canEdit, getWwmapUserInfoForMapControls} from "./util";
import {apiBase} from './config';
import {loadFragment} from "./template-data";

import './style/map.css'
import './contrib/lightbox.min.css'
import './contrib/lightbox.min'

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
var reportPopup;
var tutorialPopup;
export var userInfoFunction = null;

export function initWWMap(mapId, riversListId, options) {
    let optDefined = typeof options == 'object';

    let catalogLinkType = optDefined ? options.catalogLinkType : null;
    let riversTemplateData = optDefined ? options.riversTemplateData : null;
    userInfoFunction = optDefined ? options.userInfoFunction : null;

    if (catalogLinkType && CATALOG_LINK_TYPES.indexOf(catalogLinkType) <= -1) {
        throw "Unknown catalog link type. Available are: " + CATALOG_LINK_TYPES
    }

    // initialize popup windows
    reportPopup = new WWMapPopup('report_popup', 'report_popup_template', {
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

    // init and show map
    ymaps.ready(function () {
        loadFragment('bubble_template').then(bubbleContent => {
            wwMap = new WWMap(mapId, bubbleContent, riverList, tutorialPopup, catalogLinkType);
            ymaps.modules.require(['overlay.BiPlacemark'], function (BiPlacemarkOverlay) {
                ymaps.overlay.storage.add("BiPlacemrakOverlay", BiPlacemarkOverlay);
                wwMap.init()
            });
        });
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

export function show_report_popup(id, title, riverTitle) {
    var dataObject = {
        object_id: id,
        object_title: title,
        title: riverTitle
    };
    getWwmapUserInfoForMapControls().then(info => {
        if (info && info.login) {
            dataObject.user = info.login;
        }
        reportPopup.show(dataObject)
    }, err => reportPopup.show(dataObject));
}
