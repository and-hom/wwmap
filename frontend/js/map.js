    ymaps.ready(init);
    var myMap;

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

    function addCachedLayer(key, name, copyright, mapId, lower_scale, upper_scale) {
        return addLayer(key, name, copyright, CACHED_TILES_TEMPLATE.replace('###', mapId), lower_scale, upper_scale)
    }

    function addLayer(key, name, copyright, tilesUrlTemplate, lower_scale, upper_scale) {
        if (typeof(lower_scale) == "undefined") {
            lower_scale = 0
        }
        if (typeof(upper_scale) == "undefined") {
            upper_scale = 18
        }
        var layer = function () {
            var layer = new ymaps.Layer(tilesUrlTemplate, {
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
        Legend = function (options) {
            Legend.superclass.constructor.call(this, options);
            this._$content = null;
            this._geocoderDeferred = null;
        };

        ymaps.util.augment(Legend, ymaps.collection.Item, {
            onAddToMap: function (map) {
                Legend.superclass.onAddToMap.call(this, map);
                this._lastCenter = null;
                this.getParent().getChildElement(this).then(this._onGetChildElement, this);
            },

            onRemoveFromMap: function (oldMap) {
                this._lastCenter = null;
                if (this._$content) {
                    this._$content.remove();
                    this._mapEventGroup.removeAll();
                }
                Legend.superclass.onRemoveFromMap.call(this, oldMap);
            },

            _onGetChildElement: function (parentDomContainer) {
                // Создаем HTML-элемент с текстом.
                var content = '<div class="legend">'
                for(i=0;i<=6;i++) {
                    content += '<div class="cat' + i + '"></div>'
                }
                content += '</div>'
                this._$content = $(content).appendTo(parentDomContainer);
            },
        });

        var helpButton = new ymaps.control.Button({
            data: {
                image: 'http://wwmap.ru/img/help.png'
            },
            options: {
                selectOnClick: false
            }
        });
        helpButton.events.add('click', function (e) {
           show_info_popup()
        });



        addCachedLayer('osm#standard', 'OSM', 'OpenStreetMap contributors, CC-BY-SA', 'osm')
        addLayer('google#satellite', 'Спутник Google', 'Изображения © DigitalGlobe,CNES / Airbus, 2018,Картографические данные © Google, 2018', GOOGLE_SAT_TILES)
        addCachedLayer('ggc#standard', 'Топографическая карта', '', 'ggc', 0, 15)
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
                'google#satellite',
                ]
            )
        );

        myMap.controls.add(new Legend(), {
            float: 'none',
            position: {
                top: 10,
                left: 10
            }});

        if ($('#info_popup').length>0) {
            myMap.controls.add(helpButton, {
                float: 'none',
                position: {
                    top: 5,
                    left: 240
                }});
        }

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
