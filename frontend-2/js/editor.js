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

function getSpotMainImageUrl(spotId) {
    img =  doGetJsonSync(backendApiBase + "/spot/" + spotId + "/preview")
    if (img) {
        return img.url
    }
    return null
}

function setSpotPreview(spotId, imgId) {
    return doPostJsonSync(backendApiBase + "/spot/" + spotId + "/preview", imgId, true)
}

function dropSpotPreview(spotId) {
    return doDeleteSync(backendApiBase + "/spot/" + spotId + "/preview", true)
}

function isActive(countryId, regionId, riverId, spotId) {
    hash = window.location.hash
    if (!hash) {
        return false
    }
    hash = hash.substr(1)
    var params = hash.split(',')

    if (spotId && riverId && countryId) {
        return params.length >= 4
            && parseInt(params[0])==countryId
            && parseInt(params[1])==regionIdNvl(regionId)
            && parseInt(params[2])==riverId
            && parseInt(params[3])==spotId
    } else if (countryId && riverId) {
        return params.length >= 3
            && parseInt(params[0])==countryId
            && parseInt(params[1])==regionIdNvl(regionId)
            && parseInt(params[2])==riverId
    } else if (countryId && regionId) {
        return params.length >= 4
            && parseInt(params[0])==countryId
            && parseInt(params[1])==regionId
    } else if (countryId) {
        return params.length >= 4
            && parseInt(params[0])==countryId
    }
    return false
}

function regionIdNvl(regionId) {
    if (regionId) {
        return regionId
    } else {
        return 0
    }
}

function nvlReturningId(region) {
    if (region && region.id) {
        return region.id
    } else {
        return 0
    }
}

const COUNTRY_ACTIVE_ENTITY_LEVEL=1
const REGION_ACTIVE_ENTITY_LEVEL=2
const RIVER_ACTIVE_ENTITY_LEVEL=3
const SPOT_ACTIVE_ENTITY_LEVEL=4

function getActiveEntityLevel() {
    var hash = window.location.hash
    if (hash) {
        hash = hash.substr(1)
        var params = hash.split(',')
        return params.length
    }
    return 0
}

function isActiveEntity(countryId, regionId, riverId, spotId) {
    var hash = window.location.hash
    if (hash) {
        hash = hash.substr(1)
        var params = hash.split(',')
        if(!isEq(arguments, params, COUNTRY_ACTIVE_ENTITY_LEVEL)) {
            return false
        }
        if(regionId && !isEq(arguments, params, REGION_ACTIVE_ENTITY_LEVEL)) {
            return false
        }
        if (riverId && !isEq(arguments, params, RIVER_ACTIVE_ENTITY_LEVEL)) {
            return false
        }
        if (spotId &&  !isEq(arguments, params, SPOT_ACTIVE_ENTITY_LEVEL)) {
            return false
        }
        return true
    }
    return false
}

function getActiveId(level) {
    var hash = window.location.hash
    if (hash) {
        hash = hash.substr(1)
        var params = hash.split(',')
        var pos = level - 1
        return getFromEntityHash(params, pos)
    }
    return 0
}

function isEq(args, params, level) {
    var pos = level - 1
    return getFromEntityHash(params, pos) == args[pos]
}

function getFromEntityHash(params, pos) {
    if (params && params.length > pos && params[pos]) {
        var intVal = parseInt(params[pos])
        if (intVal) {
            return intVal
        }
    }
    return 0
}

function setActiveEntity(countryId, regionId, riverId, spotId) {
    var hash = getActiveEntityHash(countryId, regionId, riverId, spotId)
    window.location.hash = hash
}

function getActiveEntityHash(countryId, regionId, riverId, spotId) {
    hash = countryId

    if (regionId) {
        hash += "," + regionId
    } else if (riverId) {
        hash += ",0"
    } else {
        return hash
    }

    if (riverId) {
        hash += "," + riverId
    } else {
          return hash
    }

    if (spotId) {
        hash += "," + spotId
    }
    return hash
}
