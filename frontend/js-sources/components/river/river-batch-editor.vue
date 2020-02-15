<template>
    <div class="spot-editor-panel" style="padding-top:15px;">
        <div v-if="canEdit && river.id">
            <div style="margin-bottom: 20px">Изменения вступят в силу после нажатия кнопки <b>Сохранить</b> выше</div>

            <template>
                <div class="example-drag">
                    <div class="upload">
                        <ul v-if="files.length">
                            <li v-for="(file, index) in files" :key="file.id">
                                <span>{{file.name}}</span> -
                                <span>{{file.size}}</span> -
                                <span v-if="file.error">{{file.error}}</span>
                                <span v-else-if="file.success">success</span>
                                <span v-else-if="file.active">active</span>
                                <span v-else-if="file.active">active</span>
                                <span v-else></span>
                            </li>
                        </ul>

                        <div v-show="$refs.upload && $refs.upload.dropActive" class="drop-active">
                            <h3>Drop files to upload</h3>
                        </div>

                        <div class="example-btn">
                            <file-upload
                                    class="btn btn-primary"
                                    :headers="headers"
                                    :post-action="uploadPath"
                                    extensions="gpx"
                                    :multiple="false"
                                    :drop="false"
                                    :drop-directory="false"
                                    v-model="files"
                                    ref="uploadGpx">
                                <i class="fa fa-plus"></i>
                                Выберите GPX-файл с точками препятствий.
                            </file-upload>
                            <button type="button" class="btn btn-success"
                                    v-if="!$refs.uploadGpx || !$refs.uploadGpx.active"
                                    @click.prevent="$refs.uploadGpx.active = true; gpxJustUploaded=true;">
                                <i class="fa fa-arrow-up" aria-hidden="true"></i>
                                Начать загрузку
                            </button>
                            <button type="button" class="btn btn-danger" v-else
                                    @click.prevent="$refs.uploadGpx.active = false">
                                <i class="fa fa-stop" aria-hidden="true"></i>
                                Остановить загрузку
                            </button>
                        </div>
                    </div>
                </div>
            </template>
            <button type="button" class="btn btn-success" v-on:click="addEmptySpotToBatch()">Добавить в конец</button>
            <div>
                <div class="list-group" id="spot-list">
                    <div v-for="(spot, index) in spots"
                         :class="'container spot-edit-row ' + (spotsForDeleteIds.includes(spot.id) ? 'deleted-spot' : '')">
                        <div class="crossline"></div>
                        <div class="collapse wwmap-collapse" :id="'wwmap-collapse_'+index" aria-expanded="false">
                            <div class="spot-index" v-if="!spot.automatic_ordering && spot.order_index!='0'">
                                {{spot.order_index}}
                            </div>
                            <div class="row">
                                <div class="col-6">
                                    <input v-model.trim="spot.title"
                                           style="display:block; width: 100%; padding-bottom: 10px;   "/>
                                    <ya-map-location-and-coords :ref="'locationEdit_'+index"
                                                                v-bind:spot="spot"
                                                                width="100%"
                                                                height="400px"
                                                                :editable="true"
                                                                :ya-search="true"
                                                                :switch-type-hotkeys="false"
                                                                v-bind:refresh-on-change="spot.point"
                                                                :show-map-by-default="false"/>
                                </div>
                                <div class="col-4">
                                    <div>
                                        <strong>Категория сложности: </strong><a target="_blank"
                                                                                 href="https://huskytm.ru/rules2018-2019/#categories_tab"><img
                                            src="img/question_16.png"></a>
                                    </div>
                                    <div>
                                        <dl style="padding-left:40px;">
                                            <dt>По классификатору</dt>
                                            <dd>
                                                <category-select v-model="spot.category"></category-select>
                                            </dd>
                                            <dt>Низкий уровень воды</dt>
                                            <dd>
                                                <category-select v-model="spot.lw_category"></category-select>
                                            </dd>
                                            <dt>Средний уровень воды</dt>
                                            <dd>
                                                <category-select v-model="spot.mw_category"></category-select>
                                            </dd>
                                            <dt>Высокий уровень воды</dt>
                                            <dd>
                                                <category-select v-model="spot.hw_category"></category-select>
                                            </dd>
                                        </dl>
                                    </div>
                                </div>
                                <div class="col2">
                                    <div class="draggable">
                                        Для сортировки тащи меня вверх или вниз
                                    </div>
                                    <div>
                                        <button v-if="spotsForDeleteIds.includes(spot.id)" type="button"
                                                class="btn btn-secondary"
                                                style="z-index: 100000"
                                                v-on:click="spotsForDeleteIds = spotsForDeleteIds.filter(function(x) {
                          return x!=spot.id;
                        })">Не удалять
                                        </button>
                                        <button v-else type="button" class="btn btn-danger"
                                                v-on:click="spot.id ? spotsForDeleteIds.push(spot.id) : spots.splice(index, 1)">
                                            Удалить
                                        </button>
                                    </div>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-2"><strong>Описание: </strong></div>
                                <div class="col-10"><textarea rows="10" cols="120"
                                                              v-model="spot.short_description"></textarea>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-2"><strong>Ориентиры:</strong></div>
                                <div class="col-10"><textarea rows="10" style="width:100%"
                                                              v-model="spot.orient"></textarea>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-2"><strong>Подход/выход:</strong></div>
                                <div class="col-10"><textarea rows="10" cols="120" v-model="spot.approach"></textarea>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-2"><strong>Страховка:</strong></div>
                                <div class="col-10"><textarea rows="10" cols="120" v-model="spot.safety"></textarea>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-2"><strong>Описание для низкого уровня воды:</strong></div>
                                <div class="col-10"><textarea rows="10" cols="120"
                                                              v-model="spot.lw_description"></textarea>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-2"><strong>Описание для среднего уровня воды:</strong></div>
                                <div class="col-10"><textarea rows="10" cols="120"
                                                              v-model="spot.mw_description"></textarea>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-2"><strong>Описание для высокого уровня воды:</strong></div>
                                <div class="col-10"><textarea rows="10" cols="120"
                                                              v-model="spot.hw_description"></textarea>
                                </div>
                            </div>
                            <hr/>
                            <div class="row">
                                <div class="col-12"><strong>Другие варианты названия для поиска изображений:</strong>
                                    <div class="wwmap-system-hint" style="margin-bottom: 7px;">Каждое альтернативное
                                        название на
                                        новой строке
                                    </div>
                                    <textarea v-bind:text-content="spot.aliases"
                                              v-on:input="spot.aliases = parseAliases($event.target.value)"
                                              rows="10"
                                              cols="120">{{ spot.aliases ? spot.aliases.join('\n') : '' }}</textarea>
                                </div>
                            </div>
                        </div>
                        <button class="btn btn-light collapsed collapse-control"
                                style="width: 100%; height: 40px; margin-top: 10px;"
                                aria-expanded="false"
                                data-toggle="collapse"
                                :data-target="'#wwmap-collapse_'+index"
                                :aria-controls="'wwmap-collapse_'+index"></button>
                    </div>
                </div>
            </div>
            <button v-if="spots.length>0" type="button" class="btn btn-success" v-on:click="addEmptySpotToBatch()">
                Добавить
            </button>
        </div>
    </div>
