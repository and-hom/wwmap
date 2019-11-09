export function createLegend() {
    let Legend = function (options) {
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
            for (let i = 0; i <= 6; i++) {
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
    return new Legend();
}