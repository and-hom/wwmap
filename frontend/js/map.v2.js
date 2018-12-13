function WWMap(divId, bubbleTemplate, riverList, tutorialPopup, catalogLinkType) {
    this.divId = divId;
    this.bubbleTemplate = bubbleTemplate;

    this.riverList = riverList;

    this.tutorialPopup = tutorialPopup;
    this.catalogLinkType = catalogLinkType;

    addCachedLayer('osm#standard', 'OSM', 'OpenStreetMap contributors, CC-BY-SA', 'osm');
    addLayer('google#satellite', 'Спутник Google', 'Изображения © DigitalGlobe,CNES / Airbus, 2018,Картографические данные © Google, 2018', GOOGLE_SAT_TILES);
    addCachedLayer('ggc#standard', 'Топографическая карта', '', 'ggc', 0, 15);
    //      addLayer('marshruty.ru#genshtab', 'Маршруты.ру', 'marshruty.ru', MARSHRUTY_RU_TILES, 8)
}

WWMap.prototype.loadRivers = function (bounds) {
    if (this.riverList) {
        var riverList = this.riverList;
        $.get(apiBase + "/visible-rivers?bbox=" + bounds.join(','), function (data) {
            var dataObj = {
                "rivers": JSON.parse(data),
                "apiUrl": apiBase + "/gpx/river",
                "apiBase": apiBase
            };
            for (i in dataObj.rivers) {
                if (dataObj.rivers[i].bounds) {
                    dataObj.rivers[i].bounds = JSON.stringify(dataObj.rivers[i].bounds)
                }
            }
            riverList.update(dataObj)
        });
    }
};

WWMap.prototype.createHelpBtn = function () {
    var helpButton = new ymaps.control.Button({
        data: {
            image: 'http://wwmap.ru/img/help.png'
        },
        options: {
            selectOnClick: false
        }
    });
    var t = this;
    helpButton.events.add('click', function (e) {
        t.tutorialPopup.show()
    });
    return helpButton
};

WWMap.prototype.init = function () {
    var positionAndZoom = getLastPositionAndZoom();

    this.yMap = new ymaps.Map(this.divId, {
        center: positionAndZoom.position,
        zoom: positionAndZoom.zoom,
        controls: ["zoomControl", "fullscreenControl"],
        type: positionAndZoom.type
    });


    this.yMap.controls.add(
        new ymaps.control.TypeSelector([
                'osm#standard',
                'ggc#standard',
                'yandex#satellite',
                'google#satellite'
            ]
        )
    );

    var LegendClass = createLegendClass();
    this.yMap.controls.add(new LegendClass(), {
        float: 'none',
        position: {
            top: 10,
            left: 10
        }
    });

    if (this.tutorialPopup) {
        this.yMap.controls.add(this.createHelpBtn(), {
            float: 'none',
            position: {
                top: 5,
                left: 240
            }
        });
    }

    var t = this;
    this.yMap.events.add('click', function (e) {
        t.yMap.balloon.close()
    });

    this.yMap.events.add('boundschange', function (e) {
        setLastPositionZoomType(t.yMap.getCenter(), t.yMap.getZoom(), t.yMap.getType());
        t.loadRivers(e.get("newBounds"))
    });

    this.yMap.events.add('typechange', function (e) {
        setLastPositionZoomType(t.yMap.getCenter(), t.yMap.getZoom(), t.yMap.getType())
    });

    var objectManager = new ymaps.RemoteObjectManager(apiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z&link_type=' + this.catalogLinkType, {
        clusterHasBalloon: false,
        geoObjectOpenBalloonOnClick: false,
        geoObjectBalloonContentLayout: ymaps.templateLayoutFactory.createClass(this.bubbleTemplate),
        geoObjectStrokeWidth: 3,
        splitRequests: true,

        clusterHasBalloon: false
    });

    objectManager.objects.events.add(['click'], function (e) {
        objectManager.objects.balloon.open(e.get('objectId'));
    });

    this.yMap.geoObjects.add(objectManager);

    this.loadRivers(this.yMap.getBounds())
};

WWMap.prototype.setBounds = function (bounds, opts) {
    this.yMap.setBounds(bounds, opts)
};

function addCachedLayer(key, name, copyright, mapId, lower_scale, upper_scale) {
    return addLayer(key, name, copyright, CACHED_TILES_TEMPLATE.replace('###', mapId), lower_scale, upper_scale)
}

