import {backendApiBase, cronApiBase, tCacheVersionUrl} from './config'
import {getWwmapSessionId} from "wwmap-js-commons/auth";

export function sendRequest(url, _type, auth) {
    return new Promise((resolve, reject) => {
        var xhr = new XMLHttpRequest();
        xhr.open(_type, url, true);
        addAuth(xhr, auth);

        xhr.onload = () => onLoad(xhr, resolve, reject);
        xhr.onerror = () => reject({
            "statusText": xhr.responseText,
            "status": xhr.status
        });

        try {
            xhr.send();
        } catch (err) {
            console.log(err);
            reject(err);
        }
    });
}

export function doGet(url, auth) {
    return sendRequest(url, "GET", auth);
}

export function doGetJson(url, auth) {
    return doGet(url, auth).then(JSON.parse);
}

export function doDelete(url, auth) {
    return sendRequest(url, "DELETE", auth);
}

export function doDeleteWithJsonResp(url, auth) {
    return sendRequest(url, "DELETE", auth).then(JSON.parse);
}

export function doPostJson(url, value, auth) {
    return doPost(url, value, auth).then(JSON.parse)
}

export function doPost(url, value, auth) {
    return new Promise((resolve, reject) => {
        var xhr = new XMLHttpRequest();
        xhr.open('POST', url, true);
        xhr.setRequestHeader("Content-Type", "application/json");
        addAuth(xhr, auth);
        var data = JSON.stringify(value);

        xhr.onload = () => onLoad(xhr, resolve, reject);
        xhr.onerror = () => reject({
            "statusText": xhr.responseText,
            "status": xhr.status
        });

        try {
            xhr.send(data);
        } catch (err) {
            console.log(err);
            reject(err);
        }
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


function onLoad(xhr, resolve, reject) {
    if (xhr.status / 100 != 2) {
        reject({
            "statusText": xhr.responseText,
            "status": xhr.status
        });
        return;
    }
    resolve(xhr.responseText);
}

export function getBackendVersion() {
    return doGetJson(backendApiBase + "/version")
}

export function getTCacheVersion() {
    return doGetJson(tCacheVersionUrl)
}

export function getCronVersion() {
    return doGetJson(cronApiBase + "/version")
}

export function getDbVersion() {
    return doGetJson(backendApiBase + "/db-version")
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