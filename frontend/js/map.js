    ymaps.ready(init);
    var myMap;
    var trackHighlighter = {val:null};
    var pointHighlighter = {val:null};

    function init() {
        var OsmLayer = function () {
            var layer = new ymaps.Layer(STANDARD_TILES, { projection: ymaps.projection.sphericalMercator });
            //  Копирайты.
            layer.getCopyrights = function () {
                return ymaps.vow.resolve('OpenStreetMap contributors, CC-BY-SA');
            };
            layer.getZoomRange = function () {
                return ymaps.vow.resolve([0, 18]);
            };
            return layer;
        };
        ymaps.layer.storage.add('osm#standard', OsmLayer)
        ymaps.mapType.storage.add('osm#standard', new ymaps.MapType('OSM', ['osm#standard']));

        positionAndZoom = getLastPositionAndZoom()

        myMap = new ymaps.Map("map", {
            center: positionAndZoom.position,
            zoom: positionAndZoom.zoom,
            controls: ["zoomControl", "fullscreenControl"],
            type: getLastMapType()
        });

        LabelBalloonContentLayout = ymaps.templateLayoutFactory.createClass($('#bubble_template').html());

        myMap.controls.add(
            new ymaps.control.TypeSelector(
                ['osm#standard', 'yandex#satellite']
            )
        );

        myMap.events.add('click', function (e) {
            myMap.balloon.close()
        });

        myMap.events.add('boundschange', function (e) {
            setLastPositionAndZoom(myMap.getCenter(), myMap.getZoom())
        });

        myMap.events.add('typechange', function (e) {
            setLastMapType(myMap.getType())
        });

        var objectManager = new ymaps.LoadingObjectManager(apiBase + '/ymaps-tile-ww?bbox=%b', {
            clusterHasBalloon: false,
            geoObjectOpenBalloonOnClick: false,
            geoObjectBalloonContentLayout: LabelBalloonContentLayout,
            geoObjectStrokeWidth: 3,
            splitRequests: true
        });

        objectManager.objects.events.add(['click'], function (e) {
            objectManager.objects.balloon.open(e.get('objectId'));
        });

        myMap.geoObjects.add(objectManager);


        addMouseOverOutHighliterListeners('.track-list-item', '.track-geodata', 'track-list-item-selected', function (geoData) {
            return new ymaps.Polyline(
                    geoData, {},
                    {
                        strokeColor: 'ff0000ff',
                        strokeWidth: 3
                    }
            );
        }, trackHighlighter);
        addMouseOverOutHighliterListeners('.point-list-item', '.point-geodata', 'point-list-item-selected', function (geoData) {
            return new ymaps.Placemark(
                    geoData,
                    {},
                    {
                        preset: 'islands#redBookIcon',
                        zIndex: 3000
                    }
            );
        }, pointHighlighter);


        $(document).on('click', '.point-list-item', function (obj) {
            var geodataDiv = $(obj.target).find(".point-geodata")
            if (geodataDiv.length) {
                var geodataStr = geodataDiv.html();
                var geodata = $.parseJSON(geodataStr);
                myMap.setCenter(geodata);
            }
        });
    }
