var $ = require("jquery");
require("jquery.cookie");

import {HashTool} from "./hash-tool";

export const LAST_POS_EDITOR_KEY = "last-editor-map-pos";
export const LAST_ZOOM_EDITOR_KEY = "last-editor-map-zoom";
export const LAST_MAP_TYPE_EDITOR_KEY = "last-editor-map-type";
export const LAST_TOGGLES_EDITOR_KEY = "last-editor-toggles";

export const DEFAULT_POSITION = [57, 55]
export const DEFAULT_ZOOM = 3
export const DEFAULT_MAP_TYPE = "osm#standard"
export const DEFAULT_TOGGLES = "0001"

export const HASH_DELIMITER = '/'

export function MapParams(position, zoom, type, toggles) {
    this.position = position;
    this.zoom = zoom;
    this.type = type;
    this.toggles = toggles;
}

MapParams.prototype.isFilled = function () {
    return this.position != null && this.zoom != null && this.type != null && this.toggles != null
}

MapParams.prototype.addMissingFields = function (another) {
    if (another) {
        this.copyMissingField(another, "position")
        this.copyMissingField(another, "zoom")
        this.copyMissingField(another, "type")
        this.copyMissingField(another, "toggles")
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
    return `${this.position[0]},${this.position[1]},${this.zoom},${this.type.replace('#', '-')},${this.toggles}`
}

function fromArray(params) {
    let position = null
    let zoom = null
    let type = null
    let toggles = null

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
    if (params.length >= 5) {
        toggles = params[4]
    }

    return new MapParams(position, zoom, type, toggles)
}

export function createMapParamsStorage() {
    return new MapParamsStorage(
        new EditorHashMapParamsStorage(1),
        new KeyValueMapParamsStorage(
            LAST_POS_EDITOR_KEY,
            LAST_ZOOM_EDITOR_KEY,
            LAST_MAP_TYPE_EDITOR_KEY,
            LAST_TOGGLES_EDITOR_KEY,
            new JsonDataMapper(),
            new LocalStorageDataStorage()
        ),
        new DefaultMapParamsStorage(new MapParams(DEFAULT_POSITION, DEFAULT_ZOOM, DEFAULT_MAP_TYPE))
    );
}

export function MapParamsStorage(...storages) {
    this.storages = storages
}

MapParamsStorage.prototype.getLastPositionZoomTypeToggles = function (discriminator) {
    let result = new MapParams(null, null, null)
    for (let i = 0; i < this.storages.length; i++) {
        result.addMissingFields(this.storages[i].getLastPositionZoomTypeToggles(discriminator))
        if (result.isFilled()) {
            return result;
        }
    }
    return result
}

MapParamsStorage.prototype.setLastPositionZoomTypeToggles = function (pos, z, type, toggles, discriminator = null) {
    this.storages.forEach(s => s.setLastPositionZoomTypeToggles(pos, z, type, toggles, discriminator))
}

export function KeyValueMapParamsStorage(posKey, zoomKey, mapTypeKey, togglesKey, dataMapper, backend) {
    this.posKey = posKey;
    this.zoomKey = zoomKey;
    this.mapTypeKey = mapTypeKey;
    this.togglesKey = togglesKey;

    this.dataMapper = dataMapper;
    this.backend = backend;
}

KeyValueMapParamsStorage.prototype.getValue = function(Key, defaultValOrFunc) {
    let valueStr = this.backend.get(Key);
    let value = this.dataMapper.parse(valueStr)
    if (value) {
        return value;
    } else if (defaultValOrFunc && typeof defaultValOrFunc === "function") {
        return defaultValOrFunc()
    } else {
        return defaultValOrFunc;
    }
}

KeyValueMapParamsStorage.prototype.setValue = function (Key, value) {
    this.backend.set(Key, this.dataMapper.serialize(value))
}

KeyValueMapParamsStorage.prototype.setLastPositionZoomTypeToggles = function (pos, z, type, toggles, discriminator) {
    this.setValue(this.posKey + (discriminator || ''), pos)
    this.setValue(this.zoomKey + (discriminator || ''), z)
    this.setValue(this.mapTypeKey + (discriminator || ''), type)
    this.setValue(this.togglesKey + (discriminator || ''), toggles)
}

KeyValueMapParamsStorage.prototype.getLastPositionZoomTypeToggles = function (discriminator) {
    return new MapParams(
        this.getValue(this.posKey + (discriminator || ''), null),
        this.getValue(this.zoomKey + (discriminator || ''), null),
        this.getValue(this.mapTypeKey + (discriminator || ''), null),
        this.getValue(this.togglesKey + (discriminator || ''), null)
    )
}

export function EditorHashMapParamsStorage(hashPosition, hashDelimiter = HASH_DELIMITER) {
    this.hashTool = new HashTool(hashDelimiter)
    this.hashPosition = hashPosition
}

EditorHashMapParamsStorage.prototype.setLastPositionZoomTypeToggles = function (pos, z, type, toggles) {
    let data = new MapParams(pos, z, type, toggles).toCommaSeparatedString();
    this.hashTool.setHashAtPos(this.hashPosition, data)
}

EditorHashMapParamsStorage.prototype.getLastPositionZoomTypeToggles = function () {
    return fromArray(this.hashTool.getHashAtPos(this.hashPosition, [], x => x.split(',')));
}


export function DefaultMapParamsStorage(mapParams) {
    this.mapParams = mapParams ? mapParams : new MapParams(DEFAULT_POSITION, DEFAULT_ZOOM, DEFAULT_MAP_TYPE)
}

DefaultMapParamsStorage.prototype.setLastPositionZoomTypeToggles = function (pos, z, type, toggles) {
    // do nothing
}

DefaultMapParamsStorage.prototype.getLastPositionZoomTypeToggles = function () {
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
    if (!value) {
        return null
    }
    try {
        return JSON.parse(value)
    } catch (e) {
        console.error(`Can't parse json from ${value}`, e)
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
