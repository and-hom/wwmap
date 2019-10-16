import fileDownload from "js-file-download"

export function createMeasurementToolControl(measurementTool) {
    let MeasurementControl = function (options) {
        MeasurementControl.superclass.constructor.call(this, options);
        this._$content = null;
        this._geocoderDeferred = null;
    };

    ymaps.util.augment(MeasurementControl, ymaps.collection.Item, {
        onAddToMap: function (map) {
            MeasurementControl.superclass.onAddToMap.call(this, map);
            this._lastCenter = null;
            this.getParent().getChildElement(this).then(this._onGetChildElement, this);
        },

        onRemoveFromMap: function (oldMap) {
            this._lastCenter = null;
            if (this._$content) {
                this._$content.remove();
                this._mapEventGroup.removeAll();
            }
            MeasurementControl.superclass.onRemoveFromMap.call(this, oldMap);
        },

        _onGetChildElement: function (parentDomContainer) {
            // Создаем HTML-элемент с текстом.
            var content = '<div class="wwmap-route-control">' +
                '<button class="ymaps-2-1-73-float-button-text, wwmap-measure-btn">Расстояния по воде<img style="height:24px"/></button>' +
                '<button class="ymaps-2-1-73-float-button-text, wwmap-measure-download-btn" style="display: none;" title="Скачать GPX"><img style="height:24px" src="img/download.png"/></button>' +
                '<button class="ymaps-2-1-73-float-button-text, wwmap-measure-revert-btn" style="display: none;" title="Удалить последнюю точку (Esc)"><img style="height:24px" src="img/revert.png"/></button>' +
                '<button class="ymaps-2-1-73-float-button-text, wwmap-measure-delete-btn" style="display: none;" title="Очистить трек"><img style="height:24px" src="img/del.png"/></button>' +
                '</div>';
            this._$content = $(content).appendTo(parentDomContainer);

            var measureOnOffBtn = $('.wwmap-measure-btn');
            var measureDownloadBtn = $('.wwmap-measure-download-btn');
            var measureRevertBtn = $('.wwmap-measure-revert-btn');
            var measureDeleteBtn = $('.wwmap-measure-delete-btn');
            var t = this;
            measureOnOffBtn.bind('click', function (e) {
                if (measurementTool.enabled) {
                    measureOnOffBtn.removeClass("wwmap-measure-btn-pressed");
                    measurementTool.disable();
                    measureDownloadBtn.css('display', 'none');
                    measureRevertBtn.css('display', 'none');
                    measureDeleteBtn.css('display', 'none');
                } else {
                    measureOnOffBtn.addClass("wwmap-measure-btn-pressed");
                    measurementTool.enable();
                    measureDownloadBtn.css('display', 'inline-block');
                    measureRevertBtn.css('display', 'inline-block');
                    measureDeleteBtn.css('display', 'inline-block');
                }
            });

            measureDownloadBtn.bind('click', function (e) {
                if(measurementTool.multiPath.segmentCount()>0) {
                    fileDownload(measurementTool.multiPath.createGpx(), "track.gpx", "application/gpx+xml");
                } else {
                    alert("Добавьте линию")
                }
            });

            measureRevertBtn.bind('click', function (e) {
                measurementTool.multiPath.removeLastSegments(1);
            });

            measureDeleteBtn.bind('click', function (e) {
                measurementTool.reset();
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
    });

    return new MeasurementControl()
}
