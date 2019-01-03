function AuthSource(code, oauthUrl) {
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

YANDEX_AUTH = new AuthSource("yandex", "https://oauth.yandex.ru/authorize?response_type=token&client_id=f50947e6ab4944e1b1c14f2a21f76271");
GOOGLE_AUTH = new AuthSource("google", "https://accounts.google.com/o/oauth2/v2/auth?client_id=61884443528-vfpuce81u3ka0aithbpjn405avkjqrt9.apps.googleusercontent.com&redirect_uri=https%3A%2F%2Fwwmap.ru%2Fredirector-google.htm&response_type=token&scope=profile%20email");
VK_AUTH = new AuthSource("vk", "https://oauth.vk.com/authorize?client_id=6703809&display=page&redirect_uri=https://wwmap.ru/redirector-vk.htm");

WWMAP_SESSION_ID = 'wwmap_token';

function getWwmapSessionId() {
    return window.localStorage[WWMAP_SESSION_ID]
}

function clearSessionId() {
    window.localStorage.removeItem(WWMAP_SESSION_ID);
}

function parseParams(paramsStr) {
    var paramsArr = paramsStr.split('&');
    var params = {};
    for (var i=0; i < paramsArr.length; i++) {
        var keyValue = paramsArr[i].split('=');
        if( keyValue.length < 2 ) {
            continue
        }
        var key = keyValue[0];
        params[key] = keyValue[1];
    }
    return params;
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

function acquireTokenVk(code) {
    return doGetJsonSync(backendApiBase + "/vk/token?code=" + code)
}

function startWwmapSession(source, token) {
    var userInfo = getUserInfoByUrl(backendApiBase + '/session-start?provider=' + source + "&token=" + token)

    if(userInfo) {
        window.localStorage[WWMAP_SESSION_ID] = userInfo.session_id;
        cachedUserInfo = userInfo;
        userInfoWasCached = true;
    }
    return userInfo;
}

function getUserInfo(sessionId) {
    return getUserInfoByUrl(backendApiBase + '/user-info?session_id=' + sessionId)
}

function getUserInfoByUrl(url) {
    var xhr = new XMLHttpRequest();
    xhr.open('GET', url, false);
    xhr.send();
    if (xhr.status === 200) {
        return JSON.parse(xhr.response)
    }
    return null
}

function getTokenFromRequestAndStartWwmapSession(authSource) {
    var token = extractFieldFromHash('access_token');
    if (token) {
        startWwmapSession(authSource, token)
    }
}

var cachedUserInfo = null;
var userInfoWasCached = false;

function getAuthorizedUserInfoOrNull() {
    if (userInfoWasCached) {
        return cachedUserInfo
    }

    var userInfo;

    var wwmapSessionId = getWwmapSessionId();
    if (!wwmapSessionId) {
        return null;
    }
    userInfo = getUserInfo(wwmapSessionId);
    if (userInfo) {
        cachedUserInfo = userInfo;
        userInfoWasCached = true;
        return userInfo;
    }
    return null
}

function getWwmapUserLogin() {
    var user = getAuthorizedUserInfoOrNull();
    if (user) {
        return user.login
    }
    return null
}