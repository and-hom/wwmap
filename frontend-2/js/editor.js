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

function getCountries() {
    return doGetJsonSync(backendApiBase + "/country")
}

function getRegions(countryId) {
    return doGetJsonSync(backendApiBase + "/country/" + countryId + "/region")
}

function getRegion(regionId) {
    return doGetJsonSync(backendApiBase + "/region/" + regionId)
}

function getRiversByCountry(countryId) {
    return doGetJsonSync(backendApiBase + "/country/" + countryId + "/river")
}

function getRiversByRegion(countryId, regionId) {
    return doGetJsonSync(backendApiBase + "/country/" + countryId + "/region/" + regionId + "/river")
}

function getReports(riverId) {
    return doGetJsonSync(backendApiBase + "/river/" + riverId + "/reports")
}

function getSpots(riverId) {
    return doGetJsonSync(backendApiBase + "/river/" + riverId + "/spots")
}

function getSpot(spotId) {
    return doGetJsonSync(backendApiBase + "/spot/" + spotId)
}

function saveSpot(spot) {
    return doPostJsonSync(backendApiBase + "/spot/" + spot.id, spot, true)
}

function removeSpot(id) {
    return doDeleteSync(backendApiBase + "/spot/" + id, true)
}

function getAllRegions() {
    return doGetJsonSync(backendApiBase + "/region")
}


function getRiver(riverId) {
    return doGetJsonSync(backendApiBase + "/river/" + riverId)
}

function getRiverCenter(riverId) {
    return doGetJsonSync(backendApiBase + "/river/" + riverId + "/center")
}

function saveRiver(river) {
    return doPostJsonSync(backendApiBase + "/river/" + river.id, river, true)
}

function removeRiver(id) {
    return doDeleteSync(backendApiBase + "/river/" + id, true)
}

function getImages(id, _type) {
    var imgs = doGetJsonSync(backendApiBase + "/spot/" + id + "/img?type=" + _type)
    if (imgs) {
        return imgs
    } else {
        return []
    }
}

function removeImage(spotId, id, _type) {
    return doDeleteWithJsonRespSync(backendApiBase + "/spot/" + spotId + "/img/" + id + "?type=" + _type, true)
}

function setImageEnabled(spotId, id, enabled) {
    return doPostJsonSync(backendApiBase + "/spot/" + spotId + "/img/" + id + "/enabled", enabled, true)
}
