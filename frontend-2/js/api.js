const apiBase = "http://localhost:7007";

function sendRequest(url , _type, auth) {
    var xhr = new XMLHttpRequest();
    xhr.open(_type, url, false);
    if (auth) {
        xhr.setRequestHeader("Authorization", "Token " + getToken());
    }
    try {
        xhr.send();
        return xhr
    } catch (err) {
        return null
    }
}

function doGetJsonSync(url) {
    var xhr = sendRequest(url, "GET", false)
    if (xhr && xhr.status == 200) {
        return JSON.parse(xhr.response)
    }
    return null
}

function doDeleteSync(url, auth) {
    var xhr = sendRequest(url, "DELETE", auth)
    return (xhr && xhr.status == 200)
}

function doDeleteWithJsonRespSync(url, auth) {
    var xhr = sendRequest(url, "DELETE", auth)
    if (xhr && xhr.status == 200) {
        return JSON.parse(xhr.response)
    }
    return null
}

function doPostJsonSync(url, value, auth) {
    var xhr = new XMLHttpRequest();
    xhr.open('POST', url, false);
    xhr.setRequestHeader("Content-Type", "application/json");
    if (auth) {
        xhr.setRequestHeader("Authorization", "Token " + getToken());
    }
    var data = JSON.stringify(value);
    xhr.send(data);
    if (xhr.status == 200) {
        return JSON.parse(xhr.response)
    }
    return null
}