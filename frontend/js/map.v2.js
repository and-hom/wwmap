function WWMapSearchProvider() {

}

WWMapSearchProvider.prototype.geocode = function (request, options) {
    var deferred = new ymaps.vow.defer(),
        geoObjects = new ymaps.GeoObjectCollection(),
        // Сколько результатов нужно пропустить.
        offset = options.skip || 0,
        // Количество возвращаемых результатов.
        limit = options.results || 20;

    var xhr = new XMLHttpRequest();
    xhr.open("POST", apiBase + "/search", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.send(request);
    xhr.onload = function(e) {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                var respData = JSON.parse(xhr.responseText);

                for (var i = 0, l = respData.spots.length; i < l; i++) {
                    var spot = respData.spots[i];

                    geoObjects.add(new ymaps.Placemark(spot.point, {
                        name: spot.title,
                        description: spot.river_title,
                        balloonContentBody: '<p>' + spot.title + ' (' + spot.river_title + ')' + '</p>',
                        boundedBy: [addToPoint(spot.point, -0.003), addToPoint(spot.point, 0.003)]
                    },{
                        iconLayout: 'default#image',
                        iconImageHref: respData.resource_base + '/img/empty-px.png',
                        iconImageSize: [1, 1],
                        iconImageOffset: [-1, -1],

                        id: spot.id
                    }));
                }

                for (var i = 0, l = respData.rivers.length; i < l; i++) {
                    var river = respData.rivers[i];

                    geoObjects.add(new ymaps.Placemark(center(river.bounds), {
                        name: river.title,
                        description: river.region.title,
                        balloonContentBody: '<p>' + river.title + '</p>',
                        boundedBy: river.bounds
                    },{
                        iconLayout: 'default#image',
                        iconImageHref: respData.resource_base + '/img/empty-px.png',
                        iconImageSize: [1, 1],
                        iconImageOffset: [-1, -1],

                        id: river.id
                    }));
                }

                deferred.resolve({
                    geoObjects: geoObjects,
                    metaData: {
                        geocoder: {
                            request: request,
                            found: geoObjects.getLength(),
                            results: limit,
                            skip: offset
                        }
                    }
                });
            } else {
                throw xhr.responseText
            }
        }
    };

    return deferred.promise();
};

function addToPoint(p, x) {
    return [p[0] + x, p[1] + x]
}

function center(bounds) {
    return [(bounds[0][0] + bounds[1][0]) / 2, (bounds[0][1] + bounds[1][1]) / 2]
}

function WWMap(divId, bubbleTemplate, riverList, tutorialPopup, catalogLinkType) {
    this.divId = divId;
    this.bubbleTemplate = bubbleTemplate;

    this.riverList = riverList;

    this.tutorialPopup = tutorialPopup;
    this.catalogLinkType = catalogLinkType;

    this.catFilter = 1;

    addCachedLayer('osm#standard', 'OSM', 'OpenStreetMap contributors, CC-BY-SA', 'osm');
    addLayer('google#satellite', 'Спутник Google', 'Изображения © DigitalGlobe,CNES / Airbus, 2018,Картографические данные © Google, 2018', GOOGLE_SAT_TILES);
    addCachedLayer('ggc#standard', 'Топографическая карта', '', 'ggc', 0, 15);
    //      addLayer('marshruty.ru#genshtab', 'Маршруты.ру', 'marshruty.ru', MARSHRUTY_RU_TILES, 8)
}

