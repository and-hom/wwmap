const CACHED_TILES_TEMPLATE = 'http://wwmap.ru/maps/###/%z/%x/%y.png';
const GOOGLE_SAT_TILES = 'http://khms' + Math.floor(Math.random()) % 4 + '.google.com/kh/v=844&src=app&x=%x&y=%y&z=%z&s=Gal';

    function addCachedLayer(key, name, copyright, mapId, lower_scale, upper_scale) {
        return addLayer(key, name, copyright, CACHED_TILES_TEMPLATE.replace('###', mapId), lower_scale, upper_scale)
    }

    function addLayer(key, name, copyright, tilesUrlTemplate, lower_scale, upper_scale) {
        if (typeof(lower_scale) == "undefined") {
            lower_scale = 0
        }
        if (typeof(upper_scale) == "undefined") {
            upper_scale = 18
        }
        var layer = function () {
            var layer = new ymaps.Layer(tilesUrlTemplate, {
                projection: ymaps.projection.sphericalMercator,
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
        ymaps.layer.storage.add(key, layer)
        ymaps.mapType.storage.add(key, new ymaps.MapType(name, [key]));
    }

// todo remove duplicate with config.js
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
