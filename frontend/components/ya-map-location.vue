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
            editEndPoint: {
                type: Boolean,
                default: false
            },
        },
        watch: {
            // This would be called anytime the value of title changes
            editEndPoint(newValue, oldValue) {
                this.doUpdate();
            }
        },
        updated: function () {
            this.doUpdate()
        },
        created: function () {
            var component = this;
            ymaps.ready(function () {
                if (component.map) {
                    component.map.destroy();
                    component.label.geometry.setCoordinates(component.spot.point);
                } else {
                    addCachedLayer('osm#standard', 'OSM', 'OpenStreetMap contributors, CC-BY-SA', 'osm');
                    addLayer('google#satellite', 'Спутник Google', 'Изображения © DigitalGlobe,CNES / Airbus, 2018,Картографические данные © Google, 2018', GOOGLE_SAT_TILES);
                    addCachedLayer('ggc#standard', 'Топографическая карта', '', 'ggc', 0, 15)
                }

                var myMap = new ymaps.Map("map", {
                    center: component.spot.point,
                    zoom: component.zoom,
                    controls: ["zoomControl"],
                    type: "osm#standard"
                });
                myMap.controls.add(
                    new ymaps.control.TypeSelector([
                        'osm#standard',
                        'ggc#standard',
                        'yandex#satellite',
                        'google#satellite'
                    ])
                );
                if (component.yaSearch) {
                    myMap.controls.add(new ymaps.control.SearchControl({
                        options: {
                            float: 'left',
                            floatIndex: 100
                        }
                    }));
                }

                component.map = myMap;


                component.addObjectManager();
                component.addLabel();


            })
        },
        data: function () {
            return {
                mapDivStyle: function () {
                    return 'width: ' + this.width + '; height: ' + this.height + ';'
                },
                doUpdate: function () {
                    if (this.map) {
                        this.map.setCenter(this.spot.point);

                        this.map.geoObjects.remove(this.label);
                        if (this.endLabel) {
                            this.map.geoObjects.remove(this.endLabel);
                            this.map.geoObjects.remove(this.contour);
                        }
                        this.addLabel();

                        this.map.geoObjects.remove(this.objectManager);
                        this.addObjectManager();
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
                        splitRequests: true
                    });
                    this.map.geoObjects.add(objectManager);
                    this.objectManager = objectManager
                },
                addLabel: function () {
                    var component = this;
                    var props = {
                        hintContent: this.spot.title,
                    };
                    if (this.editEndPoint) {
                        props.iconContent = "Начало"
                    }
                    var label = new ymaps.GeoObject({
                        geometry: {
                            type: "Point",
                            coordinates: this.spot.point,
                        },
                        properties: props
                    }, {
                        preset: 'islands#blueStretchyIcon',
                        draggable: this.editable,
                        zIndex: 10000000,
                    });
                    label.events.add('dragend', function (e) {
                        component.spot.point = label.geometry.getCoordinates();
                        component.refreshContour();
                    });

                    if (this.editable) {
                        this.map.events.add('click', function (e) {
                            p = e.get('coords');
                            label.geometry.setCoordinates(p);
                            component.spot.point = p;
                            component.refreshContour();
                        });
                    }

                    if (this.editEndPoint) {
                        ymaps.modules.require(['overlay.BiPlacemark'], function (BiPlacemarkOverlay) {
                            var contour = new ymaps.GeoObject({
                                geometry: {
                                    type: "LineString",
                                    coordinates: [component.spot.point, component.spot.props.end_point],
                                },
                                properties: {}
                            }, {
                                lineStringOverlay: BiPlacemarkOverlay,
                                strokeColor: "#BBBBBBCC",
                                strokeStyle: 'shortdash',
                                strokeWidth: 3,
                                fill: false
                            });
                            component.map.geoObjects.add(contour);
                            component.contour = contour;
                        });

                        var endLabel = new ymaps.GeoObject({
                            geometry: {
                                type: "Point",
                                coordinates: this.spot.props.end_point,
                            },
                            properties: {
                                hintContent: this.spot.title + " конец",
                                iconContent: "Конец"
                            }
                        }, {
                            preset: 'islands#yellowStretchyIcon',
                            draggable: this.editable,
                            zIndex: 10000001,
                        });
                        endLabel.events.add('dragend', function (e) {
                            component.spot.props.end_point = endLabel.geometry.getCoordinates();
                            component.refreshContour();
                        });

                        this.map.geoObjects.add(endLabel);
                        this.endLabel = endLabel;
                    }

                    this.map.geoObjects.add(label);
                    this.label = label;
                },
                refreshContour: function () {
                    if (this.contour) {
                        this.contour.geometry.set(0, this.spot.point);
                        this.contour.geometry.set(1, this.spot.props.end_point);
                    }
                }
            }
        }
    }

</script>