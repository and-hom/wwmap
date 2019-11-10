import {CATALOG_LINK_TYPES, WWMap} from './map'
import {WWMapPopup} from "./popup";
import {RiverList} from "./riverList";
import {canEdit, getWwmapUserInfoForMapControls, loadFragment} from "./util";
import {apiBase, MAP_FRAGMENTS_URL} from './config';

import './style/map.css'
import './style/lightbox.min.css'

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

export function initWWMap(mapId, riversListId, catalogLinkType) {
    initWWMapInternal(mapId, riversListId, catalogLinkType, true)
}

export function initWWMapCustomRiverList(mapId, riversListTemplateElement, catalogLinkType) {
    initWWMapInternal(mapId, riversListTemplateElement, catalogLinkType, false)
}

function initWWMapInternal(mapId, riversListId, catalogLinkType, riverListFromTemplates = true) {
    if (catalogLinkType && CATALOG_LINK_TYPES.indexOf(catalogLinkType) <= -1) {
        throw "Unknown catalog link type. Available are: " + CATALOG_LINK_TYPES
    }

    // initialize popup windows
    reportPopup = new WWMapPopup('report_popup_template', true, 'report_popup', {
        submitUrl: apiBase + "/report",
        okMsg: "Запрос отправлен. Я прочитаю его по мере наличия свободного времени",
        failMsg: "Что-то пошло не так...",
        // To prevent contents lost
        closeOnEscape: false,
        closeOnMouseClickOutside: false
    });
    tutorialPopup = new WWMapPopup('info_popup_template', true, 'info_popup');

    // riverList
    if (riversListId) {
        riverList = new RiverList(riversListId, 'rivers_template', riverListFromTemplates)
    }

    // init and show map
    ymaps.ready(function () {
        loadFragment(MAP_FRAGMENTS_URL, 'bubble_template', function (bubbleContent) {
            wwMap = new WWMap(mapId, extractInnerHtml(bubbleContent), riverList, tutorialPopup, catalogLinkType);
            ymaps.modules.require(['overlay.BiPlacemark'], function (BiPlacemarkOverlay) {
                ymaps.overlay.storage.add("BiPlacemrakOverlay", BiPlacemarkOverlay);
                wwMap.init()
            });
        })
    });
}

export function show_river_info_popup(id) {
    $.get(apiBase + "/river-card/" + id, function (data) {
        var dataObj = JSON.parse(data);
        dataObj.canEdit = canEdit();
        dataObj.apiUrl = apiBase + "/gpx/river";
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
    var info = getWwmapUserInfoForMapControls();
    if (info && info.login) {
        dataObject.user = info.login;
    }
    reportPopup.show(dataObject)
}

function extractInnerHtml(str) {
    return $(str).html()
}