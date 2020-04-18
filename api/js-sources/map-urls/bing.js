var bingTileUrlCache = {};

export function bingSatTiles(tile, zoom) {
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