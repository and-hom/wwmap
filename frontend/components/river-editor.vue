<template>
    <div>
        <ask id="del-river" title="Точно?"
             msg="Совсем удалить? Все пороги будут также удалены! Да, совсем! Восстановить будет никак нельзя!"
             :ok-fn="function() { remove(); }"></ask>

        <div v-if="canEdit()" class="btn-toolbar justify-content-between">
            <div class="btn-group mr-2" role="group">
                <button v-if="river.id && !editMode" type="button" class="btn btn-primary" v-on:click="add_spot()">Добавить препятствие</button>
                <button type="button" class="btn btn-info" v-if="!editMode" v-on:click="editMode=true; hideError();">
                    Редактирование
                </button>
                <button type="button" class="btn btn-success" v-if="editMode" v-on:click="editMode=!save()">Сохранить</button>
                <button type="button" class="btn btn-secondary" v-if="editMode" v-on:click="editMode=false; cancelEditing()">Отменить</button>
                <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-river">Удалить
                </button>
            </div>
            <div class="btn-group mr-2">
                <log-dropdown object-type="RIVER" :object-id="river.id"/>
            </div>
        </div>

        <div v-if="editMode" class="spot-editor-panel" style="padding-top:15px;">
                    <b-tabs>
                        <b-tab title="Главное" active>
                            <input v-model.trim="river.title" style="display:block"/>
                            <dl>
                                <dt>Показывать на карте:</dt>
                                <dd>
                                    <span style="padding-left:40px;" v-if="river.visible">Да</span>
                                    <span style="padding-left:40px;" v-else>Нет</span>&nbsp;&nbsp;&nbsp;
                                    <button type="button" class="btn btn-info" v-if="canEdit() && !editMode && !river.visible" v-on:click="setVisible(true); hideError();">
                                        Показывать на карте
                                    </button>
                                    <button type="button" class="btn btn-info" v-if="canEdit() && !editMode && river.visible" v-on:click="setVisible(false); hideError();">
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
        <div v-else class="spot-display">
            <h1>{{ river.title }}</h1>
            <img border='0' style="float:right;" :src="informerUrl()">
            <dl>
                <dt>Показывать на карте:</dt>
                <dd>
                    <span style="padding-left:40px;" v-if="river.visible">Да</span>
                    <span style="padding-left:40px;" v-else>Нет</span>&nbsp;&nbsp;&nbsp;
                    <button type="button" class="btn btn-info" v-if="canEdit() && !editMode && !river.visible" v-on:click="setVisible(true); hideError();">
                        Показывать на карте
                    </button>
                    <button type="button" class="btn btn-info" v-if="canEdit() && !editMode && river.visible" v-on:click="setVisible(false); hideError();">
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
                <dd>
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
        </div>

    </div>
</template>

<script>
    module.exports = {
        props: ['initialRiver', 'reports', 'country', 'region'],
        components: {
          FileUpload: VueUploadComponent
        },
        gpxJustUploaded: false,
        updated: function() {
            this.resetToInitialIfRequired();

            if(this.$refs.uploadGpx && this.$refs.uploadGpx.value.length && this.$refs.uploadGpx.uploaded && this.gpxJustUploaded) {
                showRiverTree(this.river.region.country_id, nvlReturningId(this.river.region), this.river.id);
                this.gpxJustUploaded = false;
            }
        },
        created: function() {
            this.resetToInitialIfRequired()
        },
        computed: {
            uploadPath: function() { return backendApiBase + "/river/" + this.river.id +"/gpx"},
            headers:function(){
                return {
                    Authorization: getWwmapSessionId()
                }
            },
            editMode: {
                get:function() {
                    return app.rivereditorstate.editMode
                },

                set:function(newVal) {
                    app.rivereditorstate.editMode = newVal
                }
            },
            meteoPoints: {
                get:function () {
                    console.log(this.canEdit())
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
                        this.center = getRiverCenter(this.river.id);
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
                        this.editMode=false;
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
                        return true;
                    } else {
                        this.showError("Не удалось сохранить реку. Возможно, недостаточно прав");
                        return false;
                    }
                },
                cancelEditing:function() {
                    if(this.river && this.river.id>0) {
                        this.reload();
                    } else {
                        this.closeEditorAndShowRiver();
                    }
                },
                reload:function() {
                    this.river = getRiver(this.river.id);
                    this.meteoPoint = this.getMeteoPointById(this.river.props.meteo_point);
                    this.hideError()
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
                    app.spoteditorstate.visible = false
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
                    }
                    app.spoteditorstate.country = this.country
                    app.spoteditorstate.region = this.region
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
                    console.log(this.center)
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

                }
            }
        },
        methods: {
            onSelectSensor: function(x) { this.river.props.vodinfo_sensor = x.id },
            onSelectMeteoPoint: function(x) { this.river.props.meteo_point = x.id }
        }
    }

</script>