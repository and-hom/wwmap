    const LAST_POS_COOKIE_NAME = "last-map-pos";
    const apiBase = "http://localhost:7007";

    const STANDARD_TILES = 'http://tile.openstreetmap.org/%z/%x/%y.png';
    const THUNDERFOREST_OUTDOOR_TILES = 'http://a.tile.thunderforest.com/outdoors/%z/%x/%y.png';
    const THUNDERFOREST_LANDSCAPE_TILES = 'http://a.tile.thunderforest.com/landscape/%z/%x/%y.png';

    ymaps.ready(init);
    var myMap;
    var trackHighlighter = {val:null};
    var pointHighlighter = {val:null};

    function init() {
        myMap = new ymaps.Map("map", {
            center: lastPosition(),
            zoom: 7,
            controls: ["zoomControl", "fullscreenControl"],
            type: 'yandex#satellite'
        });

        myMap.layers.add(new ymaps.Layer(
        STANDARD_TILES, {
            projection: ymaps.projection.sphericalMercator
        }));
        myMap.copyrights.add(' OpenStreetMap contributors, CC-BY-SA');

        myMap.events.add('click', function (e) {
            myMap.balloon.close()
        });

        var objectManager = new ymaps.LoadingObjectManager(apiBase + '/ymaps-tile-wr?bbox=%b', {
            clusterHasBalloon: false,
            geoObjectOpenBalloonOnClick: false,
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

    function lastPosition() {
        lastPos = $.cookie(LAST_POS_COOKIE_NAME);
        if (lastPos) {
            return $.parseJSON(lastPos)
        } else {
            return [55.76, 37.64]
        }
    }