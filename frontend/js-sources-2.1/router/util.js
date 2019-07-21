export function pointEq(p1, p2) {
    return p1[0] == p2[0] && p1[1] == p2[1];
}


export function flip(p) {
    return [p[1], p[0]]
}

export function flipLine(arr) {
    return arr.map(flip)
}

export function distanceText(lenMeters) {
    let distanceText;
    if (lenMeters < 1000) {
        distanceText = "" + lenMeters.toFixed(0) + "m";
    } else {
        distanceText = "" + (lenMeters / 1000).toFixed(2) + "km";
    }
    return distanceText;
}


export function mouseToCoords(evt) {
    let map = evt.get('map');
    let pixelPos = evt.get('position');
    let globalPxPos = map.converter.pageToGlobal(pixelPos);
    return map.options.get('projection').fromGlobalPixels(globalPxPos, map.getZoom());
}
