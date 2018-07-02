YNDX_AUTH_URL = "https://oauth.yandex.ru/authorize?response_type=token&client_id=f50947e6ab4944e1b1c14f2a21f76271"
TOKEN_FIELD = 'token'

function getToken() {
    return window.localStorage[TOKEN_FIELD]
}

function setToken(token) {
    window.localStorage[TOKEN_FIELD] = token
}

function clearToken() {
    window.localStorage.removeItem(TOKEN_FIELD)
}

function forceRedirect() {
    window.location.href = YNDX_AUTH_URL + "&state=" + encodeURIComponent(window.location.href)
}

function requireIfNotAuthorized() {
    if (!getToken()) {
        forceRedirect()
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

function extractToken() {
    var hash = window.location.hash
    if (hash) {
        hash = hash.substr(1)
        params = parseParams(hash)
        if (params['access_token']) {
            return params['access_token']
        }
    }
    return null
}

function getUserInfo(token) {
    var xhr = new XMLHttpRequest();
    xhr.open('GET', apiBase + '/user-info?token=' + token, false);
    xhr.send();
    if (xhr.status == 200) {
        return JSON.parse(xhr.response)
    }
    return null
}

function storeTokenFromRequest() {
    token = extractToken()
    if (token) {
        setToken(token)
    }
}

var cachedUserInfo = null
var userInfoWasCached = false

function getAuthorizedUserInfoOrNull() {
    if (userInfoWasCached) {
        return cachedUserInfo
    }

    token = getToken()
    if (!token) {
        token = extractToken()
    }
    if (token) {
        setToken(token)
        try {
            userInfo = getUserInfo(token)
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