const LAST_POS_COOKIE_NAME = "last-map-pos";
const LAST_ZOOM_COOKIE_NAME = "last-map-zoom";
const LAST_MAP_TYPE_COOKIE_NAME = "last-map-type";
const apiBase = "http://localhost:7007";

const STANDARD_TILES = 'http://tile.openstreetmap.org/%z/%x/%y.png';
const THUNDERFOREST_OUTDOOR_TILES = 'http://a.tile.thunderforest.com/outdoors/%z/%x/%y.png';
const THUNDERFOREST_LANDSCAPE_TILES = 'http://a.tile.thunderforest.com/landscape/%z/%x/%y.png';



function getLastPosition() {
    lastPos = $.cookie(LAST_POS_COOKIE_NAME);
    if (lastPos) {
        return $.parseJSON(lastPos)
    } else {
        return [55.76, 37.64]
    }
}

function setLastPosition(pos) {
    $.cookie(LAST_POS_COOKIE_NAME, $.toJSON(pos), {path: '/'})
}

function getLastZoom() {
    lastZoom = $.cookie(LAST_ZOOM_COOKIE_NAME);
    if (lastZoom) {
        return $.parseJSON(lastZoom)
    } else {
        return 7
    }
}

function setLastZoom(z) {
    $.cookie(LAST_ZOOM_COOKIE_NAME, $.toJSON(z), {path: '/'})
}

function setLastPositionAndZoom(pos, z) {
    setLastPosition(pos)
    setLastZoom(z)
    window.location.hash = pos[0] + ',' + pos[1] + ',' + z
}

function getLastPositionAndZoom() {
    var position;
    var zoom;
    var hash = window.location.hash
    if (hash) {
        hash = hash.substr(1)
        var params = hash.split(',')
        if (params.length >= 2) {
            position = [parseFloat(params[0]), parseFloat(params[1])]
        }
        if (params.length >= 3) {
            zoom = parseInt(params[2])
        }
    }

    if (!position) {
        position = getLastPosition()
    }
    if (!zoom) {
        zoom = getLastZoom()
    }

    return {
        "position": position,
        "zoom": zoom
    }
}

function getLastMapType() {
    lastMapType = $.cookie(LAST_MAP_TYPE_COOKIE_NAME);
    if (lastMapType) {
        return $.parseJSON(lastMapType)
    } else {
        return "osm#standard"
    }
}

function setLastMapType(z) {
    $.cookie(LAST_MAP_TYPE_COOKIE_NAME, $.toJSON(z), {path: '/'})
}