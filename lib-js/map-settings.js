var $ = require("jquery");
require("jquery.cookie");

import {HashTool} from "./hash-tool";

export const LAST_POS_COOKIE_NAME = "last-editor-map-pos";
export const LAST_ZOOM_COOKIE_NAME = "last-editor-map-zoom";
export const LAST_MAP_TYPE_COOKIE_NAME = "last-editor-map-type";

export const DEFAULT_POSITION = [57, 55]
export const DEFAULT_ZOOM = 3
export const DEFAULT_MAP_TYPE = "osm#standard"

export const HASH_DELIMITER = '/'

export function MapParams(position, zoom, type) {
    this.position = position;
    this.zoom = zoom;
    this.type = type;
}

MapParams.prototype.isFilled = function () {
    return this.position != null && this.zoom != null && this.type != null
}

MapParams.prototype.addMissingFields = function (another) {
    if (another) {
        this.copyMissingField(another, "position")
        this.copyMissingField(another, "zoom")
        this.copyMissingField(another, "type")
    }
}
MapParams.prototype.copyMissingField = function (another, fieldName) {
    if (this[fieldName] == null) {
        this[fieldName] = another[fieldName]
    }
}
MapParams.prototype.clone = function () {
    let result = new MapParams();
    result.addMissingFields(this);
    return result;
}

MapParams.prototype.toCommaSeparatedString = function () {
    if (!this.isFilled()) {
        return ""
    }
    return this.position[0] + ',' + this.position[1] + ',' + this.zoom + ',' + this.type.replace('#', '-')
}

function fromArray(params) {
    let position = null
    let zoom = null
    let type = null

    if (params.length >= 2) {
        position = [parseFloat(params[0]), parseFloat(params[1])]
    }
    if (params.length >= 3) {
        zoom = parseInt(params[2])
    }
    if (params.length >= 4) {
        let t = params[3].replace('-', '#');
        if (ymaps.mapType.storage.get(t)) {
            type = t;
        }
    }

    return new MapParams(position, zoom, type)
}

export function createMapParamsStorage() {
    return new MapParamsStorage(
        new EditorHashMapParamsStorage(),
        new CookieMapParamsStorage(LAST_POS_COOKIE_NAME, LAST_ZOOM_COOKIE_NAME, LAST_MAP_TYPE_COOKIE_NAME),
        new DefaultMapParamsStorage(new MapParams(DEFAULT_POSITION, DEFAULT_ZOOM, DEFAULT_MAP_TYPE))
    );
}

export function MapParamsStorage(...storages) {
    this.storages = storages
}

MapParamsStorage.prototype.getLastPositionZoomType = function () {
    let result = new MapParams(null, null, null)
    for (let i = 0; i < this.storages.length; i++) {
        result.addMissingFields(this.storages[i].getLastPositionZoomType())
        if (result.isFilled()) {
            return result;
        }
    }
    return result
}

MapParamsStorage.prototype.setLastPositionZoomType = function (pos, z, type) {
    this.storages.forEach(s => s.setLastPositionZoomType(pos, z, type))
}

export function CookieMapParamsStorage(posCookieName, zoomCookieName, mapTypeCookieName) {
    this.posCookieName = posCookieName;
    this.zoomCookieName = zoomCookieName;
    this.mapTypeCookieName = mapTypeCookieName;
}

function getFromCookie(cookieName, defaultValOrFunc) {
    let valueStr = $.cookie(cookieName);
    let value = parseSafe(valueStr)
    if (value) {
        return value;
    } else if (defaultValOrFunc && typeof defaultValOrFunc === "function") {
        return defaultValOrFunc()
    } else {
        return defaultValOrFunc;
    }
}

function parseSafe(value) {
    try {
        return JSON.parse(value)
    } catch (e) {
        console.error(e)
        return null
    }
}

function setToCookie(cookieName, value) {
    $.cookie(cookieName, JSON.stringify(value), {path: '/'})
}

CookieMapParamsStorage.prototype.setLastPositionZoomType = function (pos, z, type) {
    setToCookie(this.posCookieName, pos)
    setToCookie(this.zoomCookieName, z)
    setToCookie(this.mapTypeCookieName, type)
}

CookieMapParamsStorage.prototype.getLastPositionZoomType = function () {
    getFromCookie(this.posCookieName, null)
    getFromCookie(this.zoomCookieName, null)
    getFromCookie(this.mapTypeCookieName, null)
}

export function EditorHashMapParamsStorage() {
    this.hashTool = new HashTool(HASH_DELIMITER)
}

EditorHashMapParamsStorage.prototype.setLastPositionZoomType = function (pos, z, type) {
    let data = new MapParams(pos, z, type).toCommaSeparatedString();
    this.hashTool.setHashAtPos(1, data)
}

EditorHashMapParamsStorage.prototype.getLastPositionZoomType = function () {
    return fromArray(this.hashTool.getHashAtPos(1, [], x => x.split(',')));
}


export function DefaultMapParamsStorage(mapParams) {
    this.mapParams = mapParams ? mapParams : new MapParams(DEFAULT_POSITION, DEFAULT_ZOOM, DEFAULT_MAP_TYPE)
}

DefaultMapParamsStorage.prototype.setLastPositionZoomType = function (pos, z, type) {
    // do nothing
}

DefaultMapParamsStorage.prototype.getLastPositionZoomType = function () {
    return this.mapParams.clone()
}
