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
    return doDeleteJsonSync(apiBase + "/spot/" + id, true)
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
    return doDeleteJsonSync(apiBase + "/river/" + id, true)
}
