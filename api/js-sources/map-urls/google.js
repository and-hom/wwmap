import {apiBase} from "../config";
import {doGet} from "../api"

const DEFAULT_GOOGLE_API_VERSION = 865;
let GOOGLE_API_VERSION = -1;

function requestVersion() {
    doGet(apiBase + "/gav").then(response => {
        GOOGLE_API_VERSION = JSON.parse(response)
    });
}

requestVersion();

var googleTileUrlCache = {};

export function googleSatTiles(tile, zoom) {
    let key = tile[0] + "_" + tile[1] + "_" + zoom;
    let cached = googleTileUrlCache[key];
    if (cached && GOOGLE_API_VERSION > 0) {
        return cached;
    }

    let url = `https://khms${Math.floor(Math.random()) % 4}.google.com/kh/v=${GOOGLE_API_VERSION <= 0 ? DEFAULT_GOOGLE_API_VERSION : GOOGLE_API_VERSION}&src=app&x=${tile[0]}&y=${tile[1]}&z=${zoom}`;
    googleTileUrlCache[key] = url;
    return url;
}