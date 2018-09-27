YNDX_AUTH_URL = "https://oauth.yandex.ru/authorize?response_type=token&client_id=f50947e6ab4944e1b1c14f2a21f76271"
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
    window.localStorage.removeItem(TOKEN_FIELD)
}

function forceRedirectYndx() {
    window.location.href = YNDX_AUTH_URL + "&state=" + encodeURIComponent(window.location.href)
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

function extractTokenYndx() {
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

function storeTokenFromRequest() {
    token = extractTokenYndx()
    if (token) {
        setSourceAndToken('yandex', token)
    }
}

var cachedUserInfo = null
var userInfoWasCached = false

function getAuthorizedUserInfoOrNull() {
    if (userInfoWasCached) {
        return cachedUserInfo
    }

    sourceAndToken = getSourceAndToken()
//    if (!sourceAndToken) {
//        token = extractTokenYndx()
//    }
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