import {flip} from "./util";
import {TrackStorage} from "./track.storage";
import * as turf from '@turf/turf';
import {MultiPath} from "./multi.path";
import {RiverTreeWalker} from "./tree.walker"


export function WWMapMeasurementTool(map, objectManager, apiBase) {
    this.enabled = false;
    this.edit = true;

    this.trackStorage = new TrackStorage(apiBase);

    this.map = map;
    this.objectManager = objectManager;

    this.pos = map.getCenter();

    this.multiPath = null;
    this.reset();

    $(document).keyup((e) => {
        if (e.key === "Escape" && this.edit) {
            this.multiPath.removeLastSegments(1);
        }
    });

    this.addEvents();
}

WWMapMeasurementTool.prototype.addEvents = function () {
    let t = this;
    this.objectManager.objects.events.add(['click'], e => {
        if (this.enabled && this.edit) {
            this.multiPath.pushEmptySegment();
        }
    });
    this.objectManager.objects.events.add('mousemove', e => {
        if (this.enabled && this.edit) {
            let coords = e.get('coords');
            if (coords) {
                this.onMouseMoved(t.coordsToMouse(coords), coords);
            }
        }
    });
    this.map.events.add('click', e => {
        if (this.enabled && this.edit) {
            this.multiPath.pushEmptySegment();
        }
    });
    this.map.events.add('boundschange', e => {
        if (this.enabled && this.edit) {
            this.onViewportChanged();
        }
    });
    this.map.events.add('mousemove', e => {
        if (this.enabled && this.edit) {
            this.onMouseMoved(e.get('position'), e.get('coords'));
        }
    });
};


WWMapMeasurementTool.prototype.enable = function () {
    this.multiPath.show();
    this.enabled = true;
    this.onViewportChanged();
};

WWMapMeasurementTool.prototype.disable = function () {
    this.multiPath.hide();
    this.enabled = false;
};

WWMapMeasurementTool.prototype.setEditMode = function (edit) {
    if (edit) {
        this.multiPath.showLast();
        this.onViewportChanged();
    } else {
        this.multiPath.hideLast();
    }
    this.edit = edit;
};

WWMapMeasurementTool.prototype.hasDrawnPath = function () {
    return this.multiPath.segmentCount() > 1;
};

WWMapMeasurementTool.prototype.reset = function () {
    if (this.multiPath) {
        this.multiPath.hide();
    }
    this.multiPath = new MultiPath(this.pos, this.map, this);
    if(this.enabled) {
        this.multiPath.show();
    }
    this.onViewportChanged();
};

WWMapMeasurementTool.prototype.onViewportChanged = function () {
    if (!this.enabled || !this.edit) {
        return;
    }
    let lastFixedPoint = this.multiPath.segmentCount() > 0 ? this.multiPath.pointEnd() : null;
    this.trackStorage.setBounds(this.map.getBounds(), lastFixedPoint, this.map.getZoom());
};

const sensitivity_px = 2;

WWMapMeasurementTool.prototype.getComputeMarkerPos = function (cursorPosFlipped, epsilon_m) {
    let markerPos;
    let minDist = epsilon_m * 2;
    this.currentLine = null;
    for (let id in this.trackStorage.rivers) {
        let river = this.trackStorage.rivers[id];
        let dst = turf.pointToLineDistance(cursorPosFlipped, turf.lineString(river.path), {units: 'meters'});
        if (dst < minDist) {
            minDist = dst;
            let nearestRiverPathFlipped = turf.lineString(river.path);
            let nearestPointFlipped = turf.nearestPointOnLine(nearestRiverPathFlipped, cursorPosFlipped, {units: 'meters'});
            markerPos = flip(nearestPointFlipped.geometry.coordinates);
            this.currentLine = river;
        }
    }
    if (!markerPos) {
        markerPos = flip(cursorPosFlipped);
    }
    return markerPos;
};

WWMapMeasurementTool.prototype.moveFirstPoint = function (cursorPosPx, coords, epsilon_m) {
    let cursorPosFlipped = flip(coords ? coords : this.mouseToCoords(cursorPosPx));

    if (this.currentLine) {
        let nearestRiverPathFlipped = turf.lineString(this.currentLine.path);
        let minDst = turf.pointToLineDistance(cursorPosFlipped, nearestRiverPathFlipped, {units: 'meters'});

        if (minDst < epsilon_m) {
            let nearestPointFlipped = turf.nearestPointOnLine(nearestRiverPathFlipped, cursorPosFlipped, {units: 'meters'});
            this.multiPath.setStartMarkerPos(flip(nearestPointFlipped.geometry.coordinates), this.currentLine.id);
            return;
        }
    }

    let markerPos = this.getComputeMarkerPos(cursorPosFlipped, epsilon_m);
    this.multiPath.setStartMarkerPos(markerPos, this.currentLine ? this.currentLine.id : -1);
};

