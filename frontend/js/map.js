    ymaps.ready(init);
    var myMap;
    var trackHighlighter = {val:null};
    var pointHighlighter = {val:null};

    function loadRivers(bounds) {
        if ($('#riversMenuTemplate').length==0 || $("#rivers").length==0) {
            // can not find template or container
            return
        }
        $.get(apiBase + "/visible-rivers?bbox=" + bounds.join(','), function (data) {
            $('#rivers').html('');
            var dataObj = {
                "rivers" : JSON.parse(data),
                "apiUrl": apiBase + "/gpx"
            }
            for (i in dataObj.rivers) {
                if(dataObj.rivers[i].bounds) {
                    dataObj.rivers[i].bounds =  JSON.stringify(dataObj.rivers[i].bounds)
                }
            }
            $('#riversMenuTemplate').tmpl(dataObj).appendTo('#rivers');
        });
    }

    function addLayer(key, name, copyright, tiles, lower_scale, upper_scale) {
        if (typeof(lower_scale) == "undefined") {
            lower_scale = 0
        }
        if (typeof(upper_scale) == "undefined") {
            upper_scale = 18
        }
        var layer = function () {
            var layer = new ymaps.Layer(tiles, {
                projection: ymaps.projection.sphericalMercator,
            });
            //  Копирайты.
            layer.getCopyrights = function () {
                return ymaps.vow.resolve(copyright);
            };
            layer.getZoomRange = function () {
                return ymaps.vow.resolve([lower_scale, upper_scale]);
            };
            return layer;
        };
        ymaps.layer.storage.add(key, layer)
        ymaps.mapType.storage.add(key, new ymaps.MapType(name, [key]));
    }

    function init() {
        addLayer('osm#standard', 'OSM', 'OpenStreetMap contributors, CC-BY-SA', OSM_TILES)
        addLayer('google#satellite', 'Спутник Google', 'Изображения © DigitalGlobe,CNES / Airbus, 2018,Картографические данные © Google, 2018', GOOGLE_SAT_TILES)
        addLayer('ggc#standard', 'ГГц', '', GGC_TILES, 0, 15)
//        addLayer('marshruty.ru#genshtab', 'Маршруты.ру', 'marshruty.ru', MARSHRUTY_RU_TILES, 8)

        positionAndZoom = getLastPositionAndZoom()

        myMap = new ymaps.Map("map", {
            center: positionAndZoom.position,
            zoom: positionAndZoom.zoom,
            controls: ["zoomControl", "fullscreenControl"],
            type: positionAndZoom.type
        });

        LabelBalloonContentLayout = ymaps.templateLayoutFactory.createClass($('#bubble_template').html());

        myMap.controls.add(
            new ymaps.control.TypeSelector([
                'osm#standard',
                'ggc#standard',
                'yandex#satellite',
                'google#satellite'
                ]
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

        var objectManager = new ymaps.RemoteObjectManager(apiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z', {
            clusterHasBalloon: false,
            geoObjectOpenBalloonOnClick: false,
            geoObjectBalloonContentLayout: LabelBalloonContentLayout,
            geoObjectStrokeWidth: 3,
            splitRequests: true,

            clusterHasBalloon: false,
        });

        objectManager.objects.events.add(['click'], function (e) {
            objectManager.objects.balloon.open(e.get('objectId'));
        });

        myMap.geoObjects.add(objectManager);

        loadRivers(myMap.getBounds())
    }