function addLayer(key, name, copyright, tilesUrlTemplate, lower_scale, upper_scale) {
    if (typeof (lower_scale) == "undefined") {
        lower_scale = 0
    }
    if (typeof (upper_scale) == "undefined") {
        upper_scale = 18
    }
    var layer = function () {
        var layer = new ymaps.Layer(tilesUrlTemplate, {
            projection: ymaps.projection.sphericalMercator
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
    ymaps.layer.storage.add(key, layer);
    ymaps.mapType.storage.add(key, new ymaps.MapType(name, [key]));
}

function createLegendClass() {
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
            var content = '<div class="wwmap-legend">';
            for (i = 0; i <= 6; i++) {
                content += '<div class="cat' + i + '"></div>'
            }
            content += '</div>';
            this._$content = $(content).appendTo(parentDomContainer);
        }
    });
    return Legend
}

function WWMapPopup(templateUrl, divId, submitUrl, okMsg, failMsg) {
    this.divId = divId;
    this.templateUrl = templateUrl;
    this.submitUrl = submitUrl;
    this.okMsg = okMsg;
    this.failMsg = failMsg;

    $('body').prepend('<div id="' + this.divId + '" class="wwmap-popup_area"></div>');
    this.div = $("#" + this.divId)
}

WWMapPopup.prototype.show = function (afterShowF) {
    var t = this;
    loadFragment(this.templateUrl, this.divId, function (templateText) {
        t.div.html(templateText);
        $("#" + t.divId + " input[name=cancel]").click(function() {return t.hide()});
        $("#" + t.divId + " input[type=submit]").click(function() {return t.submit_form()});
        if (afterShowF) {
            afterShowF()
        }
        initMailtoLinks();
        t.div.addClass("wwmap-show");
    })
};

WWMapPopup.prototype.hide = function () {
    this.div.html('');
    this.div.removeClass("wwmap-show");
};

WWMapPopup.prototype.submit_form = function () {
    var form = $("#" + this.divId + " form");
    var t = this;
    $.post(this.submitUrl, form.serialize())
        .done(function (data) {
            window.alert(t.okMsg);
            t.hide();
            form.trigger('reset')
        }).fail(function () {
             window.alert(t.failMsg);
        });
    return false;
};


function loadFragment(url, fromId, onLoad) {
    var virtualElement = $('<div id="loaded-content"></div>');
    virtualElement.load(url + ' #' + fromId, function () {
        onLoad(virtualElement.html())
    });
}

function extractInnerHtml(str) {
    return $(str).html()
}


function show_map_at(bounds) {
    wwMap.setBounds(bounds, {
        checkZoomRange: true,
        duration: 200
    })
}

function show_report_popup(id, title, riverTitle) {
    reportPopup.show(function () {
        $("#report_popup #object_id").val(id);
        $("#report_popup #object_title").val(title);
        $("#report_popup #title").val(riverTitle)
    })
}

function RiverList(divId, templateDivId, fromTemplates) {
    this.divId = divId;
    var t = this;

    if (fromTemplates) {
        loadFragment(MAP_FRAGMENTS_URL, templateDivId, function (templateText) {
            $('body').prepend(templateText);
            t.templateDiv = $('#' + templateDivId)
        })
    } else {
        t.templateDiv = $('#' + templateDivId)
    }
}

RiverList.prototype.update = function (rivers) {
    if (this.templateDiv) {
        var html = this.templateDiv.tmpl(rivers).html();
        $('#' + this.divId).html(html)
    }
};

function initMailtoLinks() {
    // initialize all mailto links: robots do not perform js, so this links will not be detected by robots
    user = 'info';
    domain = 'wwmap.ru';
    var emailLink = $('.email-link');
    emailLink.attr('href', 'mailto:' + user + '@' + domain);
    emailLink.text(user + '@' + domain)
}

CATALOG_LINK_TYPES = [
    'none', // do not use spot link from bubble
    'from_spot',  // use link from spot properties
    'wwmap', // use link to wwmap.ru catalog
    'huskytm' // use link to huskytm.ru catalog (upload from wwmap.ru)
];

function initWWMap(mapId, riversListId, catalogLinkType) {
    if (catalogLinkType && CATALOG_LINK_TYPES.indexOf(catalogLinkType) <= -1) {
        throw "Unknown catalog link type. Available are: " + CATALOG_LINK_TYPES
    }

    // initialize popup windows
    reportPopup = new WWMapPopup(MAP_FRAGMENTS_URL, 'report_popup', apiBase + "/report",
        "Запрос отправлен. Я прочитаю его по мере наличия свободного времени", "Что-то пошло не так...");
    var tutorialPopup = new WWMapPopup(MAP_FRAGMENTS_URL, 'info_popup');

    // riverList
    var riverList = null;
    if (riversListId) {
        riverList = new RiverList(riversListId, 'rivers_template', true)
    }

    // init and show map
    ymaps.ready(function () {
        loadFragment(MAP_FRAGMENTS_URL, 'bubble_template', function (bubbleContent) {
            var wwMap = new WWMap(mapId, extractInnerHtml(bubbleContent), riverList, tutorialPopup, catalogLinkType);
            wwMap.init()
        })
    });
}