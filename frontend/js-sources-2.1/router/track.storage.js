export function TrackStorage(apiBase) {
    this.apiBase = apiBase;
    this.bounds = [[0, 0], [0, 0]];
    this.rivers = {};
    this.zoom = 14;
}

TrackStorage.prototype.setBounds = function (rect, extPoint, zoom) {
    let extRect = extPoint ? extendBounds(rect, extPoint) : rect;
    if (!contains(this.bounds, extRect)) {
        let loadingRect = multiply(extRect, 3);
        this.loadRivers(loadingRect, zoom);
        this.bounds = loadingRect;
    }
    this.zoom = zoom;
};

TrackStorage.prototype.loadRivers = function (rect, zoom) {
    let t = this;
    let url = `${this.apiBase}/router-data?bbox=${rect[0][0]},${rect[0][1]},${rect[1][0]},${rect[1][1]}&zoom=${zoom}`;
    $.ajax(url).done(function(data) {
        t.rivers = data.tracks;
    });
};


function extendBounds(bounds, point) {
    if (bounds[0][0] < point[0] && point[0] < bounds[1][0] &&
        bounds[0][1] < point[1] && point[1] < bounds[1][1]) {
        return bounds;
    }
    return [
        [Math.min(bounds[0][0], point[0]), Math.min(bounds[0][1], point[1])],
        [Math.max(bounds[1][0], point[0]), Math.max(bounds[1][1], point[1])]
    ]
}

function contains(b1, b2) {
    return b1[0][0] < b2[0][0] && b1[1][0] > b2[1][0]
        && b1[0][1] < b2[0][1] && b1[1][1] > b1[1][1]
}

function multiply(rect, n) {
    let cx = (rect[0][0] + rect[1][0]) / 2;
    let cy = (rect[0][1] + rect[1][1]) / 2;
    let dx = rect[1][0] - rect[0][0];
    let dy = rect[1][1] - rect[0][1];
    return [[cx - dx * n / 2, cy - dy * n / 2], [cx + dx * n / 2, cy + dy * n / 2]]
}