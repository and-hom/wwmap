<template>
    <page link="timeline.htm">
        <ask id="orphan-info" title="Потерян контроль над задачей" :noBtn="false"
             msg="Если сервис был перезапущен, задачи продолжают выполняться, но отследить их, а также смотреть их логи уже нельзя"></ask>

        <div style="margin-left:10px; margin-top: 10px; ">
            <h2>Таймлайн и логи</h2>
            <div style="display: flex;">
                <div style="display:inline-block">
                    <div style="display:inline-block;">
                        <ul class="legend">
                            <li>
                                <svg>
                                    <rect :fill="COLOR_RUNNING"/>
                                </svg>
                                Выполняется
                            </li>
                            <li>
                                <svg>
                                    <rect :fill="COLOR_DONE"/>
                                </svg>
                                Выполнено
                            </li>
                            <li>
                                <svg>
                                    <rect :fill="COLOR_FAIL"/>
                                </svg>
                                Ошибка
                            </li>
                            <li>
                                <svg>
                                    <rect :fill="COLOR_ORPHAN"/>
                                </svg>
                                Потерян контроль<a href="#" data-toggle="modal" data-target="#orphan-info"><img
                                    src="img/question_16.png"></a></li>
                        </ul>
                    </div>
                    <div style="display: inline-block; margin-left: 15px; margin-right: 15px; float: right;">
                        <vue-slider v-model="from"
                                    v-bind="dateSliderOptions"
                                    @drag-end="onTimeOffsetChanged()"/>
                    </div>
                    <GChart
                            :settings="{packages: ['timeline']}"
                            :data="timeline"
                            :options="chartOptions"
                            :createChart="(el, google) => new google.visualization.Timeline(el)"
                            :events="chartEvents"
                            @ready="onChartReady"
                    />
                </div>
                <div style="display:inline-block">
                    <div class="log" v-if="log.title">
                        <div class="log-title">{{log.title}}</div>
                        <div class="log-description">{{log.description}}</div>
                        <b-tabs>
                            <b-tab title="stderr" active>
                                <div class="log-body">{{log.stderr}}</div>
                            </b-tab>
                            <b-tab title="stdout">
                                <div class="log-body">{{log.stdout}}</div>
                            </b-tab>
                        </b-tabs>
                    </div>
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
        display: inline;
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

    .log {
        margin-left: 15px;
    }

    .log-title {
        font-weight: bold;
    }

    .log-description {
        font-style: italic;
    }

    .log-body {
        margin-top: 7px;
        font-size: 80%;
        background: #ffeedd;
        white-space: pre;
    }
</style>


<script>
    import {doGet, doGetJson} from '../api'
    import {cronApiBase} from '../config'

    const moment = require('moment');

    const DEFAULT_OFFSET_HOURS = 24;

    export default {
        created: function () {

        },
        mounted: function () {
            this.refresh();
            let t = this;
            this.timerID = setInterval(function () {
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
                        isHtml: true,
                    },
                    hAxis: {
                        minValue: new Date(Date.now() - DEFAULT_OFFSET_HOURS * 3600 * 1000),
                        maxValue: new Date(),
                    }
                },
                COLOR_NEW: "#fffa96",
                COLOR_RUNNING: "#9698ff",
                COLOR_DONE: "#57ff00",
                COLOR_FAIL: "#ff8789",
                COLOR_ORPHAN: "#afafaf",

                chartEvents: {
                    'select': () => {
                        if (this.chart) {
                            let t = this;
                            let selection = this.chart.getSelection();
                            if (selection.length > 0) {
                                let rowIdx = selection[0].row + 1;
                                let title = this.timeline[rowIdx][0];
                                let status = this.timeline[rowIdx][1];
                                let from = this.timeline[rowIdx][4];
                                let to = this.timeline[rowIdx][5];
                                let executionId = this.timeline[rowIdx][6];
                                t.log = {
                                    "title": `${title} - ${status}`,
                                    "description": `${from} - ${to}`,
                                    "stdout": "",
                                    "stderr": "",
                                };
                                doGet(`${cronApiBase}/logs/${executionId}/out`, true).then(data => {
                                    t.log.stdout = data
                                }).catch(e => t.log.stdout = e);
                                doGet(`${cronApiBase}/logs/${executionId}/err`, true).then(data => {
                                    t.log.stderr = data
                                }).catch(e => t.log.stderr = e);
                            }
                        }
                    },
                },

                from: DEFAULT_OFFSET_HOURS,
                timerID: null,
                log: this.noLog(),

                dateSliderOptions: {
                    dotSize: 14,
                    width: 500,
                    height: 4,
                    min: 1,
                    max: 72,
                    adsorb: true,
                    direction: 'rtl',
                    interval: 1,
                    tooltipFormatter: val => {
                        if (val === 1) {
                            return `Начиная с 1 часа назад`
                        } else {
                            return `Начиная с ${val} часов назад`
                        }
                    },
                }
            }
        },
        methods: {
            refresh: function () {
                let t = this;
                doGetJson(`${cronApiBase}/timeline?fromTimeOffset=${-t.from}`, true).then(timeline => {
                    let processed = timeline.map((row, _) => {
                        return [
                            row.title,
                            row.status,
                            t.toColor(row.status),
                            t.tooltipHtml(row),
                            new Date(row.start * 1000),
                            new Date(row.end * 1000),
                            row.execution_id,

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
                    }, {
                        type: 'number',
                        role: 'hidden',
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
                let duration = moment.duration(1000 * (row.end - row.start)).humanize();
                let from = moment(1000 * row.start).format('HH:mm:ss');
                let to = moment(1000 * row.end).format('HH:mm:ss');
                let color = 'black';

                return `
                <div>
                    <div class="google-visualization-tooltip" style="width: 212px;">
                        <ul class="google-visualization-tooltip-item-list">
                            <li class="google-visualization-tooltip-item">
                                <span style="font-family:Arial;font-size:12px;color:${color};opacity:1;margin:0;text-decoration:none;font-weight:bold;">${row.status + (row.squashed_count>1 ? ` [ x${row.squashed_count} ]` : '')}</span>
                            </li>
                            ` + (row.manual ?
                    `
                            <li class="google-visualization-tooltip-item" style="margin-top: -10px;">
                                <span style="font-family:Arial;font-size:10px;color:red;opacity:1;margin:0;text-decoration:none;font-weight:bold;">Запуск вручную</span>
                            </li>
                            ` : '') +
                    `
                        </ul>
                        <div class="google-visualization-tooltip-separator"></div>
                        <ul class="google-visualization-tooltip-action-list">
                            <li data-logicalname="action#" class="google-visualization-tooltip-action">
                                <span style="font-family:Arial;font-size:12px;color:#000000;opacity:1;margin:0;font-style:none;text-decoration:none;font-weight:bold;">${row.title}:</span>
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
            },

            noLog: function () {
                return {
                    title: "",
                    decription: "",
                    stdout: "",
                    stderr: "",
                }
            },
            onTimeOffsetChanged: function () {
                this.chartOptions.hAxis = {
                    minValue: new Date(Date.now() - this.from * 3600 * 1000),
                    maxValue: new Date(),
                };
                this.refresh();
            },
        }
    }
</script>
