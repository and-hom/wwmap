import {SHOW_CAMPS_MIN_ZOOM, SHOW_SLOPE_MIN_ZOOM} from "wwmap-js-commons/constants";

export const FEATURE_SHOW_CAMPS = 0;
export const FEATURE_SHOW_UNPUBLISHED = 1;
export const FEATURE_SHOW_SLOPE = 2;
export const FEATURE_SHOW_ALTITUDE_COVERAGE = 3;

export function FeatureToggles(
    selectedValues = [true, false, true],
    enabled = [true, false, false],
    buttons = [null, null, null],
    minZoom = [SHOW_CAMPS_MIN_ZOOM, 0, SHOW_SLOPE_MIN_ZOOM],
    needsAuth = [false, true, true],
    zoom = 0,
) {
    this.size = selectedValues.length;
    this.state = 0;
    this.allowed = 0;
    this.minZoomMask = 0;
    this.needsAuthMask = 0;

    for (let i = 0; i < this.size; i++) {
        if (selectedValues[i]) {
            this.state |= (1 << i)
        }
        if (enabled == null || i >= enabled.length || i < enabled.length && enabled[i]) {
            this.allowed |= (1 << i)
        }
        if (minZoom == null || i >= minZoom.length || i < minZoom.length && zoom >= minZoom[i]) {
            this.minZoomMask |= (1 << i)
        }
        if (needsAuth == null || i >= needsAuth.length || i < needsAuth.length && needsAuth[i]) {
            this.needsAuthMask |= (1 << i)
        }
    }

    this.buttons = buttons;
    this.minZoom = minZoom;

    this.refreshButtonsEnabled();
}

FeatureToggles.prototype.setZoom = function (zoom) {
    let prevZoomMask = this.minZoomMask;

    this.zoom = zoom;
    this.refreshZoomMask();
    this.refreshButtonsEnabled();

    return this.minZoom != prevZoomMask;
}

FeatureToggles.prototype.refreshZoomMask = function () {
    this.minZoomMask = 0;

    for (let i = 0; i < this.size; i++) {
        if (this.minZoom == null || i >= this.minZoom.length || i < this.minZoom.length && this.zoom >= this.minZoom[i]) {
            this.minZoomMask |= (1 << i)
        }
    }
}

FeatureToggles.prototype.refreshButtonsEnabled = function () {
    if (!this.buttons) {
        return;
    }
    for (let i = 0; i < this.size; i++) {
        if (this.buttons.length > i && this.buttons[i]) {
            let mask = 1 << i;
            if ((mask & this.allowed) == 0 || (mask & this.minZoomMask) == 0) {
                this.buttons[i].state.set('selected', false);
                this.buttons[i].disable();
            } else {
                this.buttons[i].enable();
                this.buttons[i].state.set('selected', (mask & this.state) != 0);
            }
        }
    }
}

FeatureToggles.prototype.get = function (idx) {
    if (idx >= this.size) {
        throw 'Unsupported bitmask index ' + idx;
    }
    let mask = 1 << idx;
    return (this.allowed & mask) != 0 && (this.state & mask) != 0 && (this.minZoomMask & mask) != 0;
}

FeatureToggles.prototype.set = function (idx, val) {
    if (idx >= this.size) {
        throw 'Unsupported bitmask index ' + idx;
    }
    let mask = 1 << idx;

    if ((this.allowed & mask) == 0) {
        log.error("Disallowed to switch flag at position " + idx)
        return;
    }

    if (val) {
        this.state |= mask;
    } else {
        this.state &= ~mask;
    }

    if (this.buttons && this.buttons.length > idx && this.buttons[idx]) {
        this.buttons[idx].state.set('selected', val);
    }
}

FeatureToggles.prototype.getShowCamps = function () {
    return this.get(FEATURE_SHOW_CAMPS);
}

FeatureToggles.prototype.setShowCamps = function (showCamps) {
    this.set(FEATURE_SHOW_CAMPS, showCamps);
}

FeatureToggles.prototype.getShowUnpublished = function () {
    return this.get(FEATURE_SHOW_UNPUBLISHED);
}

FeatureToggles.prototype.setShowUnpublished = function (showUnpublished) {
    this.set(FEATURE_SHOW_UNPUBLISHED, showUnpublished);
}

FeatureToggles.prototype.getShowSlope = function () {
    return this.get(FEATURE_SHOW_SLOPE);
}

FeatureToggles.prototype.setShowSlope = function (showSlope) {
    this.set(FEATURE_SHOW_SLOPE, showSlope);
}

FeatureToggles.prototype.getShowAltitudeCoverage = function () {
    return this.get(FEATURE_SHOW_ALTITUDE_COVERAGE);
}

FeatureToggles.prototype.setShowAltitudeCoverage = function (showAltitudeCoverage) {
    this.set(FEATURE_SHOW_ALTITUDE_COVERAGE, showAltitudeCoverage);
}

FeatureToggles.prototype.getNeedsAuth = function () {
    return (this.needsAuthMask & this.state & this.allowed & this.minZoomMask) != 0
}

FeatureToggles.prototype.serialize = function () {
    return (this.state & this.allowed & this.minZoomMask).toString(2).padStart(this.size, '0');
}

FeatureToggles.prototype.parse = function (toggles) {
    this.state = parseInt(toggles, 2)
    this.size = toggles.length
}