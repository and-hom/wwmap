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
        if (ymaps && ymaps.mapType && ymaps.mapType.storage && ymaps.mapType.storage.get(t)) {
            type = t;
        }
    }

    return new MapParams(position, zoom, type)
}

export function createMapParamsStorage() {
    return new MapParamsStorage(
        new EditorHashMapParamsStorage(1),
        new KeyValueMapParamsStorage(LAST_POS_COOKIE_NAME, LAST_ZOOM_COOKIE_NAME, LAST_MAP_TYPE_COOKIE_NAME, new JsonDataMapper(), new CookieDataStorage()),
        new DefaultMapParamsStorage(new MapParams(DEFAULT_POSITION, DEFAULT_ZOOM, DEFAULT_MAP_TYPE))
    );
}

export function MapParamsStorage(...storages) {
    this.storages = storages
}

MapParamsStorage.prototype.getLastPositionZoomType = function (discriminator) {
    let result = new MapParams(null, null, null)
    for (let i = 0; i < this.storages.length; i++) {
        result.addMissingFields(this.storages[i].getLastPositionZoomType(discriminator))
        if (result.isFilled()) {
            return result;
        }
    }
    return result
}

MapParamsStorage.prototype.setLastPositionZoomType = function (pos, z, type, discriminator = null) {
    this.storages.forEach(s => s.setLastPositionZoomType(pos, z, type, discriminator))
}

export function KeyValueMapParamsStorage(posCookieName, zoomCookieName, mapTypeCookieName, dataMapper, backend) {
    this.posCookieName = posCookieName;
    this.zoomCookieName = zoomCookieName;
    this.mapTypeCookieName = mapTypeCookieName;

    this.dataMapper = dataMapper;
    this.backend = backend;
}

KeyValueMapParamsStorage.prototype.getValue = function(cookieName, defaultValOrFunc) {
    let valueStr = this.backend.get(cookieName);
    let value = this.dataMapper.parse(valueStr)
    if (value) {
        return value;
    } else if (defaultValOrFunc && typeof defaultValOrFunc === "function") {
        return defaultValOrFunc()
    } else {
        return defaultValOrFunc;
    }
}

KeyValueMapParamsStorage.prototype.setValue = function (cookieName, value) {
    this.backend.set(cookieName, this.dataMapper.serialize(value))
}

KeyValueMapParamsStorage.prototype.setLastPositionZoomType = function (pos, z, type, discriminator) {
    this.setValue(this.posCookieName + (discriminator || ''), pos)
    this.setValue(this.zoomCookieName + (discriminator || ''), z)
    this.setValue(this.mapTypeCookieName + (discriminator || ''), type)
}

KeyValueMapParamsStorage.prototype.getLastPositionZoomType = function (discriminator) {
    return new MapParams(
        this.getValue(this.posCookieName + (discriminator || ''), null),
        this.getValue(this.zoomCookieName + (discriminator || ''), null),
        this.getValue(this.mapTypeCookieName + (discriminator || ''), null)
    )
}

export function EditorHashMapParamsStorage(hashPosition, hashDelimiter = HASH_DELIMITER) {
    this.hashTool = new HashTool(hashDelimiter)
    this.hashPosition = hashPosition
}

EditorHashMapParamsStorage.prototype.setLastPositionZoomType = function (pos, z, type) {
    let data = new MapParams(pos, z, type).toCommaSeparatedString();
    this.hashTool.setHashAtPos(this.hashPosition, data)
}

EditorHashMapParamsStorage.prototype.getLastPositionZoomType = function () {
    return fromArray(this.hashTool.getHashAtPos(this.hashPosition, [], x => x.split(',')));
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

export function PlainDataMapper() {

}

PlainDataMapper.prototype.parse = function (value, defaultValue) {
    return value || defaultValue;
}

PlainDataMapper.prototype.serialize = function (value) {
    return value;
}

export function JsonDataMapper() {

}

JsonDataMapper.prototype.parse = function (value) {
    try {
        return JSON.parse(value)
    } catch (e) {
        console.error(e)
        return null
    }
}

JsonDataMapper.prototype.serialize = function (value) {
    return JSON.stringify(value);
}

export function CookieDataStorage() {

}

CookieDataStorage.prototype.get = function (name) {
    return $.cookie(name)
}

CookieDataStorage.prototype.set = function (name, value) {
    return $.cookie(name, value, {path: '/'})
}

export function LocalStorageDataStorage() {

}

LocalStorageDataStorage.prototype.get = function (name) {
    return window.localStorage[name]
}

LocalStorageDataStorage.prototype.set = function (name, value) {
    return window.localStorage[name] = value
}
