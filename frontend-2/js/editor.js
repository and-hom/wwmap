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

function getRiversByCountry(countryId) {
    return doGetJsonSync(apiBase + "/country/" + countryId + "/river")
}

function getRiversByRegion(countryId, regionId) {
    return doGetJsonSync(apiBase + "/country/" + countryId + "/region/" + regionId + "/river")
}

function getRiver(riverId) {
    return doGetJsonSync(apiBase + "/river/" + riverId)
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

function getAllRegions() {
    return doGetJsonSync(apiBase + "/region")
}

function saveRiver(river) {
    return doPostJsonSync(apiBase + "/river/" + river.id, river, true)
}
