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

function addMapLayers() {
    addCachedLayer('osm#standard', 'OSM (O)', 'OpenStreetMap contributors, CC-BY-SA', 'osm');
    addLayer('google#satellite', 'Спутник Google (G)', 'Изображения © DigitalGlobe,CNES / Airbus, 2018,Картографические данные © Google, 2018', GOOGLE_SAT_TILES);
    addLayer('bing#satellite', 'Спутник Bing (B)', 'Изображения © Майкрософт (Microsoft), 2019', bingSatTiles);
    addCachedLayer('ggc#standard', 'Топографическая карта (T)', '', 'ggc', 0, 15);
    // workaround to change Yandex Satellite map title
    try {
        ymaps.mapType.storage.get('yandex#satellite')._name = 'Спутник Yandex (Y)'
    } catch (err) {
        console.log(err)
    }
}

function registerMapSwitchLayersHotkeys(map) {
    $(document).keyup(function (e) {
        switch (e.key) {
            case 'g':
            case 'G':
                map.setType('google#satellite');
                break;
            case 'b':
            case 'B':
                map.setType('bing#satellite');
                break;
            case 'y':
            case 'Y':
                map.setType('yandex#satellite');
                break;
            case 'o':
            case 'O':
                map.setType('osm#standard');
                break;
            case 't':
            case 'T':
                map.setType('ggc#standard');
                break;
        }
    });
}
