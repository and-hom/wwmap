import {backendApiBase} from './config'
import {doGetJson} from './api'

const YANDEX_AUTH = new AuthSource("yandex", "https://oauth.yandex.ru/authorize?response_type=token&client_id=f50947e6ab4944e1b1c14f2a21f76271");
const GOOGLE_AUTH = new AuthSource("google", "https://accounts.google.com/o/oauth2/v2/auth?client_id=61884443528-vfpuce81u3ka0aithbpjn405avkjqrt9.apps.googleusercontent.com&redirect_uri=https%3A%2F%2Fwwmap.ru%2Fredirector-google.htm&response_type=token&scope=profile%20email");
const VK_AUTH = new AuthSource("vk", "https://oauth.vk.com/authorize?client_id=6703809&display=page&redirect_uri=https://wwmap.ru/redirector-vk.htm&response_type=code");

const WWMAP_SESSION_ID = 'wwmap_token';

export function AuthSource(code, oauthUrl) {
    this.code = code;
    this.oauthUrl = oauthUrl;
}

AuthSource.prototype.authRedirect = function () {
    window.location.href = this.oauthUrl + "&state=" + encodeURIComponent(window.location.href)
};
AuthSource.prototype.redirectIfMatch = function (code) {
    if (this.code === code) {
        this.authRedirect();
    }
};

function getWwmapSessionId() {
    return window.localStorage[WWMAP_SESSION_ID]
}

export function clearSessionId() {
    window.localStorage.removeItem(WWMAP_SESSION_ID);
}

function extractFieldFromHash(tokenParamName) {
    var hash = window.location.hash;
    if (hash) {
        hash = hash.substr(1);
        var params = parseParams(hash);
        if (params[tokenParamName]) {
            return params[tokenParamName]
        }
    }
    return null
}

export function acquireTokenVk(code, callback) {
    doGetJson(backendApiBase + "/vk/token?code=" + code).then(authObj => callback(authObj))
}

function startWwmapSession(source, token) {
    return getUserInfoByUrl(backendApiBase + '/session-start?provider=' + source + "&token=" + token)
        .then(userInfo => {

            if (userInfo) {
                window.localStorage[WWMAP_SESSION_ID] = userInfo.session_id;
                cachedUserInfo = userInfo;
                userInfoWasCached = true;
            }
            return userInfo;
        });
}

function getUserInfo(sessionId) {
    return getUserInfoByUrl(backendApiBase + '/user-info?session_id=' + sessionId)
}

function getUserInfoByUrl(url) {
    return new Promise((resolve, reject) => {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', url, false);

        xhr.onload = () => resolve(xhr.responseText);
        xhr.onerror = () => reject(xhr.statusText);

        xhr.send();
    });
}

export function getTokenFromRequestAndStartWwmapSession(authSource) {
    var token = extractFieldFromHash('access_token');
    if (token) {
        startWwmapSession(authSource, token)
    }
}

var cachedUserInfo = null;
var userInfoWasCached = false;

export function getAuthorizedUserInfoOrNull() {
    if (userInfoWasCached) {
        return new Promise((resolve, reject) => cachedUserInfo)
    }

    var wwmapSessionId = getWwmapSessionId();
    if (!wwmapSessionId) {
        return null;
    }
    return  getUserInfo(wwmapSessionId)
        .then(userInfo => {
            cachedUserInfo = userInfo;
            userInfoWasCached = true;
            return userInfo;
        });
}
