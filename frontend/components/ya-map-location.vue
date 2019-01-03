<template>
    <div id="map" :style="mapDivStyle()"></div>
</template>

<script>
    module.exports = {
        props: {
            spot: Object,
            width: {
                type: String,
                default: "600px",
            },
            height: {
                type: String,
                default: "400px",
            },
            editable: {
                type: Boolean,
                default: false,
            },
        },
        updated: function() {
            this.doUpdate()
        },
        created: function() {
            var component = this
            ymaps.ready(function() {
                if (component.map) {
                    component.map.destroy()
                    component.label.geometry.setCoordinates(component.spot.point);
                } else {
                    addCachedLayer('osm#standard', 'OSM', 'OpenStreetMap contributors, CC-BY-SA', 'osm')
                    addLayer('google#satellite', 'Спутник Google', 'Изображения © DigitalGlobe,CNES / Airbus, 2018,Картографические данные © Google, 2018', GOOGLE_SAT_TILES)
                    addCachedLayer('ggc#standard', 'Топографическая карта', '', 'ggc', 0, 15)
                }

                var myMap = new ymaps.Map("map", {
                    center: component.spot.point,
                    zoom: 12,
                    controls: ["zoomControl"],
                    type: "osm#standard",
                });
                myMap.controls.add(
                    new ymaps.control.TypeSelector([
                        'osm#standard',
                        'ggc#standard',
                        'yandex#satellite',
                        'google#satellite',
                    ])
                );
                component.map = myMap


                component.addObjectManager()
                component.addLabel()


            })
        },
        data: function() {
            return {
                mapDivStyle: function() {
                    return 'width: ' + this.width + '; height: ' + this.height + ';'
                },
                doUpdate: function() {
                    if (this.map) {
                        this.map.setCenter(this.spot.point)

                        this.map.geoObjects.remove(this.label)
                        this.addLabel()

                        this.map.geoObjects.remove(this.objectManager)
                        this.addObjectManager()
                    }
                },
                objectManagerUrlTemplate: function() {
                    var skip = 0
                    if (this.spot && this.spot.id) {
                        skip = this.spot.id
                    }
                    url = backendApiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z&skip=' + skip
                    st = getWwmapSessionId();
                    if (st) {
                        url += '&session_id=' + st
                    }
                    return url
                },
                addObjectManager:function() {
                    var objectManager = new ymaps.RemoteObjectManager(this.objectManagerUrlTemplate(), {
                        clusterHasBalloon: false,
                        geoObjectOpenBalloonOnClick: false,
                        geoObjectStrokeWidth: 3,
                        splitRequests: true,
                        clusterHasBalloon: false,
                    });
                    this.map.geoObjects.add(objectManager);
                    this.objectManager = objectManager
                },
                addLabel: function() {
                    var component = this
                    var label = new ymaps.GeoObject({
                        geometry: {
                            type: "Point",
                            coordinates: this.spot.point
                        },
                        properties: {
                            hintContent: this.spot.title,
                        }
                    }, {
                        preset: 'islands#blueIcon',
                        draggable: this.editable,
                    })
                    label.events.add('dragend', function (e) {
                        component.spot.point = label.geometry.getCoordinates()
                    });

                    if (this.editable) {
                        this.map.events.add('click', function (e) {
                            p = e.get('coords')
                            label.geometry.setCoordinates(p)
                            component.spot.point = p
                        });
                    }

                    this.map.geoObjects.add(label)
                    this.label = label
                },
            }
        },
    }

</script>