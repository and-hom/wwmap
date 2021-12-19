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

export function createUrlPart(key, value, first) {
    if (value != null) {
        let firstChar = first ? '?' : '&';
        return `${firstChar}${key}=${value}`;
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