backendApiBase = "http://localhost:7007";
frontendVersion = "development";

function sendRequest(url, _type, auth) {
    var xhr = new XMLHttpRequest();
    xhr.open(_type, url, false);
    addAuth(xhr, auth);
    try {
        xhr.send();
        return xhr
    } catch (err) {
        console.log(err);
        return null
    }
}

function doGetJsonSync(url, auth) {
    var xhr = sendRequest(url, "GET", auth);
    if (xhr && xhr.status === 200) {
        return JSON.parse(xhr.response)
    }
    return null
}

function doDeleteSync(url, auth, failResponseBodyCallback) {
    var xhr = sendRequest(url, "DELETE", auth);
    let resp = (xhr && xhr.status === 200);
    if (xhr && failResponseBodyCallback) {
        failResponseBodyCallback(xhr.response, xhr.status)
    }
    return resp;
}

function doDeleteWithJsonRespSync(url, auth, failResponseBodyCallback) {
    var xhr = sendRequest(url, "DELETE", auth);
    if (xhr && xhr.status === 200) {
        return JSON.parse(xhr.response)
    }
    if (xhr && failResponseBodyCallback) {
        failResponseBodyCallback(xhr.response, xhr.status)
    }
    return null
}

function doPostJsonSync(url, value, auth, failResponseBodyCallback) {
    var xhr = new XMLHttpRequest();
    xhr.open('POST', url, false);
    xhr.setRequestHeader("Content-Type", "application/json");
    addAuth(xhr, auth);
    var data = JSON.stringify(value);
    xhr.send(data);
    if (xhr.status === 200) {
        return JSON.parse(xhr.response, xhr.status)
    }
    if (xhr && failResponseBodyCallback) {
        failResponseBodyCallback(xhr.response)
    }
    return null
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

function getBackendVersion() {
    try {
        return doGetJsonSync(backendApiBase + "/version")
    } catch (e) {
        console.log(e);
    }
}

function getDbVersion() {
    try {
        return doGetJsonSync(backendApiBase + "/db-version")
    } catch (e) {
        console.log(e);
    }
}


function parseParams(paramsStr) {
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