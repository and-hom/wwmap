<template>
    <div>
        <window-title v-bind:text="river.title"></window-title>
        <ask id="del-river" title="Точно?"
             msg="Совсем удалить? Все пороги будут также удалены! Да, совсем! Восстановить будет никак нельзя!"
             :ok-fn="function() { remove(); }"></ask>

        <div v-if="canEdit()" class="btn-toolbar justify-content-between">
            <div class="btn-group mr-2" role="group">
                <button v-if="river.id && pageMode == 'view'" type="button" class="btn btn-primary" v-on:click="add_spot()">Добавить препятствие</button>
                <button type="button" class="btn btn-info" v-if="pageMode == 'view'" v-on:click="pageMode='edit'; hideError();">
                    Редактирование
                </button>
                <button type="button" class="btn btn-success" v-if="river.id && pageMode == 'view' && userInfo.experimental_features" v-on:click="spots = getSpots(river.id); spotIndexes=[]; pageMode='batch-edit'; hideError();">
                    Пакетное редактирование и загрузка GPX
                </button>
                <button type="button" class="btn btn-success" v-if="pageMode == 'edit'" v-on:click="pageMode=save() ? 'view' : 'edit'">Сохранить</button>
                <button type="button" class="btn btn-success" v-if="pageMode == 'batch-edit'" v-on:click="pageMode=saveSpotsBatch() ? 'view' : 'batch-edit'">Сохранить</button>
                <button type="button" class="btn btn-secondary" v-if="pageMode != 'view'" v-on:click="pageMode='view'; cancelEditing()">Отменить</button>
                <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-river">Удалить
                </button>
            </div>
            <div class="btn-group mr-2">
                <log-dropdown object-type="RIVER" :object-id="river.id"/>
            </div>
        </div>

        <div v-if="pageMode == 'edit'" class="spot-editor-panel" style="padding-top:15px;">
                    <b-tabs>
                        <b-tab title="Главное" active>
                            <input v-model.trim="river.title" style="display:block"/>
                            <dl>
                                <dt>Показывать на карте:</dt>
                                <dd>
                                    <span style="padding-left:40px;" v-if="river.visible">Да</span>
                                    <span style="padding-left:40px;" v-else>Нет</span>&nbsp;&nbsp;&nbsp;
                                    <button type="button" class="btn btn-info" v-if="canEdit() && pageMode != 'edit' && !river.visible" v-on:click="setVisible(true); hideError();">
                                        Показывать на карте
                                    </button>
                                    <button type="button" class="btn btn-info" v-if="canEdit() && pageMode != 'edit' && river.visible" v-on:click="setVisible(false); hideError();">
                                        Скрыть на карте
                                    </button>
                                    <div style="padding-left:40px;" class="wwmap-system-hint">Нужно, когда мы не хотим выставлять наполовину размеченную и описанную реку.
                                    Если добавляешь часть порогов, а остальные планируешь на потом, не делай реку видимой на карте.</div>
                                </dd>
                                <dt v-if="river.region.id>0">Регион:</dt>
                                <dd v-if="river.region.id>0">
                                    <select v-model="river.region.id">
                                        <option v-for="region in regions()" v-bind:value="region.id">{{region.title}}</option>
                                    </select>
                                </dd>
                                <dt>Описание:</dt>
                                <dd>
                                    <textarea v-bind:text-content="river.aliases"
                                              rows="10" cols="120"
                                              style="resize: none; margin-left:40px;" v-model="river.description"></textarea>
                                </dd>
                                <dt>Другие варианты названия для автоматического поиска отчётов:</dt>
                                <dd>
                                    <div class="wwmap-system-hint" style="margin-bottom: 7px;">Каждое альтернативное название на новой строке</div>
                                    <textarea v-bind:text-content="river.aliases"
                                              v-on:input="river.aliases = parseAliases($event.target.value)"
                                              rows="10" cols="120"
                                              style="resize: none; margin-left:40px;">{{ river.aliases.join('\n') }}</textarea>
                                </dd>
                            </dl>
                        </b-tab>
                        <b-tab title="Системные параметры">
                            <span class="wwmap-system-hint" style="padding-top: 10px;">Тут собраны настройки разных системных вещей для этой реки</span>
                            <props :p="river.props">
                                <template slot="before">
                                    <div class="row">
                                        <div class="col-3">
                                            <strong>Подложка общей карты реки при экспорте</strong>
                                        </div>
                                        <div class="col-9">
                                            <select v-model="river.props.river_export_map_type">
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
                                            <strong>Гидропост <a href="http://gis.vodinfo.ru/informer/">gis.vodinfo.ru/informer</a></strong>
                                        </div>
                                        <div class="col-9">
                                            <v-select v-model="activeSensor" label="title" :options="sensors"
                                                      @input="onSelectSensor" >
                                                <template slot="no-options">
                                                    Начните печатать название гидропоста
                                                </template>
                                                <template slot="option" slot-scope="option">
                                                    {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}
                                                </template>
                                                <template slot="selected-option" slot-scope="option">
                                                    {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}
                                                </template>
                                            </v-select>
                                        </div>
                                    </div>
                                    <div class="row">
                                        <div class="col-3">
                                            <strong>Точка отслеживания погоды</strong>
                                            <div class="wwmap-system-hint">Сейчас используется только для формирования ссылки на прогноз погоды.</div>
                                        </div>
                                        <div class="col-9">
                                            <div v-if="meteoPointSelectMode">
                                                <div class="wwmap-system-hint">Выберите из списка или <button v-on:click.stop="addMeteoPoint">создайте</button></div>
                                                <v-select v-model="meteoPoint" label="title" :options="meteoPoints" @input="onSelectMeteoPoint" >
                                                    <template slot="no-options">
                                                        Начните печатать название точки
                                                    </template>
                                                    <template slot="option" slot-scope="option">
                                                        {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}
                                                    </template>
                                                    <template slot="selected-option" slot-scope="option">
                                                        {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}
                                                    </template>
                                                </v-select>
                                            </div>
                                            <div v-else>
                                                <div style="padding-top:15px;">
                                                    <ya-map-location v-bind:spot="meteoPoint" width="100%" height="600px" :editable="true" :ya-search="true"/>
                                                </div>
                                                <label style="padding-right: 10px;" for="meteo_point_title_input"><strong>Название:</strong></label><input
                                                    id="meteo_point_title_input" type="text" v-model="meteoPoint.title" style="margin-top: 10px; width: 80%;"/>
                                                <div class="btn-toolbar" style="padding-top:15px;">
                                                    <div class="btn-group mr-2" role="group">
                                                        <button type="button" class="btn btn-success" v-on:click.stop="onAddMeteoPoint" :disabled="!meteoPoint.title">Добавить</button>
                                                        <button type="button" class="btn btn-cancel" v-on:click.stop="onCancelAddMeteoPoint">Отмена</button>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                    <hr/>
                                    <h2>Для отдельных порогов:</h2>
                                </template>
                            </props>
                        </b-tab>
                    </b-tabs>
        </div>
        <div v-else-if="pageMode == 'batch-edit'" class="spot-editor-panel" style="padding-top:15px;">
            <div style="margin-bottom: 20px">Изменения вступят в силу после нажатия кнопки <b>Сохранить</b> выше</div>

            <div v-if="canEdit() && river.id">
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
            </div>
            <button type="button" class="btn btn-success" v-on:click="addEmptySpotToBatch()">Добавить в конец</button>
            <div>
                <div class="list-group" id="spot-list">
                <div v-for="(spot, index) in spots"
                     :class="'container spot-edit-row ' + (spotsForDeleteIds.includes(spot.id) ? 'deleted-spot' : '')">
                    <div class="crossline"></div>
                    <div class="collapse wwmap-collapse" :id="'wwmap-collapse_'+index" aria-expanded="false">
                        <div class="spot-index" v-if="!spot.automatic_ordering && spot.order_index!='0'">{{spot.order_index}}</div>
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
                                            <select v-model="spot.category">
                                                <option v-for="cat in all_categories" v-bind:value="cat.id">
                                                    {{cat.title}}
                                                </option>
                                            </select>
                                        </dd>
                                        <dt>Низкий уровень воды</dt>
                                        <dd>
                                            <select v-model="spot.lw_category">
                                                <option v-for="cat in all_categories" v-bind:value="cat.id">
                                                    {{cat.title}}
                                                </option>
                                            </select>
                                        </dd>
                                        <dt>Средний уровень воды</dt>
                                        <dd>
                                            <select v-model="spot.mw_category">
                                                <option v-for="cat in all_categories" v-bind:value="cat.id">
                                                    {{cat.title}}
                                                </option>
                                            </select>
                                        </dd>
                                        <dt>Высокий уровень воды</dt>
                                        <dd>
                                            <select v-model="spot.hw_category">
                                                <option v-for="cat in all_categories" v-bind:value="cat.id">
                                                    {{cat.title}}
                                                </option>
                                            </select>
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
                            <div class="col-10"><textarea rows="10" style="width:100%" v-model="spot.orient"></textarea>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-2"><strong>Подход/выход:</strong></div>
                            <div class="col-10"><textarea rows="10" cols="120" v-model="spot.approach"></textarea></div>
                        </div>
                        <div class="row">
                            <div class="col-2"><strong>Страховка:</strong></div>
                            <div class="col-10"><textarea rows="10" cols="120" v-model="spot.safety"></textarea></div>
                        </div>
                        <div class="row">
                            <div class="col-2"><strong>Описание для низкого уровня воды:</strong></div>
                            <div class="col-10"><textarea rows="10" cols="120" v-model="spot.lw_description"></textarea>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-2"><strong>Описание для среднего уровня воды:</strong></div>
                            <div class="col-10"><textarea rows="10" cols="120" v-model="spot.mw_description"></textarea>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-2"><strong>Описание для высокого уровня воды:</strong></div>
                            <div class="col-10"><textarea rows="10" cols="120" v-model="spot.hw_description"></textarea>
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
            <button v-if="spots.length>0" type="button" class="btn btn-success" v-on:click="addEmptySpotToBatch()">Добавить
            </button>
        </div>
        <div v-else class="spot-display">
            <h1>{{ river.title }}</h1>&nbsp;<a
                id="gpx" :href="gpxUrl(false)" style="padding-right:10px;" alt="Скачать GPX с точками порогов">GPX</a>&nbsp;<a
                id="gpx_en" :href="gpxUrl(true)" alt="Скачать GPX с точками порогов">GPX<sub>en</sub></a>
            <div style="float:right;">
                <img border='0' :src="informerUrl()">
                <div id="map" style="width:650px; height: 450px;padding-left: 30px;"></div>
            </div>
            <dl>
                <dt>Показывать на карте:</dt>
                <dd>
                    <span style="padding-left:40px;" v-if="river.visible">Да</span>
                    <span style="padding-left:40px;" v-else>Нет</span>&nbsp;&nbsp;&nbsp;
                    <button type="button" class="btn btn-info" v-if="canEdit() && pageMode != 'edit' && !river.visible" v-on:click="setVisible(true); hideError();">
                        Показывать на карте
                    </button>
                    <button type="button" class="btn btn-info" v-if="canEdit() && pageMode != 'edit' && river.visible" v-on:click="setVisible(false); hideError();">
                        Скрыть на карте
                    </button>
                    <div  v-if="canEdit()" style="padding-left:40px;" class="wwmap-system-hint">Нужно, когда мы не хотим выставлять наполовину размеченную и описанную реку.
                    Если добавляешь часть порогов, а остальные планируешь на потом, не делай реку видимой на карте.</div>
                    <div v-else></div>
                </dd>
                <dt v-if="river.region.id>0">Регион:</dt>
                <dd v-if="river.region.id>0">
                    <div style="padding-left:40px;">
                        <div v-if="river.region.fake">{{country.title}}</div>
                        <div v-else>{{river.region.title}}</div>
                    </div>
                </dd>
                <dt>Описание:</dt>
                <dd>`
                    <div style="padding-left:40px;">
                        {{river.description}}
                    </div>

                </dd>
                <dt>Другие варианты названия для автоматического поиска отчётов:</dt>
                <dd>
                    <ul>
                        <li v-for="alias in river.aliases">{{alias}}</li>
                    </ul>
                </dd>
                <dt>Отчёты:
                </dt>
                <dd>
                    <div style="padding-left:40px;" class="wwmap-system-hint" v-if="canEdit()">Поиск отчётов происходит раз в сутки ночью. Наберитесь терпения.</div>
                    <ul>
                        <li v-for="report in reports"><a target="_blank" :href="report.url">{{report.title}}</a></li>
                    </ul>
                </dd>
            </dl>
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
        -webkit-transform:
                translateY(145px)
                translateX(-15px)
                rotate(14deg);
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
    module.exports = {
        props: ['initialRiver', 'reports', 'country', 'region'],
        components: {
          FileUpload: VueUploadComponent
        },
        updated: function() {
            this.resetToInitialIfRequired();

            let t = this;
            if(this.$refs.uploadGpx && this.$refs.uploadGpx.value.length && this.$refs.uploadGpx.uploaded && this.gpxJustUploaded) {
                showRiverTree(this.river.region.country_id, nvlReturningId(this.river.region), this.river.id);
                this.gpxJustUploaded = false;
                for(let i=0; i< this.$refs.uploadGpx.value.length; i++) {
                    this.$refs.uploadGpx.value[i].response.forEach(function (x) {
                        t.spots.push(x);
                    });
                }
            }

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
                    onEnd: function(/**Event*/evt) {
                        if (t.spotIndexes.length <= evt.newIndex || t.spotIndexes.length <= evt.oldIndex || t.spotIndexes.length < t.spots.length) {
                            if (t.spotIndexes.length == 0) {
                                t.spotIndexes = t.spots.map((_, idx) => idx)
                            } else {
                                let maxEl = Math.max(...t.spotIndexes);
                                for (let i = t.spotIndexes.length, idx=0; i < t.spots.length; i++, idx++) {
                                    spotIndexes.push(maxEl + idx);
                                }
                            }
                        }
                        let el1 = t.spotIndexes[evt.oldIndex];
                        t.spotIndexes.splice(evt.oldIndex, 1);
                        t.spotIndexes.splice(evt.newIndex, 0, el1);

                        for (let i=0; i<t.spots.length;i++) {
                            let spotIdx = t.spotIndexes[i];
                            t.spots[spotIdx].order_index = "" + (i + 1);
                            t.spots[spotIdx].automatic_ordering = false;

                        }
                    }
                });
            }
        },
        created: function() {
            this.resetToInitialIfRequired();
        },
        computed: {
            uploadPath: function() { return backendApiBase + "/river/" + this.river.id +"/gpx"},
            headers:function(){
                return {
                    Authorization: getWwmapSessionId()
                }
            },
            pageMode: {
                get:function() {
                    return app.rivereditorstate.pageMode
                },

                set:function(newVal) {
                    app.rivereditorstate.pageMode = newVal
                }
            },
            meteoPoints: {
                get:function () {
                    return this.canEdit() ? getMeteoPoints() : [];
                }
            },
        },
        data:function() {
            return {

                river: null,
                previousRiverId: this.initialRiver.id,
                shouldReInit:function(){
                    return this.river==null || this.previousRiverId !== this.initialRiver.id && this.initialRiver.id > 0
                },
                resetToInitialIfRequired:function() {
                    if (this.shouldReInit()) {
                        this.previousRiverId = this.initialRiver.id;
                        this.river = this.initialRiver;
                        this.bounds = getRiverBounds(this.river.id);
                        this.center = [(this.bounds[0][0] + this.bounds[1][0]) / 2, (this.bounds[0][1] + this.bounds[1][1]) / 2];

                        let hideMap = emptyBounds(this.bounds);
                        if (this.map && hideMap) {
                            this.map.destroy();
                            this.map = null;
                        } else if (this.map) {
                            this.objectManager.setUrlTemplate(backendApiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z&river=' + this.river.id);
                            this.map.setBounds(this.bounds);
                        } else if (!this.map && !hideMap) {
                            this.showMap();
                        }

                        this.meteoPoint = this.getMeteoPointById(this.river.props.meteo_point);

                        this.prevRegionId = nvlReturningId(this.river.region);
                        this.prevRegionFake = this.river.region.fake;
                        this.prevCountryId = this.river.region.country_id;

                        var r = this.river;
                        this.activeSensor = this.sensors.filter(function(s){return s.id==r.props.vodinfo_sensor})[0]
                    }
                },

                // for editor
                userInfo: getAuthorizedUserInfoOrNull(),
                canEdit: function(){
                 return this.userInfo!=null && (this.userInfo.roles.includes("EDITOR") || this.userInfo.roles.includes("ADMIN"))
                },
                askForRemove: true,
                save:function() {
                    if (!this.river.title || !this.river.title.replace(/\s/g, '').length) {
                        this.showError("Нельзя сохранять реку без названия");
                        return false
                    }
                    updated = saveRiver(this.river);
                    if (updated) {
                        this.river = updated;
                        this.meteoPoint = this.getMeteoPointById(this.river.props.meteo_point);
                        this.pageMode='view';
                        this.hideError();
                        var new_region_id = nvlReturningId(updated.region);
                        setActiveEntity(updated.region.country_id, new_region_id, updated.id);
                        setActiveEntityState(updated.region.country_id, new_region_id, updated.id);
                        showCountrySubentities(updated.region.country_id);

                        if (new_region_id>0 && !updated.region.fake) {
                            showRegionTree(updated.region.country_id, new_region_id)
                        }
                        if (this.prevCountryId>0 && this.prevRegionId>0 && !this.prevRegionFake && this.prevRegionId!=new_region_id) {
                            showRegionTree(this.prevCountryId, this.prevRegionId)
                        } else if(this.prevCountryId>0 && this.prevCountryId!=updated.region.country_id) {
                            showCountrySubentities(this.prevCountryId)
                        }

                        this.prevRegionId = nvlReturningId(updated.region)
                        if (updated.region) {
                            this.prevRegionFake = updated.region.fake
                            this.prevCountryId = updated.region.country_id
                        }
                        this.reloadMap();
                        return true;
                    } else {
                        this.showError("Не удалось сохранить реку. Возможно, недостаточно прав");
                        return false;
                    }
                },
                saveSpotsBatch: function () {
                    try {
                        this.hideError();
                        let forDelete = this.spotsForDeleteIds;
                        let deleteErrors = "";
                        let saveErrors = "";

                        let removed = forDelete.map(function (id) {
                            let failMsg = "";
                            let httpStatusCode = 200;
                            let success = removeSpot(id, function (msg, httpCode) {
                                if(httpCode!=404) {
                                    failMsg = msg;
                                }
                                httpStatusCode = httpCode;
                            });
                            if (!success) {
                                deleteErrors += failMsg;
                            }
                            return httpStatusCode == 404 ||success;
                        });
                        let saved = this.spots.filter(function (spot) {
                            return !forDelete.includes(spot.id);
                        }).map(function(spot) {
                            let failMsg = "";
                            let success = saveSpot(spot, function (msg) {
                                failMsg = msg;
                            });
                            if (!success) {
                                saveErrors += failMsg;
                            }
                            return success;
                        });

                        let failCount = removed.concat(saved).filter(function (x) {
                            return !x;
                        }).length;

                        if (failCount != 0) {
                            if (saveErrors) {
                                this.showError("Не получилось сохранить: " + saveErrors);
                            } else if (deleteErrors) {
                                this.showError("Не получилось удалить: " + deleteErrors);
                            }
                            return false;
                        }
                    } catch (e) {
                        this.showError("Не получилось сохранить: " + e);
                        console.log(e);
                        return false
                    }


                    let countryId = this.river.region.country_id;
                    let regionId = nvlReturningId(this.river.region);
                    let riverId = this.river.id;

                    showRiverTree(countryId, regionId, riverId);

                    this.reload();
                    return true;
                },
                cancelEditing:function() {
                    if(this.river && this.river.id>0) {
                        this.reload();
                    } else {
                        this.closeEditorAndShowRiver();
                    }
                },
                reload: function () {
                    this.river = getRiver(this.river.id);
                    this.meteoPoint = this.getMeteoPointById(this.river.props.meteo_point);
                    this.spotsForDeleteIds = [];
                    this.hideError();
                    this.reloadMap();
                },
                setVisible: function(visible) {
                    this.river = setRiverVisible(this.river.id, visible);
                    this.meteoPoint = this.getMeteoPointById(this.river.props.meteo_point);
                    // set "visible" property to global storage to set icon in the left-side tree using reactivity of vue.js
                    var regionId = this.river.region.id;
                    if (this.river.region.fake) {
                        regionId = -1
                    }
                    getRiverFromTree(this.river.region.country_id, regionId, this.river.id).visible = this.river.visible;
                },
                remove: function() {
                    this.hideError()
                    if (!removeRiver(this.river.id)) {
                        this.showError("Can not delete")
                    } else {
                        this.closeEditorAndShowRiver();
                    }
                },
                closeEditorAndShowRiver: function() {
                    setActiveEntity(this.country.id, nvlReturningId(this.region));
                    setActiveEntityState(this.country.id, nvlReturningId(this.region));
                    app.rivereditorstate.visible = false;
                    if (this.river.region.fake || this.river.region.id == 0) {
                        showCountrySubentities(this.country.id);
                        selectCountry(this.country);
                    } else {
                        showRegionTree(this.country.id, nvlReturningId(this.river.region));
                        selectRegion(this.country, nvlReturningId(this.river.region));
                    }
                },
                showError: function(errMsg) {
                    app.errMsg = errMsg
                },
                hideError: function(errMsg) {
                    app.errMsg = null
                },
                // end of editor

                files: [],
                add_spot: function() {
                    app.spoteditorstate.visible = false;
                    app.rivereditorstate.visible = false;

                    app.spoteditorstate.visible = true;
                    app.spoteditorstate.editMode = true;
                    app.spoteditorstate.spot={
                        id: 0,
                        river: this.river,
                        order_index: "0",
                        automatic_ordering: true,
                        point: this.center,
                        aliases:[],
                        props:{},
                    };
                    app.spoteditorstate.country = this.country;
                    app.spoteditorstate.region = this.region;
                },

                regions: function() {
                    var regions = getAllRegions()
                    realRegions = regions.map(function(x){
                        if (x.fake) {
                            return {
                                id:x.id,
                                title: x.country.title
                            }
                        }
                        return {
                            id:x.id,
                            title: x.country.title + " - " + x.title
                        }
                    })
                    return realRegions
                },
                parseAliases:function(strVal) {
                    return strVal.split('\n').map(function(x) {return x.trim()}).filter(function(x){return x.length>0})
                },
                prevRegionId: 0,
                prevRegionFake: null,
                prevCountryId: 0,
                sensors: app.sensors,
                activeSensor: {id:null, title:null},
                informerUrl: function() {return this.river.props.vodinfo_sensor ? "http://gis.vodinfo.ru/informer/draw/v2_" + this.river.props.vodinfo_sensor + "_400_300_30_ffffff_110_8_7_H_none.png" : null},
                meteoPoint: null,
                meteoPointSelectMode: true,
                getMeteoPointById: function(id) {
                    if (id) {
                        for(let i=0;i<this.meteoPoints.length;i++) {
                            if (this.meteoPoints[i].id == id) {
                                return this.meteoPoints[i]
                            }
                        }
                    }
                    return {id: null, title: null, point: this.center}
                },
                addMeteoPoint: function () {
                    this.meteoPointSelectMode = false;
                },
                onCancelAddMeteoPoint: function () {
                    this.meteoPointSelectMode = true;
                },
                onAddMeteoPoint: function () {
                    this.onCancelAddMeteoPoint();
                    this.meteoPoint = addMeteoPoint(this.meteoPoint);
                    this.meteoPoints = getMeteoPoints();

                },
                getDefaultMap: function () {
                    let defaultMap = $.cookie("default_editor_map");
                    if (defaultMap && ymaps.mapType.storage.get(defaultMap)) {
                        return defaultMap
                    }
                    return "osm#standard"
                },
                showMap: function () {
                    if (emptyBounds(this.bounds)) {
                        return;
                    }
                    let t = this;
                    ymaps.ready(function () {
                        addMapLayers();

                        let mapType = t.getDefaultMap();
                        let map = new ymaps.Map("map", {
                            bounds: t.bounds,
                            type: mapType,
                            controls: ["zoomControl"]
                        });
                        map.controls.add(
                            new ymaps.control.TypeSelector([
                                    'osm#standard',
                                    'ggc#standard',
                                    'yandex#satellite',
                                    'google#satellite',
                                    'bing#satellite',
                                ]
                            )
                        );
                        registerMapSwitchLayersHotkeys(map);
                        var objectManager = new ymaps.RemoteObjectManager(backendApiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z&river=' + t.river.id, {
                            clusterHasBalloon: false,
                            geoObjectOpenBalloonOnClick: false,
                            geoObjectStrokeWidth: 3,
                            splitRequests: true
                        });
                        map.geoObjects.add(objectManager);
                        t.map = map;
                        t.objectManager = objectManager;
                    });
                },
                reloadMap: function() {
                    if (this.map) {
                        this.map.destroy();
                    }
                    this.showMap();
                },

                gpxJustUploaded: false,
                gpxUrl: function (transliterate) {
                    return `${backendApiBase}/downloads/river/${this.river.id}/gpx?tr=${transliterate}`;
                },

                spots: [],
                spotIndexes: [],
                spotsForDeleteIds: [],
                getSpots: getSpotsFull,
                all_categories:all_categories,
                spotPoint0: function(spot) {
                    if (Array.isArray(spot.point[0])) {
                        return spot.point[0]
                    } else {
                        return spot.point
                    }
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
            }
        },
        methods: {
            onSelectSensor: function(x) { this.river.props.vodinfo_sensor = x.id },
            onSelectMeteoPoint: function(x) { this.river.props.meteo_point = x.id }
        }
    }

</script>