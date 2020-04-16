<template>
    <page link="timeline.htm">
        <ask id="orphan-info" title="Потерян контроль над задачей" :noBtn="false"
             msg="Если сервис был перезапущен, задачи продолжают выполняться, но отследить их, а также смотреть их логи уже нельзя"></ask>

        <div style="margin-left:10px; margin-top: 10px; ">
            <h2>Таймлайн</h2>
            <div style="display: flex;">
                <div style="display:inline-block">
                    <GChart
                            :settings="{packages: ['timeline']}"
                            :data="timeline"
                            :options="chartOptions"
                            :createChart="(el, google) => new google.visualization.Timeline(el)"
                            @ready="onChartReady"
                    />
                </div>
                <div style="display:inline-block">
                    <ul class="legend">
                        <li><svg><rect :fill="COLOR_RUNNING"/></svg>Выполняется</li>
                        <li><svg><rect :fill="COLOR_DONE"/></svg>Выполнено</li>
                        <li><svg><rect :fill="COLOR_FAIL"/></svg>Ошибка</li>
                        <li><svg><rect :fill="COLOR_ORPHAN"/></svg>Потерян контроль<a href="#" data-toggle="modal" data-target="#orphan-info"><img
                                src="img/question_16.png"></a></li>
                    </ul>
                </div>
            </div>
        </div>
    </page>
</template>

<style>
    .legend {
        list-style-type: none;
    }

    .legend li {
        margin: 7px;
        width: auto;
        height: auto;
    }

    .legend li svg {
        width: 12px;
        height: 12px;
        margin-right: 5px;
    }

    .legend li svg rect {
        width: 10px;
        height: 10px;
    }

    .google-visualization-tooltip {
        z-index: 1000;
    }
</style>


<script>
    import {doGetJson} from '../api'
    import {cronApiBase} from '../config'

    const moment = require('moment');

    function mkF(t) {
        return function () {
            if (t.chart) {
                let selection = t.chart.getSelection();
                if (selection.length > 0) {
                    let rowIdx = selection[0].row + 1;
                    let html = t.timeline[rowIdx][3];
                }
            }
        }
    }

    export default {
        created: function () {

        },
        mounted: function () {
            this.refresh();
            let t = this;
            this.timerID = setInterval(function() {
                t.refresh();
            }, 60 * 1000);

        },
        beforeDestroy: function () {
            if (this.timerID) {
                clearInterval(this.timerID)
            }
        },
        updated: function () {
        },
        data() {
            return {
                chart: null,
                timeline: [],
                chartOptions: {
                    width: 1200,
                    height: 800,
                    timeline: {
                        showBarLabels: false,
                    },
                    tooltip: {
                        // trigger: 'selection',
                        isHtml: true,
                    },
                },
                COLOR_NEW: "#fffa96",
                COLOR_RUNNING: "#9698ff",
                COLOR_DONE: "#57ff00",
                COLOR_FAIL: "#ff8789",
                COLOR_ORPHAN: "#afafaf",

                chartEvents: {
                    'select': mkF(this),
                },

                timerID: null,
            }
        },
        methods: {
            refresh: function() {
                let t = this;
                doGetJson(cronApiBase + "/timeline", true).then(timeline => {
                    let processed = timeline.map((row, i) => {
                        return [
                            row[0],
                            row[1],
                            t.toColor(row[1]),
                            t.tooltipHtml(row),
                            new Date(row[2] * 1000),
                            new Date(row[3] * 1000),

                        ]
                    });
                    processed.unshift([{
                        type: 'string',
                    }, {
                        type: 'string',
                    }, {
                        type: 'string',
                        role: 'style'
                    }, {
                        type: 'string',
                        role: 'tooltip',
                        p: {'html': true}
                    }, {
                        type: 'date',
                    }, {
                        type: 'date',
                    }]);

                    this.timeline = processed;
                })
            },
            chartHeight: function () {
                if (this.timeline.length == 0) {
                    return 200
                }

            },
            toColor: function (status) {
                switch (status) {
                    case 'NEW':
                        return this.COLOR_NEW;
                    case 'RUNNING':
                        return this.COLOR_RUNNING;
                    case 'DONE':
                        return this.COLOR_DONE;
                    case 'FAIL':
                        return this.COLOR_FAIL;
                    case 'ORPHAN':
                        return this.COLOR_ORPHAN;
                }
            },
            tooltipHtml: function (row) {
                let duration = moment.duration(1000 * (row[3] - row[2])).humanize();
                let from = moment(1000 * row[2]).format('HH:mm:ss');
                let to = moment(1000 * row[3]).format('HH:mm:ss');
                let color = 'black';

                return `
                <div>
                    <div class="google-visualization-tooltip" style="width: 212px; height: 137px;">
                        <ul class="google-visualization-tooltip-item-list">
                            <li class="google-visualization-tooltip-item">
                                <span style="font-family:Arial;font-size:12px;color:${color};opacity:1;margin:0;font-style:none;text-decoration:none;font-weight:bold;">${row[1]}</span>
<!--                                <div style="float: right; z-index: 1000000;"><button click="console.log('aaaa')">Логи</button></div>-->
                            </li>
                        </ul>
                        <div class="google-visualization-tooltip-separator"></div>
                        <ul class="google-visualization-tooltip-action-list">
                            <li data-logicalname="action#" class="google-visualization-tooltip-action">
                                <span style="font-family:Arial;font-size:12px;color:#000000;opacity:1;margin:0;font-style:none;text-decoration:none;font-weight:bold;">${row[0]}:</span>
                                <span style="font-family:Arial;font-size:12px;color:#000000;opacity:1;margin:0;font-style:none;text-decoration:none;font-weight:none;"> ${from} - ${to}</span>
                            </li>
                            <li data-logicalname="action#" class="google-visualization-tooltip-action">
                                <span style="font-family:Arial;font-size:12px;color:#000000;opacity:1;margin:0;font-style:none;text-decoration:none;font-weight:bold;">Длительность:</span>
                                <span style="font-family:Arial;font-size:12px;color:#000000;opacity:1;margin:0;font-style:none;text-decoration:none;font-weight:none;">${duration}</span>
                            </li>
                        </ul>
                    </div>
                </div>`
            },
            onChartReady: function (chart, google) {
                this.chart = chart;
            }
        }
    }
</script>
