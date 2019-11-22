import backendApiBase from './config'

export function sendRequest(url, _type, auth) {
    new Promise((resolve, reject) => {
        var xhr = new XMLHttpRequest();
        xhr.open(_type, url, true);
        addAuth(xhr, auth);

        xhr.onload = () => resolve(xhr.responseText);
        xhr.onerror = () => reject(xhr.statusText);

        try {
            xhr.send();
        } catch (err) {
            console.log(err);
            reject(err);
        }
    });
}

export function doGetJson(url, auth) {
    return sendRequest(url, "GET", auth).then(body => JSON.parse(body));
}

export function doDelete(url, auth) {
    return sendRequest(url, "DELETE", auth);
}

export function doDeleteWithJsonResp(url, auth) {
    return sendRequest(url, "DELETE", auth).then(body => JSON.parse(body));
}

export function doPostJson(url, value, auth) {
    return new Promise((resolve, reject) => {
        var xhr = new XMLHttpRequest();
        xhr.open('POST', url, false);
        xhr.setRequestHeader("Content-Type", "application/json");
        addAuth(xhr, auth);
        var data = JSON.stringify(value);

        xhr.onload = () => resolve(xhr.responseText);
        xhr.onerror = () => reject(xhr.statusText);

        xhr.send(data);
    })
}

function addAuth(xhr, auth) {
    if (!auth) {
        return
    }
    var sessionId = getWwmapSessionId();
    if (sessionId) {
        xhr.setRequestHeader("Authorization", sessionId);
    }
}

export function getBackendVersion() {
    return doGetJson(backendApiBase + "/version")
}

export function getDbVersion() {
    return doGetJsonS(backendApiBase + "/db-version")
}


export function parseParams(paramsStr) {
    if (!paramsStr) {
        return {}
    }

    var paramsArr = paramsStr.split('&');
    var params = {};
    for (var i = 0; i < paramsArr.length; i++) {
        var keyValue = paramsArr[i].split('=');
        if (keyValue.length < 2) {
            continue
        }
        var key = keyValue[0];
        params[key] = keyValue[1];
    }
    return params;
}