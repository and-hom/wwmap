import {doGetJson, doPostJson, doDelete, doDeleteWithJsonResp} from './api';
import {backendApiBase} from './config'

export const all_categories = [
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
];

export function getCountries() {
    return doGetJson(backendApiBase + "/country")
}

export function getRegions(countryId) {
    return doGetJson(backendApiBase + "/country/" + countryId + "/region")
}

export function getRegion(regionId) {
    return doGetJson(backendApiBase + "/region/" + regionId)
}

export function getRiversByCountry(countryId) {
    return doGetJson(backendApiBase + "/country/" + countryId + "/river")
}

export function getRiversByRegion(countryId, regionId) {
    return doGetJson(backendApiBase + "/country/" + countryId + "/region/" + regionId + "/river")
}

export function getReports(riverId) {
    return doGetJson(backendApiBase + "/river/" + riverId + "/reports")
}

export function getSpots(riverId) {
    return doGetJson(backendApiBase + "/river/" + riverId + "/spots")
}

export function getSpotsFull(riverId) {
    return doGetJson(backendApiBase + "/river/" + riverId + "/spots-full")
}

export function getSpot(spotId) {
    return doGetJson(backendApiBase + "/spot/" + spotId)
}

export function saveSpot(spot, failResponseBodyCallback) {
    return doPostJson(backendApiBase + "/spot/" + spot.id, spot, true, failResponseBodyCallback)
}

export function removeSpot(id, failResponseBodyCallback) {
    return doDelete(backendApiBase + "/spot/" + id, true, failResponseBodyCallback)
}

export function getAllRegions() {
    return doGetJson(backendApiBase + "/region")
}


export function getRiver(riverId) {
    return doGetJson(backendApiBase + "/river/" + riverId)
}

export function getRiverCenter(riverId) {
    var center = doGetJson(backendApiBase + "/river/" + riverId + "/center");
    if (center==null) {
        return [0,0]
    }
    return center
}

export function getRiverBounds(riverId) {
    var bounds = doGetJson(backendApiBase + "/river/" + riverId + "/bounds");
    if (bounds==null) {
        return [[0,0],[0,0]]
    }
    return bounds
}

export function emptyBounds(bounds) {
    return bounds[0][0] == bounds[1][0] || bounds[0][1] == bounds[1][1];
}

export function saveRiver(river) {
    return doPostJson(backendApiBase + "/river/" + river.id, river, true)
}

export function removeRiver(id) {
    return doDelete(backendApiBase + "/river/" + id, true)
}

export function setRiverVisible(riverId, visible) {
    return doPostJson(backendApiBase + "/river/" + riverId + "/visible", visible, true);
}

export function getImages(id, _type) {
    return doGetJson(backendApiBase + "/spot/" + id + "/img?type=" + _type);
}

export function removeImage(spotId, id, _type) {
    return doDeleteWithJsonResp(backendApiBase + "/spot/" + spotId + "/img/" + id + "?type=" + _type, true)
}

export function setImageEnabled(spotId, id, enabled, _type) {
    return doPostJson(backendApiBase + "/spot/" + spotId + "/img/" + id + "/enabled?type=" + _type, enabled, true)
}

export function getSpotMainImageUrl(spotId) {
    return doGetJson(backendApiBase + "/spot/" + spotId + "/preview").then(img => img.preview_url).catch(err => null);
}

export function setSpotPreview(spotId, imgId, _type) {
    return doPostJson(backendApiBase + "/spot/" + spotId + "/preview?type=" + _type, imgId, true)
}

export function dropSpotPreview(spotId) {
    return doDelete(backendApiBase + "/spot/" + spotId + "/preview", true)
}

export function getLogEntries(objectType, objectId) {
    return doGetJson(backendApiBase + "/log?object_type=" + objectType + "&object_id=" + objectId, true)
}

export function getMeteoPoints() {
    return doGetJson(backendApiBase + "/meteo-point", true)
}


export function addMeteoPoint(p) {
    return doPostJson(backendApiBase + "/meteo-point", p, true)
}

export function isActive(countryId, regionId, riverId, spotId) {
    hash = window.location.hash;
    if (!hash) {
        return false
    }
    hash = hash.substr(1);
    var params = hash.split(',');

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

export function regionIdNvl(regionId) {
    if (regionId) {
        return regionId
    } else {
        return 0
    }
}

export function nvlReturningId(region) {
    if (region && region.id && !region.fake) {
        return region.id
    } else {
        return 0
    }
}

export const COUNTRY_ACTIVE_ENTITY_LEVEL=1;
export const REGION_ACTIVE_ENTITY_LEVEL=2;
export const RIVER_ACTIVE_ENTITY_LEVEL=3;
export const SPOT_ACTIVE_ENTITY_LEVEL=4;

export function getActiveEntityLevel() {
    var hash = window.location.hash
    if (hash) {
        hash = hash.substr(1)
        var params = hash.split(',')
        return params.length
    }
    return 0
}

export function isActiveEntity(countryId, regionId, riverId, spotId) {
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

export function getActiveId(level) {
    var hash = window.location.hash
    if (hash) {
        hash = hash.substr(1)
        var params = hash.split(',')
        var pos = level - 1
        return getFromEntityHash(params, pos)
    }
    return 0
}

export function isEq(args, params, level) {
    var pos = level - 1
    return getFromEntityHash(params, pos) == args[pos]
}

export function getFromEntityHash(params, pos) {
    if (params && params.length > pos && params[pos]) {
        var intVal = parseInt(params[pos])
        if (intVal) {
            return intVal
        }
    }
    return 0
}

export function setActiveEntityUrlHash(countryId, regionId, riverId, spotId) {
    window.location.hash = createActiveEntityHash(countryId, regionId, riverId, spotId)
}

export function createActiveEntityHash(countryId, regionId, riverId, spotId) {
    var hash = countryId;

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