</template>
<style type="text/css">
    .spot-edit-row {
        margin-left: 10px;
        margin-bottom: 35px;
        margin-top: 15px;
        border-top: #555555;
        border-top-style: dashed;
    }

    .spot-edit-row .collapse-control.collapsed:after {
        content: "Развернуть"
    }

    .spot-edit-row .collapse-control:not(.collapsed):after {
        content: "Свернуть";
    }

    .spot-edit-row .wwmap-collapse.collapse:not(.show) {
        display: block;
        height: 290px;
        overflow: hidden;
    }

    .spot-edit-row .wwmap-collapse.collapsing {
        display: block;
    }

    .spot-edit-row .row {
        margin-left: 0;
    }

    .deleted-spot {
        background: #ffeeee;
        position: relative;
        height: 290px;
        overflow: hidden;
    }

    .deleted-spot .collapse-control {
        display: none;
    }

    .deleted-spot .crossline {
        width: 100%;
        height: 1px;
        z-index: 10000;
        border-bottom: 5px solid red;
        -webkit-transform: translateY(145px) translateX(-15px) rotate(14deg);
        position: absolute;
    }

    .draggable {
        width: 140px;
        height: 140px;
        margin-top: 50px;
        margin-bottom: 20px;
        border: 1px;
        border-style: dashed;
        text-align: center;
        padding-top: 30px;
        cursor: grab;
    }

    .spot-index {
        margin-top: 10px;
        float: left;
        font-size: large;
        font-weight: bold;
        color: #555555;
        border: solid;
        border-radius: 40px;
        min-width: 30px;
        height: 30px;
        text-align: center;
        line-height: 21px;
    }
</style>

