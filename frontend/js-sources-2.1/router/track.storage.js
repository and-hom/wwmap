export function TrackStorage(apiBase) {
    this.apiBase = apiBase;
    this.bounds = [[0, 0], [0, 0]];
    this.rivers = {};
    this.zoom = 14;
    this.loadingCounter = 0;
    this.onLoadingChanged = null;
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


function doFetch(url) {
    return new Promise((resolve, reject) => {
        var xhr = new XMLHttpRequest();
        xhr.open("GET", url, true);

        xhr.onload = () => resolve(xhr.responseText);
        xhr.onerror = () => reject(xhr.statusText);

        try {
            xhr.send();
        } catch (err) {
            console.log(err);
            reject(err);
        }
    });
}

TrackStorage.prototype.loadRivers = function (rect, zoom) {
    let t = this;

    let x0 = rect[0][0];
    let x1 = rect[1][0];
    let y0 = rect[0][1];
    let y1 = rect[1][1];

    let xInt = Math.floor(x0 / 2);
    let yInt = Math.floor(y0 / 2);
    for (let x = xInt * 2; x < x1; x += 2) {
        for (let y = yInt * 2; y < y1; y += 2) {
            if (!t.rivers[`${x}_${y}`]) {
                let url = `${this.apiBase}/router-data?bbox=${x},${y},${x + 2},${y + 2}&zoom=${zoom}`;

                this.loadingCounter++;
                this.loadingChangedHander();
                doFetch(url)
                    .then(resp => JSON.parse(resp))
                    .then(data => t.rivers[`${x}_${y}`] = data.tracks)
                    .finally(() => {
                        this.loadingCounter--;
                        this.loadingChangedHander();
                    });
            }
        }
    }

    let keys = Object.keys(t.rivers);
    if (keys.length > 50) {
        for (let i in keys) {
            let key = keys[i];
            let parts = key.split("_");
            let tileX = parseInt(parts[0]);
            let tileY = parseInt(parts[1]);
            if (tileX < x0 - 1 || tileX > x1 + 1 || tileX < y0 - 1 || tileY > y1 + 1) {
                delete t.rivers[key]
            }
        }
    }
};


TrackStorage.prototype.loadingChangedHander = function() {
    if (this.onLoadingChanged) {
        this.onLoadingChanged(this.loadingCounter > 0)
    }
};

TrackStorage.prototype.getRiver = function (id) {
    let found;
    for (let polygonId in this.rivers) {
        found = this.rivers[polygonId][id];
        if (found) {
            return found;
        }
    }
    return null;
};

TrackStorage.prototype.getRivers = function (x, y, epsilonDegrees) {
    let xInt = Math.floor(x / 2);
    let yInt = Math.floor(y / 2);
    let segment = this.rivers[`${xInt * 2}_${yInt * 2}`];
    if (!segment) {
        return []
    }
    return segment
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