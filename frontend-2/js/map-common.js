const CACHED_TILES_TEMPLATE = 'http://wwmap.ru/maps/###/%z/%x/%y.png';
const GOOGLE_SAT_TILES = 'http://khms' + Math.floor(Math.random()) % 4 + '.google.com/kh/v=794&src=app&x=%x&y=%y&z=%z&s=Gal';

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
