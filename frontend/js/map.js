    ymaps.ready(init);
    var myMap;
    var trackHighlighter = {val:null};
    var pointHighlighter = {val:null};

    function init() {
        myMap = new ymaps.Map("map", {
            center: getLastPosition(),
            zoom: getLastZoom(),
            controls: ["zoomControl", "fullscreenControl"],
            type: 'yandex#satellite'
        });
        // Создание вложенного макета содержимого балуна.
            LabelBalloonContentLayout = ymaps.templateLayoutFactory.createClass(
                '<h3 class="popover-title">'+
                '[if properties.link]<a target="_blank" href="$[properties.link]">$[properties.title]</a>[else]$[properties.title][endif]</h3>'+
                    '<div class="popover-content">' +
                    '<div>Категория сложности: [if properties.category=="0"]не определена[else]$[properties.category][endif]</div>' +
                    '<div>$[properties.short_desc]</div>' +
                    '<a id="report_$[properties.link]" href="button">Сообщить о неточном местоположении<a/>' +
                    '</div>'
            );

        myMap.layers.add(new ymaps.Layer(
        STANDARD_TILES, {
            projection: ymaps.projection.sphericalMercator
        }));
        myMap.copyrights.add(' OpenStreetMap contributors, CC-BY-SA');

        myMap.events.add('click', function (e) {
            myMap.balloon.close()
        });

        myMap.events.add('boundschange', function (e) {
            setLastPosition(myMap.getCenter())
            setLastZoom(myMap.getZoom())
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