WWMap.prototype.loadRivers = function (bounds) {
    if (this.riverList) {
        var riverList = this.riverList;
        $.get(apiBase + "/visible-rivers-light?bbox=" + bounds.join(',') + "&max_cat=" + this.catFilter, function (data) {
            var dataObj = {
                "rivers": JSON.parse(data)
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

    var yMap;
    try {
        yMap = new ymaps.Map(this.divId, {
            center: positionAndZoom.position,
            zoom: positionAndZoom.zoom,
            controls: ["zoomControl", "fullscreenControl"],
            type: positionAndZoom.type
        });
    } catch (err) {
        setLastPositionZoomType(defaultPosition(), defaultZoom(), defaultMapType());
        throw err
    }
    this.yMap = yMap;


    this.yMap.controls.add(
        new ymaps.control.TypeSelector([
                'osm#standard',
                'ggc#standard',
                'yandex#satellite',
                'google#satellite'
            ]
        )
    );

    this.yMap.controls.add(new ymaps.control.SearchControl({
        options: {
            provider: new WWMapSearchProvider(),
            placeholderContent: 'Река или порог'
        }
    }));

    if (this.tutorialPopup) {
        this.yMap.controls.add(this.createHelpBtn(), {
            float: 'left'
        });
    }

    this.yMap.controls.add(createLegend(this), {
        float: 'left'
    });

    this.yMap.controls.add('rulerControl', {
        scaleLine: true
    });

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
        splitRequests: false
    });

    objectManager.setFilter(function(obj) {
        if (obj.properties.category) {
            var objCategory = parseInt(obj.properties.category[0]);
            return t.catFilter === 1 || objCategory < 0 || objCategory >= t.catFilter;
        }
        return true
    });

    objectManager.objects.events.add(['click'], function (e) {
        objectManager.objects.balloon.open(e.get('objectId'));
    });

    this.yMap.geoObjects.add(objectManager);
    this.objectManager = objectManager;

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

function createLegend(wwmap) {
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
                content += '<div class="cat' + i + ' cat-bold"></div>'
            }
            content += '</div>';
            this._$content = $(content).appendTo(parentDomContainer);

            var legendDiv = $('.wwmap-legend');
            var t = this;
            legendDiv.bind('click', function (e) {
                t.onFilterStateChanged(e)
            });
            legendDiv.bind('mousedown', function (e) {
                t.onDragStart(e)
            });
            legendDiv.bind('mouseup', function (e) {
                t.onDragStop(e)
            });
            legendDiv.bind('mousemove', function (e) {
                t.onDrag(e)
            });
        },

        onDragStart: function (e) {
            this.drag = true
        },
        onDragStop: function (e) {
            this.drag = false
        },
        onDrag: function (e) {
            if (this.drag) {
                this.onFilterStateChanged(e)
            }
        },
        onFilterStateChanged: function (e) {
            var category = $(e.target)
                .attr("class")
                .split(' ')
                .filter(function (c) {
                    return c.startsWith('cat')
                })
                .map(function (value) {
                    return parseInt(value.substring(3))
                })[0];
            if (!category || category === wwmap.catFilter) {
                return
            }
            for (var i = 1; i <= 6; i++) {
                if (i < category) {
                    $('.wwmap-legend .cat' + i).removeClass("cat-bold")
                } else {
                    $('.wwmap-legend .cat' + i).addClass("cat-bold")
                }
            }
            wwmap.catFilter = category;
            wwmap.objectManager.reloadData();
            wwmap.loadRivers(wwmap.yMap.getBounds())
        }
    });

    return new Legend()
}

function WWMapPopup(templateDivId, fromTemplates, divId, options) {
    this.divId = divId;
    if (fromTemplates) {
        loadFragment(MAP_FRAGMENTS_URL, templateDivId, function(templateText) {
            $('body').prepend(templateText);
            t.templateDiv = $('#' + templateDivId);
        })
    } else {
        t.templateDiv = $('#' + templateDivId);
    }

    this.submitUrl = (options) ? options.submitUrl : null;
    this.okMsg = (options) ? options.okMsg : null;
    this.failMsg = (options) ? options.failMsg : null;

    $('body').prepend('<div id="' + this.divId + '" class="wwmap-popup_area"></div>');
    this.div = $("#" + this.divId);

    var t = this;

    // close on mouse click outside the window
    if (!options || options.closeOnMouseClickOutside!==false) {
        this.div.click(function (source) {
            var classAttr = $(source.target).attr('class');
            if (classAttr && classAttr.indexOf('wwmap-popup_are') > -1) {
                t.hide()
            }
        });
    }

    // close on escape pressed
    if (!options || options.closeOnEscape!==false) {
        $(document).keyup(function (e) {
            if (e.key === "Escape") {
                t.hide()
            }
        });
    }
}

