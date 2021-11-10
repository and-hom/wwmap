import * as turf from "@turf/turf";
import {flip} from "../router/util";

const MAX_SLOPE = 200

ymaps.modules.define("overlay.RiverTrack", [
    'overlay.Polyline',
    'util.extend',
    'event.Manager',
    'option.Manager',
    'Event',
    'geometry.pixel.LineString',
], function (provide, PolylineOverlay, extend, EventManager, OptionManager, Event, PolylineGeometry) {
    var RiverTrackOverlay = function (pixelGeometry, data, options) {
        this.events = new EventManager();
        this.options = new OptionManager(options);
        this._map = null;
        this._data = data;
        this._geometry = pixelGeometry;
        this._segments = []
        this._heights = data.properties.heights;
    };

    RiverTrackOverlay.prototype = extend(RiverTrackOverlay.prototype, {
        getData: function () {
            return this._data;
        },

        setData: function (data) {
            if (this._data != data) {
                var oldData = this._data;
                this._data = data;
                this.events.fire('datachange', {
                    oldData: oldData,
                    newData: data
                });
            }
        },

        getMap: function () {
            return this._map;
        },

        setMap: function (map) {
            if (this._map != map) {
                var oldMap = this._map;
                if (!map) {
                    this._onRemoveFromMap();
                }
                this._map = map;
                if (map) {
                    this._onAddToMap();
                }
                this.events.fire('mapchange', {
                    oldMap: oldMap,
                    newMap: map
                });
            }
        },

        setGeometry: function (geometry) {
            if (this._geometry != geometry) {
                var oldGeometry = geometry;
                this._geometry = geometry;

                if (this.getMap() && geometry) {
                    this._rebuild();
                }
                this.events.fire('geometrychange', {
                    oldGeometry: oldGeometry,
                    newGeometry: geometry
                });
            }
        },

        getGeometry: function () {
            return this._geometry;
        },

        getShape: function () {
            return null;
        },

        isEmpty: function () {
            return false;
        },

        _rebuild: function () {
            this._onRemoveFromMap();
            this._onAddToMap();
        },

        _onAddToMap: function () {
            let prj = this._map.options.get('projection');
            let z = this._map.getZoom();

            let geoCoords = this._data.geometry.coordinates;
            let pixelCoords = this.getGeometry().getCoordinates()

            if (
                this._heights == null ||
                this._heights.length != geoCoords.length
            ) {
                this._pushSegment(pixelCoords, '#0000ff');
                return;
            }

            let reverse = this._heights[0] < this._heights[this._heights.length - 1]
            let slice = [pixelCoords[0]]
            let len = 0;
            let slope = 0;
            let prevSlope = 0;
            let firstSliceIdx = 0;
            for (let i = 1; i < pixelCoords.length; i++) {
                len += turf.distance(flip(geoCoords[i - 1]), flip(geoCoords[i]), {units: 'meters'});
                prevSlope = slope;
                slope = (this._heights[firstSliceIdx] - this._heights[i]) * 1000.0 / len;
                if (reverse) {
                    slope = -slope;
                }
                slice.push(pixelCoords[i])
                if (len > 1000 && slice.length > 4 && prevSlope != 0 && prevSlope != slope) {
                    if (slope < MAX_SLOPE) {
                        this._pushSegment(slice, this._segmentColor(prevSlope), prevSlope);
                    }
                    len = 0;
                    firstSliceIdx = i;
                    prevSlope = 0;
                    slice = [pixelCoords[i]];
                }
            }
            if (len > 500) {
                this._pushSegment(slice, this._segmentColor(prevSlope), prevSlope);
            }
        },

        _pushSegment: function (coords, color, slope) {
            let segment = new PolylineOverlay(new PolylineGeometry(coords));
            coords.forEach(c => {
                c[0] += 1;
                c[1] += 1;
            });
            segment.options.set("strokeWidth", 3);
            segment.options.set("fill", false);
            segment.options.set("strokeColor", color);
            segment.options.set("interactive", true);
            segment.options.set("hintContent", slope + " м/км");
            segment.options.set("zIndex", 10000);
            segment.options.set("zIndexActive", 10000);
            segment.options.setParent(this.options);
            segment.setMap(this.getMap());

            segment.events.add('mouseenter', e => {
                this._map.hint.open(e.get('coords'), Math.round(slope) + "м/км", {
                    // Опция: задержка перед открытием.
                    openTimeout: 500
                });
            });

            segment.events.add('mouseleave', e => {
                this._map.hint.close();
            });

            this._segments.push(segment);
        },

        _segmentColor: function (slope) {
            if (slope > 25) {
                return '#ff0000';
            }
            if (slope > 15) {
                return '#ffaa00';
            }
            if (slope > 10) {
                return '#ffff00';
            }
            if (slope > 7) {
                return '#aaff00';
            }
            return '#00ff00';
        },

        _onRemoveFromMap: function () {
            this._segments.forEach(segment => {
                segment.setMap(null);
                segment.options.setParent(null);
            })
        },

        _onDomEvent: function (e) {
            this.events.fire(e.get('type'), new Event({target: this}, e));
        },
    });

    provide(RiverTrackOverlay);
});
