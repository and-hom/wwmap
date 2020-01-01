<template>
    <div>
        <datepicker format="yyyy-MM-dd" v-model="date" @selected="function (d) {
                    levelData = setImageDate(spotId, imgId, d);
                    $emit('input', d);
                    $emit('level', levelData)
                }" :clear-button="true"></datepicker>

        <div style="margin-top: 10px;">
            Задать уровень воды на фото вручную:
            <div>
                <image-rating src="img/lvl5.png" @rating-selected="manualLevel = $event" style="display: inline-block"
                              :rating="manualLevel"></image-rating>
                <div class="vdp-datepicker__clear-button"
                     style="margin-left: 3px; height: 100%; display: inline-block"
                     @click.prevent="manualLevel = null">×</div>
                <div class="wwmap-system-hint" style="width: 300px;">Уровень воды по шкале от 1 до 5. 1 - крайне низкая вода, 5 - сильный паводок. Крестик правее элемента сбрасывает заданное значение.</div>
            </div>
        </div>


        <div style="margin-top: 12px;" v-for="(value, sensorId) in levelData">
            <svg style="vertical-align: middle" width="40px" height="40px" viewBox="0 0 2 2" class="donut">
                <title>{{getLevelTitle(value)}}</title>
                <circle class="donut-ring" cx="1" cy="1" r="0.795774715" :fill="getBaseFillColor(value)"
                        :stroke="getBaseColor(value)"
                        stroke-width="0.28"></circle>
                <circle v-if="value>=0" class="donut-segment" cx="1" cy="1" r="0.795774715" fill="transparent"
                        stroke="#2222ff"
                        stroke-width="0.28" :stroke-dasharray="value +' ' + (5-value)" stroke-dashoffset="2.5"></circle>
                <g v-if="value>=0">
                    <text x="50%" y="50%" class="chart-number">
                        {{value}}/5
                    </text>
                </g>
            </svg>
            <span v-if="sensorId!='0'">
                &nbsp{{sensorId}}: {{sensorName(sensorId)}}
            </span>
            <span v-else>
                Уровень задан вручную
            </span>
        </div>
        <div class="wwmap-system-hint" style="width: 300px;" v-if="Object.keys(levelData).length == 0">
            Нет показаний за выбранную дату. Автоматическая настройка уровня не сработает, но вы можете задать уровень воды вручную.
        </div>
    </div>
</template>

<style type="text/css">
    .chart-number {
        font-weight: bold;
        font-size: 0.04em;
        text-anchor: middle;
        -moz-transform: translateY(0.03em);
        -ms-transform: translateY(0.03em);
        -webkit-transform: translateY(0.03em);
        transform: translateY(0.03em);
    }
</style>


<script>
    module.exports = {
        props: {
            spotId: {
                type: Number,
                required: true,
            },
            imgId: {
                type: Number,
                required: true,
            },
            value: {
                validator: prop => typeof prop === 'object' || prop === null,
                required: true,
            },
            level: {
                type: Object,
                required: true,
            }
        },
        computed: {
            manualLevel: {
                get: function () {
                    let mLvl = this.levelData[0];
                    return mLvl;
                },
                set: function (mLvl) {
                    if (mLvl && mLvl>0) {
                        this.levelData = setManualLevel(this.spotId, this.imgId, mLvl);
                    } else {
                        this.levelData = resetManualLevel(this.spotId, this.imgId)
                    }
                },
            }
        },
        data: function () {
            return {
                date: this.value,
                levelData: this.level,
                setImageDate: setImageDate,
                sensorName: function (sensorId) {
                    return sensorsById[sensorId];
                },
                getLevelTitle: function (value) {
                    if (value == -2) {
                        return "Ошибка поиска датчика в базе"
                    }
                    if (value == -1) {
                        return "Нет данных по датчику за " + this.date
                    }
                    if (value < 2) {
                        return "Низкая вода"
                    }
                    if (value > 3) {
                        return "Высокая вода"
                    }
                    if (value == 2 || value == 3) {
                        return "Средняя вода"
                    }
                    return "Неизвестное значение " + value
                },
                getBaseColor: function (value) {
                    if (value >= 0) {
                        return "#d2d3d4"
                    } else if (value == -1) {
                        return "#aaaaaa"
                    }
                    return "#ff5555"
                },
                getBaseFillColor: function (value) {
                    if (value >= 0) {
                        return "transparent"
                    }
                    return this.getBaseColor(value)
                },
            }
        },
    }
</script>