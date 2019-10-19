export function TrackStorage(apiBase) {
    this.apiBase = apiBase;
    this.bounds = [[0, 0], [0, 0]];
    this.rivers = {};
    this.zoom = 14;
}

TrackStorage.prototype.setBounds = function (rect, zoom) {
    if (!contains(this.bounds, rect)) {
        let loadingRect = multiply(rect, 2);
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