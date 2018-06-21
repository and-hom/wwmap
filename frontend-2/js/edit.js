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

function getAllRegions() {
    return doGetJsonSync(apiBase + "/region")
}
