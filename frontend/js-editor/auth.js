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

SOURCE_FIELD = 'auth_source';
TOKEN_FIELD = 'auth_token';
EXPIRES_AT_FIELD = 'expires_at';

function getSourceAndToken() {
    var source = window.localStorage[SOURCE_FIELD];
    var token = window.localStorage[TOKEN_FIELD];
    var expires_at = window.localStorage[EXPIRES_AT_FIELD];
    if (!source || !token) {
        return null
    }
    return {
        source: source,
        token: token,
        expires_at: expires_at
    }
}

function setSourceAndToken(source, token, expires_at) {
    window.localStorage[SOURCE_FIELD] = source;
    window.localStorage[TOKEN_FIELD] = token;
    if (expires_at) {
        window.localStorage[EXPIRES_AT_FIELD] = expires_at;
    }
}

function clearToken() {
    window.localStorage.removeItem(SOURCE_FIELD);
    window.localStorage.removeItem(TOKEN_FIELD);
    window.localStorage.removeItem(EXPIRES_AT_FIELD);
}

function parseParams(paramsStr) {
    paramsArr = paramsStr.split('&')
    params = {}
    for (var i=0; i < paramsArr.length; i++) {
        keyValue = paramsArr[i].split('=')
        if( keyValue.length < 2 ) {
            continue
        }
        key = keyValue[0]
        value = keyValue[1]
        params[key] = value
    }
    return params;
}

function extractFieldFromHash(tokenParamName) {
    var hash = window.location.hash;
    if (hash) {
        hash = hash.substr(1);
        params = parseParams(hash);
        if (params[tokenParamName]) {
            return params[tokenParamName]
        }
    }
    return null
}

function acquireTokenVk(code) {
    return doGetJsonSync(backendApiBase + "/vk/token?code=" + code)
}

function getUserInfo(source, token) {
    var xhr = new XMLHttpRequest();
    xhr.open('GET', backendApiBase + '/user-info?provider=' + source + "&token=" + token, false);
    xhr.send();
    if (xhr.status == 200) {
        return JSON.parse(xhr.response)
    }
    return null
}

function storeTokenFromRequest(authSource) {
    var token = extractFieldFromHash('access_token');
    var expires_in = extractFieldFromHash('expires_in');
    var expires_at = null;
    if (expires_in) {
        expires_at = Date.now()/1000 + parseInt(expires_in);
    }
    if (token) {
        setSourceAndToken(authSource, token, expires_at)
    }
}

var cachedUserInfo = null;
var userInfoWasCached = false;

function getAuthorizedUserInfoOrNull() {
    if (userInfoWasCached) {
        return cachedUserInfo
    }

    var sourceAndToken = getSourceAndToken();
    if (sourceAndToken) {
        if (sourceAndToken.expires_at < Date.now() / 1000) {
            // token is expired - renew
            YANDEX_AUTH.redirectIfMatch(sourceAndToken.source);
            GOOGLE_AUTH.redirectIfMatch(sourceAndToken.source);
            VK_AUTH.redirectIfMatch(sourceAndToken.source);
        }

//        setSourceAndToken(sourceAndToken.source, sourceAndToken.token)
        try {
            var userInfo = getUserInfo(sourceAndToken.source, sourceAndToken.token);
            cachedUserInfo = userInfo;
            userInfoWasCached = true;
            return userInfo
         } catch (err) {
            console.error(err);
            return null;
         }
    } else {
        return null;
    }
}

function getWwmapUserLogin() {
    var user = getAuthorizedUserInfoOrNull();
    if (user) {
        return user.login
    }
    return null
}

function authHeaderVal(sourceAndToken) {
    return sourceAndToken.source + " " + sourceAndToken.token
}