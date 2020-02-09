<template>
    <page link="level.htm">
        <div style="margin-left:10px; margin-top: 10px;">
            <h2>Уровни воды</h2>
            <div class="container" style="width: 100%; margin-left: 0; margin-bottom: 15px;">
                <div class="row">
                    <div class="col-1">
                        С даты:
                    </div>
                    <div class="col-2">
                        <datepicker format="yyyy-MM-dd" v-model="fromDate" :clear-button="false"></datepicker>
                    </div>
                </div>
                <div class="row">
                    <div class="col-1">
                        По дату:
                    </div>
                    <div class="col-2">
                        <datepicker format="yyyy-MM-dd" v-model="toDate" :clear-button="false"></datepicker>
                    </div>
                </div>
            </div>
            <table class="table">
                <thead>
                <tr>
                    <td style="width: 10%">Датчик</td>
                    <td style="width: 20%">Оригинальный информер</td>
                    <td style="width: 20%">Считанный график</td>
                    <td style="width: 50%">Реки, для которых используется датчик</td>
                </tr>
                </thead>
                <tr v-for="(data,id) in sensors">
                    <td>
                        <div style="font-weight: bold;">{{id}}</div>
                        <div>{{sensorName(id)}}</div>
                        <div style="margin-top: 10px">
                            <span style="font-weight: bold">Градации уровня:</span>
                            <ul>
                                <li>{{data.sensor_metrics.l0}}</li>
                                <li>{{data.sensor_metrics.l1}}</li>
                                <li>{{data.sensor_metrics.l2}}</li>
                                <li>{{data.sensor_metrics.l3}}</li>
                            </ul>
                        </div>
                    </td>
                    <td>
                        <img :src="'http://gis.vodinfo.ru/informer/draw/v2_' + id + '_400_300_10_ffffff_110_8_7_H_none.png'"/>
                    </td>
                    <td>
                        <canvas :id="canvasId(id)" width="400" height="300" :ref="canvasId(id)"></canvas>
                    </td>
                    <td>
                        <ul>
                            <li v-for="river in sensors[id].rivers">
                                <a :href="'https://wwmap.ru/editor.htm#' + river.region.country_id +',' + river.region.id + ',' + river.id">{{river.title}}</a>
                            </li>
                        </ul>
                    </td>
                </tr>
            </table>
        </div>
    </page>
</template>

<script>
    import {backendApiBase} from '../config'
    import {doGetJson} from '../api'
    import Chart from 'chart.js';
    import {sensorsById} from '../sensors'

    const moment = require('moment');

    export default {
        data() {
            return {
                sensors: [],
                firstLoad: true,
                fromDate: function () {
                    let d = new Date();
                    d.setDate(d.getDate() - 10);
                    return d;
                }(),
                toDate: new Date(),
                charts: {},
                canvasId: function (id) {
                    return 'line' + id
                },
                sensorName: function (id) {
                    return sensorsById[parseInt(id)]
                },
                onLoadPlot: function (id) {
                    let canvas = this.$refs[this.canvasId(id)];
                    var ctx = (Array.isArray(canvas) ? canvas[0] : canvas).getContext('2d');
                    var chartData = this.sensors[id].chart_data;
                    var min = null;
                    var max = null;
                    for (var i in chartData.datasets[0].data) {
                        var l = chartData.datasets[0].data[i];
                        if (l == null) {
                            continue
                        }
                        if (l < min || min == null) {
                            min = l
                        }
                        if (l > max || max == null) {
                            max = l
                        }
                    }
                    if (max - min < 120) {
                        var border = (120 - max + min) / 2;
                        max += border;
                        min -= border;
                    }
                    max = Math.round(max / 10) * 10;
                    min = Math.round(min / 10) * 10;

                    let existing = this.charts[id];
                    if (existing) {
                        existing.data = chartData;
                        existing.options.scales.yAxes[0].ticks.min = min;
                        existing.options.scales.yAxes[0].ticks.max = max;
                        existing.update();
                    } else
                        this.charts[id] = new Chart(ctx, {
                            type: 'line',
                            data: chartData,
                            options: {
                                title: {
                                    display: true,
                                    text: this.sensorName(id)
                                },
                                tooltips: {
                                    mode: 'index',
                                    intersect: false,
                                },
                                hover: {
                                    mode: 'nearest',
                                    intersect: true
                                },
                                legend: {
                                    display: false,
                                },

                                scales: {
                                    xAxes: [{
                                        display: true,
                                        ticks: {
                                            callback: function (dataLabel, index) {
                                                return dataLabel
                                            }
                                        }
                                    }],
                                    yAxes: [{
                                        ticks: {
                                            stepSize: 10,
                                            min: min,
                                            max: max,
                                        }
                                    }]
                                }
                            }
                        });
                },
                getSensors: function (fromDate, toDate) {
                    let params = {
                        from: fromDate == null ? null : moment(fromDate).format('YYYY-MM-DD'),
                        to: toDate == null ? null : moment(toDate).format('YYYY-MM-DD')
                    };

                    return doGetJson(backendApiBase + "/dashboard/levels?" + jQuery.param(params))
                },
                refreshSensorData: function () {
                    this.getSensors(this.fromDate, this.toDate).then(sensors => {
                        this.sensors = sensors;
                    });
                },
                renderSensorPlots: function () {
                    for (let id in this.sensors) {
                        this.onLoadPlot(id);
                    }
                },
            }
        },

        watch: {
            fromDate: {
                handler: function (val, oldVal) {
                    this.refreshSensorData();
                }
            },
            toDate: {
                handler: function (val, oldVal) {
                    this.refreshSensorData();
                }
            },
        },

        updated: function () {
            this.renderSensorPlots();
        },
        mounted: function () {
            this.refreshSensorData();
        },
    }
</script>
