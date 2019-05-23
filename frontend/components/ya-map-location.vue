<template>
    <div id="map" :style="mapDivStyle()"></div>
</template>

<script>
    module.exports = {
        props: {
            spot: Object,
            zoom: {
                type: Number,
                default: 12
            },
            width: {
                type: String,
                default: "600px"
            },
            height: {
                type: String,
                default: "400px"
            },
            editable: {
                type: Boolean,
                default: false
            },
            yaSearch: {
                type: Boolean,
                default: false
            },
            refreshOnChange: {
                default: null
            },
            defaultMap: {
                type: String,
                default: null
            },
        },
        watch: {
            // This would be called anytime the value of title changes
            refreshOnChange(newValue, oldValue) {
                this.doUpdate();
            }
        },
        updated: function () {
            this.doUpdate()
        },
        created: function () {
            var component = this;
            ymaps.ready(function () {
                ymaps.modules.require(['overlay.BiPlacemark'], function (BiPlacemarkOverlay) {
                    if (ymaps.overlay.storage.get("BiPlacemrakOverlay") == null) {
                        ymaps.overlay.storage.add("BiPlacemrakOverlay", BiPlacemarkOverlay);
                    }
                    if (component.map) {
                        component.map.destroy();
                        component.label.geometry.setCoordinates(component.getP(0));
                        return
                    } else {
                        addCachedLayer('osm#standard', 'OSM (O)', 'OpenStreetMap contributors, CC-BY-SA', 'osm');
                        addLayer('google#satellite', 'Спутник Google (G)', 'Изображения © DigitalGlobe,CNES / Airbus, 2018,Картографические данные © Google, 2018', GOOGLE_SAT_TILES);
                        addLayer('bing#satellite', 'Спутник Bing (B)', 'Изображения © Майкрософт (Microsoft), 2019', bingSatTiles);
                        addCachedLayer('ggc#standard', 'Топографическая карта (T)', '', 'ggc', 0, 15);

                        // workaround to change Yandex Satellite map title
                        try {
                            ymaps.mapType.storage.get('yandex#satellite')._name = 'Спутник Yandex (Y)'
                        } catch (err) {
                            console.log(err)
                        }
                    }

                    let myMap;
                    if (Array.isArray(component.spot.point[0])) {
                        myMap = new ymaps.Map("map", {
                            bounds: component.pBounds(),
                            controls: ["zoomControl"],
                            type: component.getDefaultMap(),
                        });
                    } else {
                        myMap = new ymaps.Map("map", {
                            center: component.pCenter(),
                            zoom: component.zoom,
                            controls: ["zoomControl"],
                            type: component.getDefaultMap(),
                        });
                    }
                    myMap.controls.add(
                        new ymaps.control.TypeSelector([
                            'osm#standard',
                            'ggc#standard',
                            'yandex#satellite',
                            'google#satellite',
                            'bing#satellite',
                        ])
                    );
                    $(document).keyup(function (e) {
                        switch(e.key) {
                            case 'g':
                            case 'G':
                                myMap.setType('google#satellite');
                                break;
                            case 'b':
                            case 'B':
                                myMap.setType('bing#satellite');
                                break;
                            case 'y':
                            case 'Y':
                                myMap.setType('yandex#satellite');
                                break;
                            case 'o':
                            case 'O':
                                myMap.setType('osm#standard');
                                break;
                            case 't':
                            case 'T':
                                myMap.setType('ggc#standard');
                                break;
                        }
                    });

                    myMap.controls.add('rulerControl', {
                        scaleLine: true
                    });
                    if (component.yaSearch) {
                        myMap.controls.add(new ymaps.control.SearchControl({
                            options: {
                                float: 'left',
                                floatIndex: 100
                            }
                        }));
                    }

                    myMap.events.add('typechange', function (e) {
                        component.setDefaultMap(myMap.getType());
                    });

                    component.map = myMap;


                    component.addObjectManager();
                    component.addLabels();
                    component.addClickHandler();

                });
            })
        },
        data: function () {
            return {
                mapDivStyle: function () {
                    return 'width: ' + this.width + '; height: ' + this.height + ';'
                },
                doUpdate: function () {
                    if (this.map) {
                        this.map.setCenter(this.pCenter());

                        this.map.geoObjects.remove(this.label);
                        if (this.endLabel) {
                            this.map.geoObjects.remove(this.endLabel);
                            this.map.geoObjects.remove(this.contour);
                        }
                        if (this.midPoints) {
                            for (let i = 0; i < this.midPoints.length; i++) {
                                this.map.geoObjects.remove(this.midPoints[i])
                            }
                        }
                        this.addLabels();
                        this.addClickHandler();

                        this.map.geoObjects.remove(this.objectManager);
                        this.addObjectManager();
                    }
                },
                addClickHandler: function () {
                    let t = this;
                    if (this.editable) {
                        this.map.events.add('click', function (e) {
                            let p = e.get('coords');
                            if (t.endLabel) {
                                t.endLabel.geometry.setCoordinates(p);
                                t.setP(-1, p);
                            } else {
                                t.label.geometry.setCoordinates(p);
                                t.setP(0, p);
                            }
                            t.refreshContour();
                        });
                    }
                },
                objectManagerUrlTemplate: function () {
                    var skip = 0;
                    if (this.spot && this.spot.id) {
                        skip = this.spot.id
                    }
                    url = backendApiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z&skip=' + skip;
                    st = getWwmapSessionId();
                    if (st) {
                        url += '&session_id=' + st
                    }
                    return url
                },
                addObjectManager: function () {
                    var objectManager = new ymaps.RemoteObjectManager(this.objectManagerUrlTemplate(), {
                            clusterHasBalloon: false,
                            geoObjectOpenBalloonOnClick: false,
                            geoObjectStrokeWidth: 3,
                            splitRequests: false
                        })
                    ;
                    this.map.geoObjects.add(objectManager);
                    this.objectManager = objectManager
                },
                addLabels: function () {
                    var component = this;
                    var props = {
                        hintContent: this.spot.title,
                    };
                    if (Array.isArray(this.spot.point[0])) {
                        props.iconContent = "Начало"
                    }
                    var label = new ymaps.GeoObject({
                        geometry: {
                            type: "Point",
                            coordinates: this.getP(0),
                        },
                        properties: props
                    }, {
                        preset: 'islands#blueStretchyIcon',
                        draggable: this.editable,
                        zIndex: 1000000,
                        zIndexHover: 1000000,
                        zIndexActive: 1000000,
                    });
                    label.events.add('dragend', function (e) {
                        component.setP(0, label.geometry.getCoordinates());
                        component.refreshContour();
                    });

                    if (Array.isArray(this.spot.point[0])) {
                        var contour = new ymaps.GeoObject({
                            geometry: {
                                type: "LineString",
                                coordinates: component.spot.point,
                            },
                            properties: {}
                        }, {
                            lineStringOverlay: "BiPlacemrakOverlay",
                            strokeColor: "#BBBBBBCC",
                            strokeStyle: 'shortdash',
                            strokeWidth: 3,
                            fill: false
                        });
                        component.map.geoObjects.add(contour);
                        component.contour = contour;

                        var endLabel = new ymaps.GeoObject({
                            geometry: {
                                type: "Point",
                                coordinates: this.getP(-1),
                            },
                            properties: {
                                hintContent: this.spot.title + " конец",
                                iconContent: "Конец"
                            }
                        }, {
                            preset: 'islands#yellowStretchyIcon',
                            draggable: this.editable,
                            zIndex: 10001000,
                            zIndexHover: 10001000,
                            zIndexActive: 10001000,
                        });
                        endLabel.events.add('dragend', function () {
                            component.setP(-1, endLabel.geometry.getCoordinates());
                            component.refreshContour();
                        });

                        if (this.spot.point.length > 2) {
                            let midPoints = [];
                            for (let i = 1; i < this.spot.point.length - 1; i++) {
                                let pos = i;
                                let properties = {
                                    hintContent: "" + i,
                                    iconContent: "" + i,
                                };

                                if (this.editable) {
                                    properties.balloonContent = "<button class='del" + i + "'>Удалить точку</button>"
                                }

                                let midPoint = new ymaps.GeoObject({
                                    geometry: {
                                        type: "Point",
                                        coordinates: this.getP(i),
                                    },
                                    properties: properties
                                }, {
                                    preset: 'islands#lightBlueStretchyIcon',
                                    balloonContentLayout: component.createMidPointPopupLayout(pos - 1),
                                    draggable: this.editable,
                                    zIndex: 10000000 + i,
                                    zIndexHover: 10000000 + i,
                                    zIndexActive: 10000000 + i,
                                });
                                midPoint.events.add('dragend', function () {
                                    component.setP(pos, midPoint.geometry.getCoordinates());
                                    component.refreshContour();
                                });
                                midPoints.push(midPoint);
                                this.map.geoObjects.add(midPoint);
                            }
                            this.midPoints = midPoints;
                        } else {
                            this.midPoints = null;
                        }

                        this.map.geoObjects.add(endLabel);
                        this.endLabel = endLabel;
                    }

                    this.map.geoObjects.add(label);
                    this.label = label;
                },
                createMidPointPopupLayout: function (pos) {
                    var component = this;

                    var layout = ymaps.templateLayoutFactory.createClass(
                        '<div class="item">' +
                        '<h3>{{properties.title}}</h3>' +
                        '<img src="{{properties.img}}">' +
                        '<p>{{properties.description}}</p>' +
                        '<button id="remove-placemark">Удалить</button>' +
                        '</div>', {
                            build: function () {
                                layout.superclass.build.call(this);
                                document.getElementById('remove-placemark').addEventListener('click', this.onRemove);
                            },
                            clear: function () {
                                document.getElementById('remove-placemark').removeEventListener('click', this.onRemove);
                                layout.superclass.clear.call(this);
                            },
                            onRemove: function () {
                                component.map.geoObjects.remove(component.midPoints[pos]);
                                component.spot.point.splice(pos + 1, 1);
                                component.midPoints.splice(pos, 1);
                            }
                        });
                    return layout;
                },
                refreshContour: function () {
                    if (this.contour && Array.isArray(this.spot.point[0])) {
                        for (let i = 0; i < this.spot.point.length; i++) {
                            this.contour.geometry.set(i, this.spot.point[i]);
                        }
                    }
                },
                pIdx: function (i) {
                    return i < 0 ? (this.spot.point.length + i) : i;
                },
                getP: function (i) {
                    if (Array.isArray(this.spot.point[0])) {
                        return this.spot.point[this.pIdx(i)];
                    } else if (i == 0) {
                        return this.spot.point;
                    } else {
                        return null;
                    }
                },
                setP: function (i, p) {
                    if (Array.isArray(this.spot.point[0])) {
                        this.spot.point[this.pIdx(i)] = p;
                    } else if (i == 0) {
                        this.spot.point = p;
                    } else {
                        throw "Geometry is point, not linestring"
                    }
                },
                pCenter: function () {
                    if (!Array.isArray(this.spot.point[0])) {
                        return this.spot.point;
                    }
                    let x = 0;
                    let y = 0;
                    let len = this.spot.point.length;
                    for (let i = 0; i < len; i++) {
                        x += this.spot.point[i][0];
                        y += this.spot.point[i][1];
                    }
                    return [x / len, y / len]
                },
                pBounds: function () {
                    if (!Array.isArray(this.spot.point[0])) {
                        throw "For linestring only!"
                    }
                    let margin = 0.20;
                    let result = [[100000, 100000], [-100000, -100000]];
                    for (let i = 0; i < this.spot.point.length; i++) {
                        if (this.spot.point[i][0] < result[0][0]) {
                            result[0][0] = this.spot.point[i][0];
                        }
                        if (this.spot.point[i][1] < result[0][1]) {
                            result[0][1] = this.spot.point[i][1];
                        }
                        if (this.spot.point[i][0] > result[1][0]) {
                            result[1][0] = this.spot.point[i][0];
                        }
                        if (this.spot.point[i][1] > result[1][1]) {
                            result[1][1] = this.spot.point[i][1];
                        }
                    }
                    let dx = result[1][0] - result[0][0];
                    let dy = result[1][1] - result[0][1];
                    result[0][0] = result[0][0] - margin * dx;
                    result[1][0] = result[1][0] + margin * dx;
                    result[0][1] = result[0][1] - margin * dy;
                    result[1][1] = result[1][1] + margin * dy;
                    return result
                },
                getDefaultMap: function () {
                    if (this.defaultMap) {
                        return this.defaultMap
                    }

                    let defaultMap = $.cookie("default_editor_map");
                    if (defaultMap && ymaps.mapType.storage.get(defaultMap)) {
                        return defaultMap
                    }
                    return "osm#standard"
                },
                setDefaultMap: function(type) {
                    $.cookie("default_editor_map", type, {path: '/'})
                },
                midPoints: null
            }
        }
    }

</script>