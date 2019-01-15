<template>
    <div>
        <transition name="fade">
            <div class="alert alert-danger" role="alert" v-if="errMsg">
                {{errMsg}}
            </div>
        </transition>
        <ask id="del-river" title="Точно?"
             msg="Совсем удалить? Все пороги будут также удалены! Да, совсем! Восстановить будет никак нельзя!"
             :okfn="function() { remove(); }"></ask>

        <div v-if="canEdit()" class="btn-toolbar">
            <div v-if="river.id" class="btn-group mr-2" role="group">
                <button type="button" class="btn btn-primary" v-if="!editMode" v-on:click="add_spot()">Добавить препятствие</button>
            </div>
            <div class="btn-group mr-2" role="group" aria-label="First group">
                <button type="button" class="btn btn-info" v-if="!editMode" v-on:click="editMode=true; hideError();">
                    Редактирование
                </button>
                <button type="button" class="btn btn-success" v-if="editMode" v-on:click="editMode=false; save()">Сохранить</button>
                <button type="button" class="btn btn-secondary" v-if="editMode" v-on:click="editMode=false; reload()">Отменить</button>
            </div>
            <div class="btn-group">
                <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-river">Удалить
                </button>
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
                                             <template slot="selected-option" scope="option">
                                                     {{ option.id }}&nbsp;&dash;&nbsp;{{ option.title }}
                                             </template>
                                         </v-select>
                                    </div>
                                </div>
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
                                <button type="button" class="btn btn-success" v-if="!$refs.uploadGpx || !$refs.uploadGpx.active"
                                        @click.prevent="$refs.uploadGpx.active = true">
                                    <i class="fa fa-arrow-up" aria-hidden="true"></i>
                                    Начать загрузку
                                </button>
                                <button type="button" class="btn btn-danger" v-else
                                        @click.prevent="$refs.uploadGpx.active = false">
                                    <i class="fa fa-stop" aria-hidden="true"></i>
                                    Stop Upload
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
        updated: function() {
            this.resetToInitialIfRequired()

            if(this.$refs.uploadGpx && this.$refs.uploadGpx.value.length && this.$refs.uploadGpx.uploaded) {
                if (this.imagesOutOfDate) {
                    var t = this;
                    setTimeout(function() {
                        t.refresh()
                    }, 700);
                    this.imagesOutOfDate = false
                } else {
                    this.imagesOutOfDate = true
                }
            }
        },
        created: function() {
            this.resetToInitialIfRequired()
        },
        computed: {
                uploadPath: function() { return backendApiBase + "/river/" + this.river.id +"/gpx"},
        },
        data:function() {
            return {

                river: null,
                previousRiverId: this.initialRiver.id,
                shouldReInit:function(){
                    return this.river==null || this.previousRiverId != this.initialRiver.id && this.initialRiver.id > 0
                },
                resetToInitialIfRequired:function() {
                    if (this.shouldReInit()) {
                        this.previousRiverId = this.initialRiver.id;
                        this.river = this.initialRiver;

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
                editMode: app.rivereditorstate.editMode,
                errMsg:null,
                askForRemove: true,
                save:function() {
                    console.log(this.river.props)
                    updated = saveRiver(this.river)
                    if (updated) {
                        var prev = this.river
                        this.river = updated
                        this.editMode=false
                        this.hideError()
                        var new_region_id = nvlReturningId(updated.region)
                        setActiveEntity(updated.region.country_id, new_region_id, updated.id)
                        setActiveEntityState(updated.region.country_id, new_region_id, updated.id)
                        showCountrySubentities(updated.region.country_id)

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
                    } else {
                        this.showError("Не удалось сохранить реку. Возможно, недостаточно прав")
                    }
                },
                reload:function() {
                    this.river = getRiver(this.river.id)
                    this.hideError()
                },
                setVisible: function(visible) {
                    this.river = setRiverVisible(this.river.id, visible);
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
                        setActiveEntity(this.country.id, nvlReturningId(this.region))
                        setActiveEntityState(this.country.id, nvlReturningId(this.region))
                        if (this.river.region.fake) {
                            showCountrySubentities(this.country.id)
                        } else {
                            showRegionTree(this.country.id, nvlReturningId(this.region))
                        }
                        app.rivereditorstate.visible = false;
                    }
                },
                showError: function(errMsg) {
                    this.errMsg = errMsg
                },
                hideError: function(errMsg) {
                    this.errMsg = null
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
                        point:getRiverCenter(this.river.id),
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
                activeSensor: {id: 75402, title: "г.Звенигород [р. Москва]"},
                informerUrl: function() {return this.river.props.vodinfo_sensor ? "http://gis.vodinfo.ru/informer/draw/v2_" + this.river.props.vodinfo_sensor + "_400_300_30_ffffff_110_8_7_H_none.png" : null}
            }
        },
        methods: {
            onSelectSensor: function(x) { this.river.props.vodinfo_sensor = x.id }
        }
    }

</script>