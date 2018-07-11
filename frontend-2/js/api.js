const apiBase = "http://localhost:7007";

function doGetJsonSync(url) {
    var xhr = new XMLHttpRequest();
    xhr.open('GET', url, false);
    xhr.send();
    if (xhr.status == 200) {
        return JSON.parse(xhr.response)
    }
    return null
}

function doDeleteJsonSync(url, auth) {
    var xhr = new XMLHttpRequest();
    xhr.open('DELETE', url, false);
    if (auth) {
        xhr.setRequestHeader("Authorization", "Token " + getToken());
    }
    try {
        xhr.send();
    } catch (err) {
        return false
    }
    return (xhr.status == 200)
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