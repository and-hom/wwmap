<template>
<div class="container-fluid wwmap-river-techincal-props" style="margin-top: 20px;">
    <slot name="before"></slot>
    <div class="row">
        <div class="col-3">
            <strong>Подложка карты при экспорте</strong>
        </div>
        <div class="col-9">
            <select v-model="p.spot_export_map_type">
                    <option :value="null">По-умолчанию</option>
                    <option value="google#satellite">Спутник Google</option>
                    <option value="yandex#satellite">Спутник Яндекс</option>
                    <option value="osm#standard">OSM</option>
                    <option value="ggc#standard">Топографичсекая карта</option>
            </select>
        </div>
    </div>
    <div class="row">
        <div class="col-3">
            <strong>Масштаб карты при экспорте</strong>
        </div>
        <div class="col-9">
            <select v-model="p.spot_export_map_zoom">
                <option :value="null">Как у остальных препятствий на реке</option>
                <option v-for="zoom in 5" v-bind:value="zoom + 11">{{zoom + 11}}</option>
            </select>
        </div>
    </div>
    <slot></slot>
</div>
</template>

<script>
    module.exports = {
        props: ['p'],
        data: function() {
            return {
            }
        }
    }
</script>

<style>
</style>