WWMapPopup.prototype.show = function (dataObject) {
    var t = this;

    var html = "";
    if (dataObject) {
        html = this.templateDiv.tmpl(dataObject)[0].outerHTML
    } else {
        html = this.templateDiv.html()
    }

    this.div.html(html);
    $("#" + this.divId + " input[name=cancel]").click(function() {return t.hide()});
    $("#" + this.divId + " input[type=submit]").click(function() {return t.submit_form()});

    initMailtoLinks();
    this.div.addClass("wwmap-show");
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


function loadFragment(url, fromId, onLoad, data) {
    var virtualElement = $('<div id="loaded-content"></div>');
    virtualElement.load(url + ' #' + fromId, function () {
        if (data) {
            onLoad(virtualElement.tmpl(data).html())
        } else {
            onLoad(virtualElement.html())
        }
    });
}

function extractInnerHtml(str) {
    return $(str).html()
}

function show_report_popup(id, title, riverTitle) {
    var dataObject = {
        object_id: id,
        object_title: title,
        title: riverTitle
    };
    var info = getWwmapUserInfoForMapControls();
    if (info && info.login) {
        dataObject.user = info.login;
    }
    reportPopup.show(dataObject)
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

    this.riverInfoPopup = new WWMapPopup('river_desc_template', true, "river_desc");
}

RiverList.prototype.update = function (rivers) {
    if (this.templateDiv) {
        rivers.canEdit = canEdit();
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

var wwMap;

function show_map_at(bounds) {
    wwMap.setBounds(bounds, {
        checkZoomRange: true,
        duration: 200
    })
}

function initWWMap(mapId, riversListId, catalogLinkType) {
    if (catalogLinkType && CATALOG_LINK_TYPES.indexOf(catalogLinkType) <= -1) {
        throw "Unknown catalog link type. Available are: " + CATALOG_LINK_TYPES
    }

    // initialize popup windows
    reportPopup = new WWMapPopup('report_popup_template', true, 'report_popup', {
            submitUrl : apiBase + "/report",
            okMsg:  "Запрос отправлен. Я прочитаю его по мере наличия свободного времени",
            failMsg: "Что-то пошло не так...",
            // To prevent contents lost
            closeOnEscape: false,
            closeOnMouseClickOutside: false
        });
    var tutorialPopup = new WWMapPopup('info_popup_template', true, 'info_popup');

    // riverList
    if (riversListId) {
        riverList = new RiverList(riversListId, 'rivers_template', true)
    }

    // init and show map
    ymaps.ready(function () {
        loadFragment(MAP_FRAGMENTS_URL, 'bubble_template', function (bubbleContent) {
            wwMap = new WWMap(mapId, extractInnerHtml(bubbleContent), riverList, tutorialPopup, catalogLinkType);
            ymaps.modules.require(['overlay.BiPlacemark'], function (BiPlacemarkOverlay) {
                ymaps.overlay.storage.add("BiPlacemrakOverlay", BiPlacemarkOverlay);
                wwMap.init()
            });
        })
    });
}

function getWwmapUserInfoForMapControls() {
    if (typeof getWwmapUserInfo == 'function') {
        return getWwmapUserInfo();
    }
    return null;
}

function canEdit() {
    var info = getWwmapUserInfoForMapControls();
    return (info && info.roles && ['EDITOR', 'ADMIN'].filter(function (r) {
        return info.roles.includes(r)
    }).length > 0)
}

function show_river_info_popup(id) {
    $.get(apiBase + "/river-card/" + id, function (data) {
        var dataObj = JSON.parse(data);
        dataObj.canEdit = canEdit();
        dataObj.apiUrl = apiBase + "/gpx/river";
        dataObj.apiBase = apiBase;
        riverList.riverInfoPopup.show(dataObj);
    });
    return false;
}
