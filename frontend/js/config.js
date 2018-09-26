const LAST_POS_COOKIE_NAME = "last-map-pos";
const LAST_ZOOM_COOKIE_NAME = "last-map-zoom";
const LAST_MAP_TYPE_COOKIE_NAME = "last-map-type";

// environment dependent section
const apiBase = "http://localhost:7007";
// end

const CACHED_TILES_TEMPLATE = 'http://wwmap.ru/maps/###/%z/%x/%y.png';

const GOOGLE_SAT_TILES = 'http://khms' + Math.floor(Math.random()) % 4 + '.google.com/kh/v=802&src=app&x=%x&y=%y&z=%z&s=Gal';
const THUNDERFOREST_OUTDOOR_TILES = 'http://a.tile.thunderforest.com/outdoors/%z/%x/%y.png';
const THUNDERFOREST_LANDSCAPE_TILES = 'http://a.tile.thunderforest.com/landscape/%z/%x/%y.png';

const MAP_FRAGMENTS_URL = 'https://wwmap.ru/map-components/map-html-components.htm'

function getLastPosition() {
    lastPos = $.cookie(LAST_POS_COOKIE_NAME);
    if (lastPos) {
        return JSON.parse(lastPos)
    } else {
        return [78, 46]
    }
}

function setLastPosition(pos) {
    $.cookie(LAST_POS_COOKIE_NAME, JSON.stringify(pos), {path: '/'})
}

function getLastZoom() {
    lastZoom = $.cookie(LAST_ZOOM_COOKIE_NAME);
    if (lastZoom) {
        return JSON.parse(lastZoom)
    } else {
        return 3
    }
}

function setLastZoom(z) {
    $.cookie(LAST_ZOOM_COOKIE_NAME, JSON.stringify(z), {path: '/'})
}

function setLastPositionZoomType(pos, z, type) {
    setLastPosition(pos)
    setLastZoom(z)
    setLastMapType(type)
    window.location.hash = pos[0] + ',' + pos[1] + ',' + z + ',' + type.replace('#','-')
}

function getLastPositionAndZoom() {
    var position;
    var zoom;
    var type;
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
        if (params.length >= 4) {
            type = params[3].replace('-','#')
        }
    }

    if (!position) {
        position = getLastPosition()
    }
    if (!zoom) {
        zoom = getLastZoom()
    }
    if (!type) {
        type = getLastMapType();
    }

    return {
        "position": position,
        "zoom": zoom,
        "type": type
    }
}

function getLastMapType() {
    lastMapType = $.cookie(LAST_MAP_TYPE_COOKIE_NAME);
    if (lastMapType) {
        return JSON.parse(lastMapType)
    } else {
        return "osm#standard"
    }
}

function setLastMapType(z) {
    $.cookie(LAST_MAP_TYPE_COOKIE_NAME, JSON.stringify(z), {path: '/'})
}