<script>
    import FileUpload from 'vue-upload-component';
    import {getWwmapSessionId, hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth'
    import {store} from '../../app-state'
    import {getSpotsFull, nvlReturningId, saveSpotBatch} from '../../editor'
    import {backendApiBase} from '../../config'
    import Sortable from 'sortablejs';

    var $ = require("jquery");
    require("jquery.cookie");

    module.exports = {
        props: ['river', 'reports', 'country', 'region'],
        components: {
            FileUpload: FileUpload
        },
        created: function () {
            let t = this;

            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);

            getSpotsFull(this.river.id).then(spots => {
                this.spots = spots;
                this.spotIndexes = [];
                this.pageMode = 'batch-edit';
                this.hideError();

                let el = document.getElementById('spot-list');
                if (!el) {
                    this.sortable = null;
                } else if (this.pageMode == 'batch-edit' && !this.sortable) {
                    this.sortable = new Sortable(el, {
                        animation: 150,
                        group: 'spotList',
                        draggable: '.spot-edit-row',
                        handle: '.draggable',
                        sort: true,
                        filter: '.sortable-disabled',
                        chosenClass: 'active',
                        onEnd: function (/**Event*/evt) {
                            if (t.spotIndexes.length <= evt.newIndex || t.spotIndexes.length <= evt.oldIndex || t.spotIndexes.length < t.spots.length) {
                                if (t.spotIndexes.length == 0) {
                                    t.spotIndexes = t.spots.map((_, idx) => idx)
                                } else {
                                    let maxEl = Math.max(...t.spotIndexes);
                                    for (let i = t.spotIndexes.length, idx = 0; i < t.spots.length; i++, idx++) {
                                        spotIndexes.push(maxEl + idx);
                                    }
                                }
                            }
                            let el1 = t.spotIndexes[evt.oldIndex];
                            t.spotIndexes.splice(evt.oldIndex, 1);
                            t.spotIndexes.splice(evt.newIndex, 0, el1);

                            for (let i = 0; i < t.spots.length; i++) {
                                let spotIdx = t.spotIndexes[i];
                                t.spots[spotIdx].order_index = "" + (i + 1);
                                t.spots[spotIdx].automatic_ordering = false;

                            }
                        }
                    });
                }
            });
        },
        updated() {
            let t = this;
            if (this.$refs.uploadGpx && this.$refs.uploadGpx.value.length && this.$refs.uploadGpx.uploaded && this.gpxJustUploaded) {
                store.commit('showRiverSubentities', {
                    countryId: this.river.region.country_id,
                    regionId: nvlReturningId(this.river.region),
                    riverId: this.river.id,
                });
                this.gpxJustUploaded = false;
                for (let i = 0; i < this.$refs.uploadGpx.value.length; i++) {
                    this.$refs.uploadGpx.value[i].response.forEach(x => t.spots.push(x));
                }
            }
        },
        computed: {
            uploadPath: function () {
                return backendApiBase + "/river/" + this.river.id + "/gpx"
            },
            headers: function () {
                return {
                    Authorization: getWwmapSessionId()
                }
            },
            pageMode: {
                get: function () {
                    return store.state.rivereditorstate.pageMode
                },

                set: function (newVal) {
                    store.commit("setRiverEditorPageMode", newVal);
                }
            },
        },
        data: function () {
            return {
                map: null,
                canEdit: false,
                files: [],

                gpxJustUploaded: false,

                spots: [],
                spotIndexes: [],
                spotsForDeleteIds: [],
            }
        },
        methods: {
            saveSpotsBatch: function () {
                this.hideError();
                let forDelete = this.spotsForDeleteIds;
                let forUpdate = this.spots
                    .filter(spot => !forDelete.includes(spot.id));

                saveSpotBatch({
                    "delete": forDelete,
                    "update": forUpdate,
                }).then(_ => {

                    let countryId = this.river.region.country_id;
                    let regionId = nvlReturningId(this.river.region);
                    let riverId = this.river.id;
                    store.commit('showRiverSubentities', {
                        countryId: countryId,
                        regionId: regionId,
                        riverId: riverId,
                    });

                    this.pageMode = 'view';
                }, err => this.showError(err));
            },
            addEmptySpotToBatch: function () {
                let orderIndex;
                let automaticOrdering;

                if (this.spots.length == 0) {
                    orderIndex = 1;
                    automaticOrdering = false;
                } else if (this.spotIndexes.length != 0) {
                    orderIndex = Math.max(...this.spotIndexes) + 2; // order_index starts from 1
                    automaticOrdering = false;
                } else if (this.spots.every(s => !s.automatic_ordering)) {
                    orderIndex = Math.max(...this.spots.map(s => s.order_index)) + 1;
                    automaticOrdering = false;
                } else {
                    orderIndex = 0;
                    automaticOrdering = true;
                }

                let location = this.spots.map(s => s.point).reduce((p1, p2) => [p1[0] + p2[0], p1[1] + p2[1]], [0, 0]);
                if (this.spots.length > 0) {
                    location = [location[0] / this.spots.length, location[1] / this.spots.length]
                }

                this.spots.push({
                    id: 0,
                    river_id: this.river.id,
                    river: {id: this.river.id},
                    point: location,
                    order_index: "" + (orderIndex),
                    automatic_ordering: automaticOrdering,
                    aliases: [],
                })
            },

            showError: function (errMsg) {
                store.commit("setErrMsg", errMsg);
            },
            hideError: function () {
                store.commit("setErrMsg", null);
            },
        }
    }

</script>