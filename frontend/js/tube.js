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
                    this._frontOverlay.options.set("imageHref",this.options.get("iconImageHref"));
                    this._frontOverlay.options.set("imageSize",this.options.get("iconImageSize"));
                    this._frontOverlay.options.set("imageOffset",this.options.get("iconImageOffset"));
                    this._frontOverlay.setMap(this.getMap());
                }
            },

            _onRemoveFromMap: function () {
                this._frontOverlay.setMap(null);
                this._frontOverlay.options.setParent(null);

                if(this._backOverlay) {
                    this._backOverlay.setMap(null);
                    this._backOverlay.options.setParent(null);
                }

                this._stopOverlayListening();
            },

            _startOverlayListening: function () {
                this._frontOverlay.events.add(domEvents, this._onDomEvent, this);
                if(this._backOverlay) {
                    this._backOverlay.events.add(domEvents, this._onDomEvent, this);
                }
            },

            _stopOverlayListening: function () {
                this._frontOverlay.events.remove(domEvents, this._onDomEvent, this);
                if(this._backOverlay) {
                    this._backOverlay.events.remove(domEvents, this._onDomEvent, this);
                }
            },

            _onDomEvent: function (e) {
                this.events.fire(e.get('type'), new Event({target: this}, e));
            },

            _createBiPlacemarkContours: function () {

                var mainLineCoordinates = this.getGeometry().getCoordinates();
                var p1 = mainLineCoordinates[0];
                var p2 = mainLineCoordinates[mainLineCoordinates.length - 1];

                var dx = p2[0] - p1[0];
                var dy = p2[1] - p1[1];
                var hypotenuse = Math.sqrt(dx * dx + dy * dy);
                var m_0_0 = dx / hypotenuse;
                var m_0_1 = dy / hypotenuse;


                var radius = 16;

                var half_circle1 = [];
                var half_circle2 = [];
                for (var theta = 0; theta < Math.PI; theta += 0.1) {
                    var x = radius * Math.cos(Math.PI/2 - theta);
                    var y = radius * Math.sin(Math.PI/2 - theta);

                    var xReal = (m_0_0 * x + m_0_1 * y);
                    var yReal = (m_0_0 * y - m_0_1 * x);
                    half_circle1.push([p1[0] - xReal, p1[1] + yReal]);
                    half_circle2.push([p2[0] + xReal, p2[1] - yReal]);
                }

                var line1 = [
                    [p1[0] - m_0_1 * radius, p1[1] + m_0_0 * radius],
                    [p2[0] - m_0_1 * radius, p2[1] + m_0_0 * radius],
                ];
                var line2 = [
                    [p1[0] + m_0_1 * radius, p1[1] - m_0_0 * radius],
                    [p2[0] + m_0_1 * radius, p2[1] - m_0_0 * radius],
                ];
                return {
                    path: [half_circle2.concat(line1).concat(half_circle1).concat(line2)],
                    length: hypotenuse,
                };
            }
        });

    provide(BiPlacemarkOverlay);
});
