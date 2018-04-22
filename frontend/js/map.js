    ymaps.ready(init);
    var myMap;
    var trackHighlighter = {val:null};
    var pointHighlighter = {val:null};

    function loadRivers(bounds) {
        $.get(apiBase + "/visible-rivers?bbox=" + bounds.join(','), function (data) {
            $('#rivers').html('');
            var dataObj = {
                "rivers" : $.parseJSON(data),
                "apiUrl": apiBase + "/gpx"
            }
            $('#riversMenuTemplate').tmpl(dataObj).appendTo('#rivers');
        });
    }

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
            type: positionAndZoom.type
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
            setLastPositionZoomType(myMap.getCenter(), myMap.getZoom(), myMap.getType())
            loadRivers(e.get("newBounds"))
        });

        myMap.events.add('typechange', function (e) {
            setLastPositionZoomType(myMap.getCenter(), myMap.getZoom(), myMap.getType())
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

        loadRivers(myMap.getBounds())
    }
