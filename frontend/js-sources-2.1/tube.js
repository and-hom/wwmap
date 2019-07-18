ymaps.modules.define("overlay.BiPlacemark", [
    'overlay.Polygon',
    'overlay.Placemark',
    'util.extend',
    'event.Manager',
    'option.Manager',
    'Event',
    'geometry.pixel.Polygon',
    'geometry.pixel.Point'
], function (provide, PolygonOverlay, PlacemarkOverlay, extend, EventManager, OptionManager, Event, PolygonGeometry, PointGeometry) {
    var domEvents = [
            'click',
            'contextmenu',
            'dblclick',
            'mousedown',
            'mouseenter',
            'mouseleave',
            'mousemove',
            'mouseup',
            'multitouchend',
            'multitouchmove',
            'multitouchstart',
            'wheel'
        ],

        BiPlacemarkOverlay = function (pixelGeometry, data, options) {
            this.events = new EventManager();
            this.options = new OptionManager(options);
            this._map = null;
            this._data = data;
            this._geometry = pixelGeometry;
            this._frontOverlay = null;
            this._backOverlay = null;
        };

    BiPlacemarkOverlay.prototype = extend(BiPlacemarkOverlay.prototype, {
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
            var polyLine = this._createBiPlacemarkContours();
            if (polyLine.length > 15) {
                this._frontOverlay = new PolygonOverlay(new PolygonGeometry(polyLine.path));
                this._backOverlay = new PolygonOverlay(new PolygonGeometry(polyLine.path));
                this._startOverlayListening();

                this._frontOverlay.options.set("strokeWidth", 1.5);
                this._frontOverlay.options.set("fill", false);

                this._backOverlay.options.setParent(this.options);
                this._backOverlay.options.set("strokeWidth", 2.5);
                this._backOverlay.options.set("fill", false);
                this._backOverlay.options.set("strokeColor", "#444444FF");
                this._backOverlay.setMap(this.getMap());

                this._frontOverlay.options.setParent(this.options);
                this._frontOverlay.setMap(this.getMap());
            } else {
                this._frontOverlay = new PlacemarkOverlay(new PointGeometry(this.getGeometry().getCoordinates()[0]));
                this._startOverlayListening();

                this._frontOverlay.options.setParent(this.options);
                this._frontOverlay.options.set("layout", 'default#image');
                this._frontOverlay.options.set("imageHref", this.options.get("iconImageHref"));
                this._frontOverlay.options.set("imageSize", this.options.get("iconImageSize"));
                this._frontOverlay.options.set("imageOffset", this.options.get("iconImageOffset"));
                this._frontOverlay.setMap(this.getMap());
            }
        },

        _onRemoveFromMap: function () {
            this._frontOverlay.setMap(null);
            this._frontOverlay.options.setParent(null);

            if (this._backOverlay) {
                this._backOverlay.setMap(null);
                this._backOverlay.options.setParent(null);
            }

            this._stopOverlayListening();
        },

        _startOverlayListening: function () {
            this._frontOverlay.events.add(domEvents, this._onDomEvent, this);
            if (this._backOverlay) {
                this._backOverlay.events.add(domEvents, this._onDomEvent, this);
            }
        },

        _stopOverlayListening: function () {
            this._frontOverlay.events.remove(domEvents, this._onDomEvent, this);
            if (this._backOverlay) {
                this._backOverlay.events.remove(domEvents, this._onDomEvent, this);
            }
        },

        _onDomEvent: function (e) {
            this.events.fire(e.get('type'), new Event({target: this}, e));
        },


        minMod: function (px1, hypotenuseL, hypotenuseR) {
            return px1 > 0 ? Math.min(hypotenuseL, hypotenuseR, px1) : -Math.min(hypotenuseL, hypotenuseR, -px1);
        },

        _createBiPlacemarkContours: function () {
            let placemarkRadius = 16;

            let path = this.getGeometry().getCoordinates();
            let dx = [];
            let dy = [];
            for (let i = 0; i < path.length - 1; i++) {
                let dx_i = path[i + 1][0] - path[i][0];
                let dy_i = path[i + 1][1] - path[i][1];

                if (Math.abs(dx_i) < placemarkRadius && Math.abs(dy_i) < placemarkRadius) {
                    // if distance less then radius, remove point
                    path.splice(i + 1, 1);
                    i--;
                } else {
                    dx.push(dx_i);
                    dy.push(dy_i);
                }
            }

            let beginPoint = path[0];
            let endPoint = path[path.length - 1];
            let dxDirect = endPoint[0] - beginPoint[0];
            let dyDirect = endPoint[1] - beginPoint[1];
            let lengthDirect = Math.sqrt(dxDirect * dxDirect + dyDirect * dyDirect);

            let dxBegin = dx[0];
            let dyBegin = dy[0];
            let beginSegmentLength = Math.sqrt(dxBegin * dxBegin + dyBegin * dyBegin);
            // rotation matrix for begin rounding
            let m_0_0_begin = dxBegin / beginSegmentLength;
            let m_0_1_begin = dyBegin / beginSegmentLength;

            let dxEnd = dx[dx.length - 1];
            let dyEnd = dy[dy.length - 1];
            let endSegmentLen = Math.sqrt(dxEnd * dxEnd + dyEnd * dyEnd);
            // rotation matrix for end rounding
            let m_0_0_end = dxEnd / endSegmentLen;
            let m_0_1_end = dyEnd / endSegmentLen;


            let halfCircleBegin = [];
            let halfCircleEnd = [];
            for (let theta = Math.PI; theta >= 0; theta -= 0.1) {
                let x = placemarkRadius * Math.cos(Math.PI / 2 - theta);
                let y = placemarkRadius * Math.sin(Math.PI / 2 - theta);

                let xRealA = (m_0_0_begin * x + m_0_1_begin * y);
                let yRealA = (m_0_0_begin * y - m_0_1_begin * x);
                halfCircleBegin.push([beginPoint[0] - xRealA, beginPoint[1] + yRealA]);

                let xRealB = (m_0_0_end * x + m_0_1_end * y);
                let yRealB = (m_0_0_end * y - m_0_1_end * x);
                halfCircleEnd.push([endPoint[0] + xRealB, endPoint[1] - yRealB]);
            }

            let forwardLine = [];
            let backLine = [];

            for (let i = 0; i < dx.length - 1; i++) {
                let fi1 = Math.atan2(-dx[i], -dy[i]);
                let fi2 = Math.atan2(dx[i + 1], dy[i + 1]);
                let fi = fi1 - fi2;
                fi = fi < 0 ? 2 * Math.PI + fi : fi; // angle between segments in current mid point

                let lengthLeft = Math.sqrt(dx[i] * dx[i] + dy[i] * dy[i]);
                // rotation matrix for prev segment (before point)
                let m_0_0_l = dx[i] / lengthLeft;
                let m_0_1_l = dy[i] / lengthLeft;

                let lengthRight = Math.sqrt(dx[i + 1] * dx[i + 1] + dy[i + 1] * dy[i + 1]);
                // rotation matrix for next segment (after point)
                let m_0_0_r = dx[i + 1] / lengthRight;
                let m_0_1_r = dy[i + 1] / lengthRight;

                let px1 = 0; //segment top and bottom sides offset to prevent line cross on bending
                let px2 = 0;
                if (fi > Math.PI) {
                    px1 = placemarkRadius / Math.tan(Math.PI - (2 * Math.PI - fi) / 2);
                    px1 = this.minMod(px1, lengthLeft, lengthRight)
                } else {
                    px2 = placemarkRadius / Math.tan(Math.PI - fi / 2);
                    px2 = this.minMod(px2, lengthLeft, lengthRight)
                }

                if (px2 != 0) {
                    // rounding on middle point
                    for (let theta = 0; theta < Math.PI - fi; theta += 0.1) {
                        let x = placemarkRadius * Math.cos(Math.PI / 2 + theta);
                        let y = placemarkRadius * Math.sin(Math.PI / 2 + theta);

                        let xReal = (m_0_0_l * x + m_0_1_l * y);
                        let yReal = (m_0_0_l * y - m_0_1_l * x);
                        forwardLine.push([path[i + 1][0] - xReal, path[i + 1][1] + yReal]);
                    }
                } else {
                    forwardLine.push([path[i + 1][0] + m_0_0_l * px1 - m_0_1_l * placemarkRadius, path[i + 1][1] + m_0_0_l * placemarkRadius + m_0_1_l * px1],);
                    forwardLine.push([path[i + 1][0] - m_0_0_r * px1 - m_0_1_r * placemarkRadius, path[i + 1][1] + m_0_0_r * placemarkRadius - m_0_1_r * px1],);
                }

                if (px1 != 0) {
                    // rounding on middle point
                    for (let theta = fi; theta > Math.PI; theta -= 0.1) {
                        let x = placemarkRadius * Math.cos(Math.PI / 2 + theta);
                        let y = placemarkRadius * Math.sin(Math.PI / 2 + theta);

                        let xReal = (m_0_0_r * x + m_0_1_r * y);
                        let yReal = (m_0_0_r * y - m_0_1_r * x);
                        backLine.unshift([path[i + 1][0] - xReal, path[i + 1][1] + yReal]);
                    }
                } else {
                    backLine.unshift([path[i + 1][0] + m_0_0_l * px2 + m_0_1_l * placemarkRadius, path[i + 1][1] - m_0_0_l * placemarkRadius + m_0_1_l * px2],);
                    backLine.unshift([path[i + 1][0] - m_0_0_r * px2 + m_0_1_r * placemarkRadius, path[i + 1][1] - m_0_0_r * placemarkRadius - m_0_1_r * px2],);
                }
            }

            return {
                path: [halfCircleBegin.concat(forwardLine).concat(halfCircleEnd).concat(backLine)],
                length: lengthDirect,
            };
        }
    });

    provide(BiPlacemarkOverlay);
});
