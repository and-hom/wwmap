LAST_POS_COOKIE_NAME = "last-map-pos";
LAST_ZOOM_COOKIE_NAME = "last-map-zoom";
LAST_MAP_TYPE_COOKIE_NAME = "last-map-type";

// environment dependent section
apiBase = "https://wwmap.ru/api";
// end

CACHED_TILES_TEMPLATE = 'http://wwmap.ru/maps/###/%z/%x/%y.png';

GOOGLE_SAT_TILES = 'http://khms' + Math.floor(Math.random()) % 4 + '.google.com/kh/v=845&src=app&x=%x&y=%y&z=%z&s=Gal';
THUNDERFOREST_OUTDOOR_TILES = 'http://a.tile.thunderforest.com/outdoors/%z/%x/%y.png';
THUNDERFOREST_LANDSCAPE_TILES = 'http://a.tile.thunderforest.com/landscape/%z/%x/%y.png';

MAP_FRAGMENTS_URL = 'https://wwmap.ru/map-components/map-html-components.htm';

function defaultPosition() {
    return [78, 46];
}

function getLastPosition() {
    let lastPos = $.cookie(LAST_POS_COOKIE_NAME);
    if (lastPos) {
        return JSON.parse(lastPos)
    } else {
        return defaultPosition()
    }
}

function setLastPosition(pos) {
    $.cookie(LAST_POS_COOKIE_NAME, JSON.stringify(pos), {path: '/'})
}

function defaultZoom() {
    return 3;
}

function getLastZoom() {
    let lastZoom = $.cookie(LAST_ZOOM_COOKIE_NAME);
    if (lastZoom) {
        return JSON.parse(lastZoom)
    } else {
        return defaultZoom()
    }
}

function setLastZoom(z) {
    $.cookie(LAST_ZOOM_COOKIE_NAME, JSON.stringify(z), {path: '/'})
}

function setLastPositionZoomType(pos, z, type) {
    setLastPosition(pos);
    setLastZoom(z);
    setLastMapType(type);
    window.location.hash = pos[0] + ',' + pos[1] + ',' + z + ',' + type.replace('#','-')
}

function getLastPositionAndZoom() {
    var position;
    var zoom;
    var type;
    var hash = window.location.hash;
    if (hash) {
        hash = hash.substr(1);
        var params = hash.split(',');
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

function defaultMapType() {
    return "osm#standard";
}

function getLastMapType() {
    let lastMapType = $.cookie(LAST_MAP_TYPE_COOKIE_NAME);
    if (lastMapType) {
        let lastMapTypeStr = JSON.parse(lastMapType);
        if (ymaps.mapType.storage.get(lastMapTypeStr)) {
            return lastMapTypeStr
        }
    }
    return defaultMapType()
}

function setLastMapType(z) {
    $.cookie(LAST_MAP_TYPE_COOKIE_NAME, JSON.stringify(z), {path: '/'})
}

bingTileUrlCache = {};

function bingSatTiles(tile, zoom) {
    let key = tile[0] + "_" + tile[1] + "_" + zoom;
    let cached = bingTileUrlCache[key];
    if (cached) {
        return cached;
    }

    let version = 0;
    let res = '';
    let osX = Math.pow(2, zoom) / 2;
    let osY = Math.pow(2, zoom) / 2;
    let prX = osX;
    let prY = osY;
    for (let i = 2; i <= zoom+1; i++) {
        prX = Math.floor(prX / 2);
        prY = Math.floor(prY / 2);
        if (tile[0] < osX) {
            osX -= prX;
            if (tile[1] < osY) {
                osY -= prY;
                res += '0';
            } else {
                osY += prY;
                res += '2';
            }
        } else {
            osX += prX;
            if (tile[1] < osY) {
                osY -= prY;
                res += '1';
            } else {
                osY += prY;
                res += '3';
            }
        }
    }
    let url = 'http://ecn.t0.tiles.virtualearth.net/tiles/a' + res + '.jpeg?g=' + version;
    bingTileUrlCache[key] = url;
    return url;
}