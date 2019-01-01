YNDX_AUTH_URL = "https://oauth.yandex.ru/authorize?response_type=token&client_id=f50947e6ab4944e1b1c14f2a21f76271"
GOOGLE_AUTH_URL = "https://accounts.google.com/o/oauth2/v2/auth?client_id=61884443528-vfpuce81u3ka0aithbpjn405avkjqrt9.apps.googleusercontent.com&redirect_uri=https%3A%2F%2Fwwmap.ru%2Fredirector-google.htm&response_type=token&scope=profile%20email"
VK_AUTH_URL = "https://oauth.vk.com/authorize?client_id=6703809&display=page&redirect_uri=https://wwmap.ru/redirector-vk.htm"
SOURCE_FIELD = 'auth_source'
TOKEN_FIELD = 'auth_token'

function getSourceAndToken() {
    var source = window.localStorage[SOURCE_FIELD]
    var token = window.localStorage[TOKEN_FIELD]
    if (!source || !token) {
        return null
    }
    return {
        source: source,
        token: token,
    }
}

function setSourceAndToken(source, token) {
    window.localStorage[SOURCE_FIELD] = source
    window.localStorage[TOKEN_FIELD] = token
}

function clearToken() {
    window.localStorage.removeItem(SOURCE_FIELD)
    window.localStorage.removeItem(TOKEN_FIELD)
}

function forceRedirectYndx() {
    window.location.href = YNDX_AUTH_URL + "&state=" + encodeURIComponent(window.location.href)
}

function forceRedirectGoogle() {
    window.location.href = GOOGLE_AUTH_URL + "&state=" + encodeURIComponent(window.location.href)
}

function forceRedirectVk() {
    window.location.href = VK_AUTH_URL + "&state=" + encodeURIComponent(window.location.href)
}

function requireIfNotAuthorized() {
    if (!getToken()[1]) {
        forceRedirectYndx()
    }
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
    token = extractFieldFromHash('access_token');
    if (token) {
        setSourceAndToken(authSource, token)
    }
}

var cachedUserInfo = null
var userInfoWasCached = false

function getAuthorizedUserInfoOrNull() {
    if (userInfoWasCached) {
        return cachedUserInfo
    }

    sourceAndToken = getSourceAndToken()
    if (sourceAndToken) {
//        setSourceAndToken(sourceAndToken.source, sourceAndToken.token)
        try {
            userInfo = getUserInfo(sourceAndToken.source, sourceAndToken.token)
            cachedUserInfo = userInfo
            userInfoWasCached = true
            return userInfo
         } catch (err) {
            console.error(err)
            return null
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