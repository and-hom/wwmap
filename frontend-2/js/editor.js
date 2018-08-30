const apiBase = "http://localhost:7007";

var all_categories = [
    {id:"-1", title:"Непроход."},
    {id:"0", title:"Неизестно"},
    {id:"1", title:"1"},
    {id:"2", title:"2"},
    {id:"3", title:"3"},
    {id:"4", title:"4"},
    {id:"4a", title:"   4a"},
    {id:"4b", title:"   4b"},
    {id:"4c", title:"   4c"},
    {id:"5", title:"5"},
    {id:"5a", title:"   5a"},
    {id:"5b", title:"   5b"},
    {id:"5c", title:"   5c"},
    {id:"6", title:"6"},
    {id:"6a", title:"   6a"},
    {id:"6b", title:"   6b"},
    {id:"6c", title:"   6c"},
]

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

function getCountries() {
    return doGetJsonSync(apiBase + "/country")
}

function getRegions(countryId) {
    return doGetJsonSync(apiBase + "/country/" + countryId + "/region")
}

function getRegion(regionId) {
    return doGetJsonSync(apiBase + "/region/" + regionId)
}

function getRiversByCountry(countryId) {
    return doGetJsonSync(apiBase + "/country/" + countryId + "/river")
}

function getRiversByRegion(countryId, regionId) {
    return doGetJsonSync(apiBase + "/country/" + countryId + "/region/" + regionId + "/river")
}

function getReports(riverId) {
    return doGetJsonSync(apiBase + "/river/" + riverId + "/reports")
}

function getSpots(riverId) {
    return doGetJsonSync(apiBase + "/river/" + riverId + "/spots")
}

function getSpot(spotId) {
    return doGetJsonSync(apiBase + "/spot/" + spotId)
}

function saveSpot(spot) {
    return doPostJsonSync(apiBase + "/spot/" + spot.id, spot, true)
}

function removeSpot(id) {
    return doDeleteSync(apiBase + "/spot/" + id, true)
}

function getAllRegions() {
    return doGetJsonSync(apiBase + "/region")
}


function getRiver(riverId) {
    return doGetJsonSync(apiBase + "/river/" + riverId)
}

function saveRiver(river) {
    return doPostJsonSync(apiBase + "/river/" + river.id, river, true)
}

function removeRiver(id) {
    return doDeleteSync(apiBase + "/river/" + id, true)
}

function getImages(id, _type) {
    var imgs = doGetJsonSync(apiBase + "/spot/" + id + "/img?type=" + _type)
    if (imgs) {
        return imgs
    } else {
        return []
    }
}

function removeImage(spotId, id, _type) {
    return doDeleteWithJsonRespSync(apiBase + "/spot/" + spotId + "/img/" + id + "?type=" + _type, true)
}

function setImageEnabled(spotId, id, enabled) {
    return doPostJsonSync(apiBase + "/spot/" + spotId + "/img/" + id + "/enabled", enabled, true)
}