WWMapMeasurementTool.prototype.onMouseMoved = function (cursorPosPx, coords) {
    if (!cursorPosPx ||  !this.enabled || !this.edit
        || this.trackStorage.rivers.length == 0 && !(this.multiPath.segmentCount() > 0 && this.currentLine)) {
        return
    }

    // mouse sensivity optimization
    if (this.pixelPos && cursorPosPx && (
        Math.abs(this.pixelPos[0] - cursorPosPx[0]) < sensitivity_px
        && Math.abs(this.pixelPos[1] - cursorPosPx[1]) <= sensitivity_px)) {
        return;
    }

    this.pixelPos = cursorPosPx;
    let epsilon_m = 100 / (2 ** (this.map.getZoom() - 16));

    // move first marker before drawing started
    if (this.multiPath.segmentCount() == 0) {
        this.moveFirstPoint(cursorPosPx, coords, epsilon_m);
        return;
    }

    let cursorPosYMap = coords ? coords : this.mouseToCoords(cursorPosPx);
    let cursorPosFlipped = flip(cursorPosYMap);

    let nearestRiver = this.currentLine;
    let nearestRiverLineString = nearestRiver ? turf.lineString(nearestRiver.path) : null;
    let minDst = nearestRiver
        ? turf.pointToLineDistance(cursorPosFlipped, nearestRiverLineString, {units: 'meters'})
        : Number.MAX_SAFE_INTEGER;

    // move tail marker  - same line
    if (nearestRiver && nearestRiver.id == this.multiPath.riverSegmentIdPrev() && minDst <= epsilon_m) {
        let nearestPointFlipped = turf.nearestPointOnLine(nearestRiverLineString, cursorPosFlipped, {units: 'meters'});
        let fromPoint = flip(this.multiPath.pointEnd());
        let toPoint = nearestPointFlipped.geometry.coordinates;
        let seg = turf.lineSlice(fromPoint, toPoint, nearestRiverLineString);
        if (turf.distance(seg.geometry.coordinates[0], toPoint, {units: "meters"}) < 10) {
            this.multiPath.setTrack(turf.flip(seg).geometry.coordinates.reverse(), nearestRiver.id, turf.length(seg, {units: 'meters'}));
        } else {
            this.multiPath.setTrack(turf.flip(seg).geometry.coordinates, nearestRiver.id, turf.length(seg, {units: 'meters'}));
        }
        return;
    }

    let markerPos = this.getComputeMarkerPos(cursorPosFlipped, epsilon_m);

    // search for sutable neighbout tracks
    if (this.currentLine) {
        let walker = new RiverTreeWalker(this.trackStorage, this.multiPath.riverSegmentIdPrev(), this.currentLine.id, 12);
        let found = walker.searchRiver();
        if (found) {
            this.currentLine = found.river;
            let path = this.makePath(flip(this.multiPath.pointEnd()), cursorPosFlipped, found);
            let pathLine = turf.lineString(path);
            this.multiPath.setTrack(turf.flip(pathLine).geometry.coordinates, found.river.id, turf.length(pathLine, {units: 'meters'}));
            return;
        }
    }

    // geometry direct
    this.multiPath.setLine(markerPos, nearestRiver ? nearestRiver.id : -1, minDst==Number.MAX_SAFE_INTEGER ?
        turf.distance(markerPos, this.multiPath.pointEnd(), {units: 'meters'}) : minDst);
};

WWMapMeasurementTool.prototype.mouseToCoords = function (pixelPos) {
    let globalPxPos = this.map.converter.pageToGlobal(pixelPos);
    return this.map.options.get('projection').fromGlobalPixels(globalPxPos, this.map.getZoom());
};

WWMapMeasurementTool.prototype.coordsToMouse = function (coords) {
    let prj = this.map.options.get('projection');
    let globalPx = prj.toGlobalPixels(coords, this.map.getZoom());
    return this.map.converter.globalToPage(globalPx);
};

WWMapMeasurementTool.prototype.getGeomNearestRiver = function (cursorPosFlipped) {
    let nr = null;
    let minDst = Number.MAX_SAFE_INTEGER;
    for (var id in this.trackStorage.rivers) {
        if (!this.trackStorage.rivers.hasOwnProperty(id)) {
            continue
        }
        let dst = turf.pointToLineDistance(cursorPosFlipped, turf.lineString(this.trackStorage.rivers[id].path), {units: 'meters'});
        if (minDst > dst) {
            minDst = dst;
            nr = this.trackStorage.rivers[id];
        }
    }
    return nr;
};

WWMapMeasurementTool.prototype.makePath = function (start, end, found) {
    let ids = found.path.concat([found.river.id]);
    let result = [];
    let fromPoint = start;

    for (let i = 0; i < ids.length; i++) {
        let id = ids[i];
        let track = this.trackStorage.rivers[id];
        let toPoint;
        if (i == ids.length - 1) {
            toPoint = end;
        } else {
            let refs = track.refs[ids[i + 1]];
            if (refs && refs.length > 0) {
                toPoint = refs[0];
            } else {
                return result;
            }
        }

        let path = track.path;
        let pathGeom = turf.lineString(path);
        let lSeg = turf.lineSlice(fromPoint, toPoint, pathGeom);
        let lSegCoords = lSeg.geometry.coordinates;
        if (turf.distance(fromPoint, lSegCoords[lSegCoords.length - 1], {units: "meters"}) < 30) {
            lSegCoords = lSegCoords.reverse();
        }

        result = result.concat(lSegCoords);
        fromPoint = toPoint;
    }
    return result;
};