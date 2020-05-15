<template>
    <div v-if="showMap" style="margin-top: 20px;">
        <div style="margin-bottom: 6px;">
            <button type="button" v-if="showMap" class="btn btn-info"
                    v-on:click="showMap=false" style="margin-right: 5px;">Редактировать координаты</button>
            <div style="padding-top:4px; display: inline-block; font-size: smaller;">
                <strong>Широта:</strong>&nbsp;{{points[0][0] }}&nbsp;&nbsp;<strong>Долгота:</strong>&nbsp;{{
                points[0][1] }}
            </div>
        </div>
        <ya-map-location
                v-bind:spot="spot"
                :zoom="zoom"
                :width="width"
                :height="height"
                :editable="editable"
                :ya-search="yaSearch"
                :refresh-on-change="refreshOnChange"
                :default-map="defaultMap"
                :switch-type-hotkeys="switchTypeHotkeys"
                v-on:spotClick = "$emit('spotClick', $event)"/>
    </div>
    <div v-else :style="divStyle">
        <div class="btn-group" role="group" aria-label="Basic example">
            <button type="button" v-if="!showMap" data-toggle="button" :aria-pressed="mode=='raw'"
                    class="btn btn-primary" v-on:click="mode='raw'">50.77465369743°
            </button>
            <button type="button" v-if="!showMap" data-toggle="button" :aria-pressed="mode=='deg-min-sec'"
                    class="btn btn-success" v-on:click="mode='deg-min-sec'">N50° 46' 28.753''
            </button>
            <button type="button" v-if="!showMap" data-toggle="button" :aria-pressed="mode=='deg-min'"
                    class="btn btn-warning" v-on:click="mode='deg-min'">N50° 46.47922'
            </button>
            <button type="button" v-if="!showMap" class="btn btn-info" v-on:click="showMap=true">Карта</button>
        </div>
        <div v-if="points.length>1" class="wwmap-system-hint">Добавления/удаление точек протяжённого порога тут пока не сделано. Пользуйтесь редактором порога, выбрав его в дереве слева и нажав "Редактирование"</div>
        <div v-for="(point, index) in points">
            <div v-if="points.length > 1 && index == 0" style="font-weight: bold">Начало:</div>
            <div v-else style="height: 10px;">&nbsp;</div>
            <div v-if="points.length > 1 && index == points.length - 1"
                 style="font-weight: bold; margin-top: 12px;">Конец:
            </div>
            <lat-lon-input v-bind:value="point" :mode="mode" v-on:input="setPoint($event, index)"/>
            <div v-if="index==0" style="height: 20px; display: block"></div>
        </div>
    </div>
</template>

<script>
    module.exports = {
        props: {
            spot: Object,
            zoom: {
                type: Number,
                default: 12
            },
            width: {
                type: String,
                default: "600px"
            },
            height: {
                type: String,
                default: "400px"
            },
            editable: {
                type: Boolean,
                default: false
            },
            yaSearch: {
                type: Boolean,
                default: false
            },
            refreshOnChange: {
                default: null
            },
            defaultMap: {
                type: String,
                default: null
            },
            switchTypeHotkeys: {
                type: Boolean,
                default: true
            },
            showMapByDefault: {
                type: Boolean,
                default: true
            },
        },
        computed: {
            points: {
                get: function () {
                    if (Array.isArray(this.spot.point[0])) {
                        return this.spot.point;
                    } else if (this.spot.point) {
                        return [this.spot.point];
                    } else {
                        return [];
                    }
                },
            }
        },
        data: function () {
            return {
                showMap: this.showMapByDefault,
                mode: 'raw',
                divStyle: 'width: ' + this.width + '; height: ' + this.height + ';overflow: scroll; overflow-x: hidden; margin-top: 20px; margin-bottom: 15px;',
            }
        },
        methods: {
            setPoint: function (p, index) {
                if (index >= 0 && index < this.points.length) {
                    if (Array.isArray(this.spot.point[0])) {
                        this.spot.point[index] = p
                    } else {
                        this.spot.point = p;
                    }
                } else {
                    console.error(`Can't set point at index ${index} for array ${this.spot.point}`)
                }
            },
        },
    }

</script>