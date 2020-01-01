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

function getSpotsFull(riverId) {
    return doGetJsonSync(backendApiBase + "/river/" + riverId + "/spots-full")
}

function getSpot(spotId) {
    return doGetJsonSync(backendApiBase + "/spot/" + spotId)
}

function saveSpot(spot, failResponseBodyCallback) {
    return doPostJsonSync(backendApiBase + "/spot/" + spot.id, spot, true, failResponseBodyCallback)
}

function removeSpot(id, failResponseBodyCallback) {
    return doDeleteSync(backendApiBase + "/spot/" + id, true, failResponseBodyCallback)
}

function getAllRegions() {
    return doGetJsonSync(backendApiBase + "/region")
}


function getRiver(riverId) {
    return doGetJsonSync(backendApiBase + "/river/" + riverId)
}

function getRiverCenter(riverId) {
    var center = doGetJsonSync(backendApiBase + "/river/" + riverId + "/center");
    if (center==null) {
        return [0,0]
    }
    return center
}

function getRiverBounds(riverId) {
    var bounds = doGetJsonSync(backendApiBase + "/river/" + riverId + "/bounds");
    if (bounds==null) {
        return [[0,0],[0,0]]
    }
    return bounds
}

function emptyBounds(bounds) {
    return bounds[0][0] == bounds[1][0] || bounds[0][1] == bounds[1][1];
}

function saveRiver(river) {
    return doPostJsonSync(backendApiBase + "/river/" + river.id, river, true)
}

function removeRiver(id) {
    return doDeleteSync(backendApiBase + "/river/" + id, true)
}

function setRiverVisible(riverId, visible) {
    return doPostJsonSync(backendApiBase + "/river/" + riverId + "/visible", visible, true)
}

function prepareImgs(imgs) {
    if (imgs) {
        for (let i in imgs) {
            let d = imgs[i].date;
            if (d) {
                imgs[i].date = new Date(d)
            }
        }
        return imgs
    } else {
        return []
    }
}

function getImages(id, _type) {
    return prepareImgs(doGetJsonSync(backendApiBase + "/spot/" + id + "/img?type=" + _type));
}

function removeImage(spotId, id, _type) {
    return prepareImgs(doDeleteWithJsonRespSync(backendApiBase + "/spot/" + spotId + "/img/" + id + "?type=" + _type, true))
}

function setImageEnabled(spotId, id, enabled, _type) {
    return prepareImgs(doPostJsonSync(backendApiBase + "/spot/" + spotId + "/img/" + id + "/enabled?type=" + _type, enabled, true))
}

function setManualLevel(spotId, id, l) {
    return doPostJsonSync(backendApiBase + "/spot/" + spotId + "/img/" + id + "/manual-level", l, true)
}

function resetManualLevel(spotId, id, l) {
    return doDeleteWithJsonRespSync(backendApiBase + "/spot/" + spotId + "/img/" + id + "/manual-level",true)
}

function setImageDate(spotId, id, date) {
    return prepareImgs(doPostJsonSync(backendApiBase + "/spot/" + spotId + "/img/" + id + "/date", date, true))
}

function getSpotMainImageUrl(spotId) {
    img =  doGetJsonSync(backendApiBase + "/spot/" + spotId + "/preview");
    if (img) {
        return img.preview_url
    }
    return null
}

function setSpotPreview(spotId, imgId, _type) {
    return doPostJsonSync(backendApiBase + "/spot/" + spotId + "/preview?type=" + _type, imgId, true)
}

function dropSpotPreview(spotId) {
    return doDeleteSync(backendApiBase + "/spot/" + spotId + "/preview", true)
}

function getLogEntries(objectType, objectId) {
    return doGetJsonSync(backendApiBase + "/log?object_type=" + objectType + "&object_id=" + objectId, true)
}

function getMeteoPoints() {
    return doGetJsonSync(backendApiBase + "/meteo-point", true)
}


function addMeteoPoint(p) {
    return doPostJsonSync(backendApiBase + "/meteo-point", p, true)
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
    if (region && region.id && !region.fake) {
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
    var hash = createActiveEntityHash(countryId, regionId, riverId, spotId)
    window.location.hash = hash
}

function createActiveEntityHash(countryId, regionId, riverId, spotId) {
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
