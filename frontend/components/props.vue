<template>
<div class="container-fluid" style="margin-top: 20px;">
                        <div class="row">
                            <div class="col-3">
                                <strong>Подложка карты при рендеринге порогов в PDF</strong>
                            </div>
                            <div class="col-9">
                                <select v-model="p.pdf_map_type">
                                        <option :value="null">Как у остальных препятствий на реке</option>
                                        <option value="google#satellite">Спутник Google</option>
                                        <option value="yandex#satellite">Спутник Яндекс</option>
                                        <option value="ggc#standard">Топографичсекая карта</option>
                                </select>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-3">
                                <strong>Масштаб карты при рендеринге порогов в PDF</strong>
                            </div>
                            <div class="col-9">
                                <select v-model="p.pdf_map_zoom">
                                    <option :value="null">Как у остальных препятствий на реке</option>
                                    <option v-for="zoom in 5" v-bind:value="zoom + 11">{{zoom + 11}}</option>
                                </select>
                            </div>
                        </div>
                    </div>
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