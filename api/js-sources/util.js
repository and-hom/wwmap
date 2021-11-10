import {userInfoFunction} from './main'
import {getWwmapSessionId} from "wwmap-js-commons/auth";
import {tileBase} from "./config";

var $ = require( "jquery" );
require( "jquery.cookie" );

export const LAST_POS_COOKIE_NAME = "last-map-pos";
export const LAST_ZOOM_COOKIE_NAME = "last-map-zoom";
export const LAST_MAP_TYPE_COOKIE_NAME = "last-map-type";

export const CACHED_TILES_TEMPLATE = tileBase + '/###/%z/%x/%y.png';

export const GOOGLE_SAT_TILES = 'http://khms' + Math.floor(Math.random()) % 4 + '.google.com/kh/v=845&src=app&x=%x&y=%y&z=%z&s=Gal';
export const THUNDERFOREST_OUTDOOR_TILES = 'http://a.tile.thunderforest.com/outdoors/%z/%x/%y.png';
export const THUNDERFOREST_LANDSCAPE_TILES = 'http://a.tile.thunderforest.com/landscape/%z/%x/%y.png';


export function defaultPosition() {
    return [57, 55];
}

export function getLastPosition(discriminator, defaultPositionValue) {
    let lastPos = $.cookie(discriminator ? LAST_POS_COOKIE_NAME + discriminator : LAST_POS_COOKIE_NAME);
    if (lastPos) {
        return JSON.parse(lastPos)
    } else if (defaultPositionValue) {
        return defaultPositionValue
    } else {
        return defaultPosition()
    }
}

export function setLastPosition(pos, discriminator) {
    $.cookie(discriminator ? LAST_POS_COOKIE_NAME + discriminator : LAST_POS_COOKIE_NAME, JSON.stringify(pos), {path: '/'})
}

export function defaultZoom() {
    return 3;
}

export function getLastZoom(discriminator) {
    let lastZoom = $.cookie(discriminator ? LAST_ZOOM_COOKIE_NAME + discriminator : LAST_ZOOM_COOKIE_NAME);
    if (lastZoom) {
        return JSON.parse(lastZoom)
    } else {
        return defaultZoom()
    }
}

export function setLastZoom(z, discriminator) {
    $.cookie(discriminator ? LAST_ZOOM_COOKIE_NAME + discriminator : LAST_ZOOM_COOKIE_NAME, JSON.stringify(z), {path: '/'})
}

export function setLastPositionZoomType(pos, z, type, useHash, discriminator) {
    setLastPosition(pos, discriminator);
    setLastZoom(z, discriminator);
    setLastMapType(type, discriminator);
    if (useHash) {
        window.location.hash = pos[0] + ',' + pos[1] + ',' + z + ',' + type.replace('#', '-')
    }
}

export function getLastPositionAndZoom(discriminator, useHash, defaultPositionValue) {
    var position;
    var zoom;
    var type;
    var hash = window.location.hash;
    if (hash && useHash) {
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
        position = getLastPosition(discriminator, defaultPositionValue)
    }
    if (!zoom) {
        zoom = getLastZoom(discriminator)
    }
    if (!type) {
        type = getLastMapType(discriminator);
    }

    return {
        "position": position,
        "zoom": zoom,
        "type": type
    }
}

export function defaultMapType() {
    return "osm#standard";
}

export function getLastMapType(discriminator) {
    let lastMapType = $.cookie(discriminator ? LAST_MAP_TYPE_COOKIE_NAME + discriminator : LAST_MAP_TYPE_COOKIE_NAME);
    if (lastMapType) {
        let lastMapTypeStr = JSON.parse(lastMapType);
        if (ymaps.mapType.storage.get(lastMapTypeStr)) {
            return lastMapTypeStr
        }
    }
    return defaultMapType()
}

export function setLastMapType(z, discriminator) {
    $.cookie(discriminator ? LAST_MAP_TYPE_COOKIE_NAME + discriminator : LAST_MAP_TYPE_COOKIE_NAME, JSON.stringify(z), {path: '/'})
}

export function initMailtoLinks() {
    // initialize all mailto links: robots do not perform js, so this links will not be detected by robots
    let user = 'info';
    let domain = 'wwmap.ru';
    var emailLink = $('.email-link');
    emailLink.attr('href', 'mailto:' + user + '@' + domain);
    emailLink.text(user + '@' + domain)
}

export function getWwmapUserInfoForMapControls() {
        return new Promise(function (resolve, reject) {
            if (typeof userInfoFunction == 'function') {
                let userInfo = userInfoFunction();
                if (userInfo) {
                    resolve(userInfo)
                } else {
                    reject("Unauthorized")
                }
            }else {
                reject("Authorization is not enabled")
            }
        });
}

export function canEdit() {
    return getWwmapUserInfoForMapControls()
        .then(info =>
            (info && info.roles && ['EDITOR', 'ADMIN'].filter(function (r) {
                return info.roles.includes(r)
            }).length > 0))
        .catch(_ => false)
}

export function isMobileBrowser() {
    return !!(navigator.userAgent.match(/Android/i)
        || navigator.userAgent.match(/webOS/i)
        || navigator.userAgent.match(/iPhone/i)
        || navigator.userAgent.match(/iPad/i)
        || navigator.userAgent.match(/iPod/i)
        || navigator.userAgent.match(/BlackBerry/i)
        || navigator.userAgent.match(/Windows Phone/i));
}


export function createUnpublishedUrlPart(showUnpublished, first) {
    if (showUnpublished) {
        let sessionId = getWwmapSessionId();
        let firstChar = first ? '?' : '&';
        return `${firstChar}session_id=${sessionId}&show_unpublished=${showUnpublished}`;
    }
    return '';
}

export function createCampsUrlPart(showCamps, first) {
    if (showCamps) {
        let firstChar = first ? '?' : '&';
        return `${firstChar}show_camps=${showCamps}`;
    }
    return '';
}
export function createSlopeUrlPart(showSlope, first) {
    if (showSlope) {
        let sessionId = getWwmapSessionId();
        let firstChar = first ? '?' : '&';
        return `${firstChar}session_id=${sessionId}&show_slope=${showSlope}`;
    }
    return '';
}


export function createCountryUrlPart(countryId, first) {
    if (countryId) {
        let firstChar = first ? '?' : '&';
        return `${firstChar}country=${countryId}`;
    }
    return '';
}