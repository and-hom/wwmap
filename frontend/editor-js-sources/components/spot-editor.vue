<template>
    <div>
        <window-title v-bind:text="river.title + ' / ' + spot.title"></window-title>
        <div style="display:none;" :id="initialSpot.id"></div>
        <ask id="del-spot" title="Точно?" msg="Совсем удалить?" :ok-fn="function() { remove(); }"></ask>
        <ask id="mid-point-about" title="Опорные точки" :no-btn="false" ok-btn-title="Понятно"
             msg='Иногда река в месте порога изгибается. Чтобы контур вокруг порога повторял этот изгиб, нужно добавить несколько опорных точек в середине. Они не несут смысловой нагрузки и нужны только для отрисовки'
             :ok-fn="function() {}"></ask>

        <div v-if="canEdit" class="btn-toolbar justify-content-between">
            <div class="btn-group mr-2" role="group" aria-label="First group">
                <button type="button" class="btn btn-info" v-if="!editMode" v-on:click="editMode=true; hideError();">
                    Редактирование
                </button>
                <button type="button" class="btn btn-success" v-if="editMode" v-on:click="editMode = !save()">Сохранить</button>
                <button type="button" class="btn btn-secondary" v-if="editMode" v-on:click="editMode=!editMode; cancelEditing()">Отменить</button>
                <button type="button" class="btn btn-danger" v-if="spot.id>0" data-toggle="modal" data-target="#del-spot">Удалить
                </button>
            </div>
            <div class="btn-group mr-2">
                <log-dropdown object-type="SPOT" :object-id="spot.id"/>
            </div>
        </div>

        <div v-if="editMode" class="spot-editor-panel" style="padding-top:15px;">
            <b-tabs>
                <b-tab title="Главное" active>
                    <div class="container-fluid" style="margin-top: 20px;">
                        <div class="row">
                            <div class="col-5">
                                <div class="short-div">
                                    <strong>Название: </strong>
                                    <input v-model.trim="spot.title" style="display:block; width: 100%;"/>
                                </div>
                                <div class="short-div">
                                    <strong>Река: </strong>
                                    <v-select v-model="river" label="name" :filterable="false" :options="options"
                                              @search="onSearch">
                                        <template slot="no-options">
                                            Начните печатать название реки
                                        </template>
                                        <template slot="option" slot-scope="option">
                                            <div class="d-center">
                                                {{ option.title }}
                                            </div>
                                        </template>
                                        <template slot="selected-option" slot-scope="option">
                                            <div class="selected d-center">
                                                {{ option.title }}
                                            </div>
                                        </template>
                                    </v-select>
                                </div>
                            </div>
                            <div class='col-3' style="margin-left:0px">
                                <strong>Категория сложности: </strong><a target="_blank"
                                                                         href="https://huskytm.ru/rules2018-2019/#categories_tab"><img
                                    src="img/question_16.png"></a>
                                <dl style="padding-left:40px;">
                                    <dt>По классификатору</dt>
                                    <dd>
                                        <select v-model="spot.category">
                                            <option v-for="cat in all_categories" v-bind:value="cat.id">{{cat.title}}
                                            </option>
                                        </select>
                                    </dd>
                                    <dt>Низкий уровень воды</dt>
                                    <dd>
                                        <select v-model="spot.lw_category">
                                            <option v-for="cat in all_categories" v-bind:value="cat.id">{{cat.title}}
                                            </option>
                                        </select>
                                    </dd>
                                    <dt>Средний уровень воды</dt>
                                    <dd>
                                        <select v-model="spot.mw_category">
                                            <option v-for="cat in all_categories" v-bind:value="cat.id">{{cat.title}}
                                            </option>
                                        </select>
                                    </dd>
                                    <dt>Высокий уровень воды</dt>
                                    <dd>
                                        <select v-model="spot.hw_category">
                                            <option v-for="cat in all_categories" v-bind:value="cat.id">{{cat.title}}
                                            </option>
                                        </select>
                                    </dd>
                                </dl>
                            </div>
                            <div class="col-3">
                                <strong>Порядок следования:</strong><a href="#"><img src="img/question_16.png"></a>
                                <div><label for="auto-order">Автоматически</label>
                                    <input id="auto-order" type="checkbox" v-model="spot.automatic_ordering"></input>
                                </div>
                                <div class="wwmap-system-hint">Если препятствия не отсортированны, не будет
                                    сгенерированно
                                    PDF-описание.
                                </div>
                                <transition name="fade">
                                    <div v-if="spot.automatic_ordering" class="wwmap-system-hint">Автосортировка
                                        на основе треков из <a href="https://www.openstreetmap.org/">OSM</a> срабатывает
                                        ежесуточно ночью.
                                    </div>
                                    <div v-else><label for="auto-order" style="display:block;"><strong>Индекс
                                        препятствия</strong><br/>
                                        <div class="wwmap-system-hint">Чем меньше, тем выше по течению (раньше)</div>
                                    </label>
                                        <input v-model="spot.order_index" type="number"></input></div>
                                </transition>
                            </div>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col-2">
                            <strong style="display: block; maring-bottom: 20px;">Расположение: </strong><br/>
                            <div v-if="editEndPoint()">
                                <a href="javascript:void(0);" v-on:click.stop="removeSpotEndPoint()" style="font-size: 80%;">Убрать конечную точку</a><br/>
                                <a href="javascript:void(0);" v-on:click.stop="addSpotAnchorPoint()" style="font-size: 80%;">Добавить опорную точку в середину</a><a target="_blank"
                                    href="javascript:void(0)" data-toggle="modal" data-target="#mid-point-about"><img src="img/question_16.png"></a>
                            </div>
                            <a v-else href="javascript:void(0);" v-on:click.stop="addSpotEndPoint()" style="font-size: 80%;">Препятствие протяжённое. Добавить конечную точку</a>
                        </div>
                        <div class="col-10">
                            <ya-map-location ref="locationEdit" v-bind:spot="spot" width="100%" height="600px" :editable="true" :ya-search="true" v-bind:refresh-on-change="spot.point"/>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col-2"><strong>Описание: </strong></div>
                        <div class="col-10"><textarea rows="10" cols="120" v-model="spot.short_description"></textarea></div>
                    </div>
                    <div class="row">
                        <div class="col-2"><strong>Ориентиры:</strong></div>
                        <div class="col-10"><textarea rows="10" style="width:100%" v-model="spot.orient"></textarea></div>
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
                        <div class="col-10"><textarea rows="10" cols="120" v-model="spot.lw_description"></textarea></div>
                    </div>
                    <div class="row">
                        <div class="col-2"><strong>Описание для среднего уровня воды:</strong></div>
                        <div class="col-10"><textarea rows="10" cols="120" v-model="spot.mw_description"></textarea></div>
                    </div>
                    <div class="row">
                        <div class="col-2"><strong>Описание для высокого уровня воды:</strong></div>
                        <div class="col-10"><textarea rows="10" cols="120" v-model="spot.hw_description"></textarea></div>
                    </div>
                    <hr/>
                    <div class="row">
                        <div class="col-12"><strong>Другие варианты названия для поиска изображений:</strong>
                            <div class="wwmap-system-hint" style="margin-bottom: 7px;">Каждое альтернативное название на новой строке</div>
                            <textarea v-if="editMode" v-bind:text-content="spot.aliases"
                                                                      v-on:input="spot.aliases = parseAliases($event.target.value)"
                                                                      rows="10" cols="120">{{ spot.aliases.join('\n') }}</textarea>
                                                                      </div>
                    </div>
                </b-tab>
                <b-tab title="Схемы" :disabled="spot.id<=0">
                    <img-upload ref="schemasList" :spot="spot" v-bind:images="schemas" type="schema" :auth="true"></img-upload>
                </b-tab>
                <b-tab   title="Фото" :disabled="spot.id<=0">
                    <img-upload ref="imagesList" :spot="spot" v-bind:images="images" type="image" :auth="true"></img-upload>
                </b-tab>
                <b-tab title="Видео" :disabled="spot.id<=0">
                    <video-add :spot="spot" v-bind:images="videos" type="video" :auth="true"></video-add>
                </b-tab>
                <b-tab title="Системные параметры">
                    <span class="wwmap-system-hint" style="padding-top: 10px;">Тут собраны настройки разных системных вещей для каждого порога в отдельности</span>
                    <props v-if="spot.props" :p="spot.props"/>
                </b-tab>
            </b-tabs>
        </div>
        <div v-else class="spot-display">
            <h1>{{ spot.title }}</h1>
            <div style="float:right; width:400px; margin-left: 20px;">
                <img v-if="spotMainUrl" :src="spotMainUrl" style="width:100%; cursor: pointer;"
                     @click="imgIndex = mainImageIndex; schIndex = mainSchemaIndex" class="wwmap-desc-section-div"/>
                <img v-else src="img/no-photo.png" style="width:100%" class="wwmap-desc-section-div"/>
                <div v-if="spot.aliases && spot.aliases.length > 0 && canEdit" class="wwmap-desc-section-div">
                    <strong>Другие варианты названия для поиска картинок роботом:</strong>
                    <ul style="padding-left: 0;">
                        <li v-for="alias in spot.aliases">{{alias}}</li>
                    </ul>
                </div>
                <div v-if="canEdit" class="wwmap-desc-section-div">
                    <div class="wwmap-system-hint">Эти параметры предназначены для определения порядка следования порогов. Автоматическое упорядочивание проходит раз в сутки ночью.</div>
                    <div><strong>Порядок следования:</strong> {{ spot.order_index }}</div>
                    <div><strong>Автоматическое упорядочивание:</strong>&nbsp;<span v-if="spot.automatic_ordering">Да</span><span v-else>Нет</span></div>
                    <div><strong>В последний раз автоматическое упорядочивание срабатывало:</strong> {{ lastAutoOrdering() }}</div>
                </div>
            </div>
            <div class="wwmap-desc-section-div">
                <div v-if="spot.lw_category!=='0' ||spot.mw_category!=='0' || spot.hw_category!=='0'"><strong>К.с. нв/св/вв:</strong>&nbsp;<category :category="spot.lw_category"></category>/<category :category="spot.mw_category"></category>/<category :category="spot.hw_category"></category></div>
                <div><strong>К.с. по классификатору:</strong>&nbsp;<category :category="spot.category"></category></div>
            </div>
            <div class="wwmap-desc-section-div" v-if="spot.short_description">
                {{ spot.short_description }}
            </div>
            <div class="wwmap-desc-section-div">
                <ya-map-location ref="locationView" v-bind:spot="spot" width="70%" height="600px" :editable="false" :zoom="15"></ya-map-location>
                <div style="padding-top:4px;">
                <strong>Широта:</strong>&nbsp;{{ spotPoint0()[0] }}&nbsp;&nbsp;&nbsp;<strong>Долгота:</strong>&nbsp;{{ spotPoint0()[1] }}
                </div>
            </div>
            <div v-if="spot.orient" class="wwmap-desc-section-div">
                <strong>Ориентиры:</strong><br/>
                {{ spot.orient }}
            </div>
            <div class="container-fluid border-inside" style="padding-bottom:15px;"
                 v-if="spot.lw_description || spot.mw_description || spot.hw_description">
                <div class="row">
                    <div class="col-4"><strong>Уровень воды</strong></div>
                    <div class="col-8"><strong>Тех. описание</strong></div>
                </div>
                <div class="row">
                    <div class="col-4"><strong>Низкая вода</strong></div>
                    <div class="col-8">{{ spot.lw_description }}</div>
                </div>
                <div class="row">
                    <div class="col-4"><strong>Средняя вода</strong></div>
                    <div class="col-8">{{ spot.mw_description }}</div>
                </div>
                <div class="row">
                    <div class="col-4"><strong>Высокая вода</strong></div>
                    <div class="col-8">{{ spot.hw_description }}</div>
                </div>
            </div>

            <div v-if="spot.approach" class="wwmap-desc-section-div">
                <strong>Подход/выход</strong>
                {{ spot.approach }}
            </div>
            <div v-if="spot.safety" class="wwmap-desc-section-div">
                <strong>Страховка</strong>
                {{ spot.safety }}
            </div>
            <div v-if="schemas.length">
                <h2>Схемы</h2>
                <div>
                    <gallery id="schemas-gallery" :images="schemas.map(function(x) {return x.url})" :index="schIndex" @close="schIndex = null"></gallery>
                    <div
                            class="image wwmap-gallery-cell"
                            v-for="schema, schemaIndex in schemas"
                            @click="schIndex = schemaIndex"
                            :style="{ backgroundImage: 'url(' + schema.preview_url + ')', width: '300px', height: '200px', cursor: 'pointer' }"
                    ></div>
                </div>
            </div>
            <div v-if="images.length">
                <h2>Фото галерея</h2>
                <div>
                    <gallery id="image-gallery" :images="images.map(function(x) {return x.url})" :index="imgIndex" @close="imgIndex = null"></gallery>
                    <div
                            class="image wwmap-gallery-cell"
                            v-for="image, imageIndex in images"
                            @click="imgIndex = imageIndex"
                            :style="{ backgroundImage: 'url(' + image.preview_url + ')', width: '300px', height: '200px', cursor: 'pointer' }"
                    ></div>
                </div>
            </div>
            <div v-if="videos.length">
                <h2>Видео</h2>
                <div>
                    <iframe width="450" height="300"
                            v-for="image in videos"
                            :src="embeddedVideoUrl(image.remote_id)"
                            frameborder="0"
                            allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
                            allowfullscreen
                    style="margin-right: 2px; margin-bottom: 2px;"></iframe>
                </div>
            </div>
        </div>
    </div>
</template>
<style type="text/css">
    .wwmap-desc-section-div {
        margin-bottom: 10px;
    }
    .btn-toolbar {
        margin-bottom: 5px;
    }
</style>

<script>
    import {
        all_categories,
        setActiveEntityUrlHash,
        getImages,
        getRiver,
        getSpotMainImageUrl,
        nvlReturningId,
        removeSpot,
        saveSpot
    } from '../editor'
    import {store} from "../main";
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from "../auth";
    import {backendApiBase} from '../config'

    module.exports = {
        props: ['initialSpot', 'country', 'region'],

        methods: {
            onSearch: function(search, loading) {
                loading(true);
                var component = this
                fetch(
                    backendApiBase + '/river?q=' + search
                ).then(function(res) {
                    res.json().then(function(json){
                        component.options = json
                    });
                    loading(false);
                });
            },
        },
        beforeUpdate: function() {
        },
        created: function() {
            this.resetToInitialIfRequired()
        },
        updated: function() {
            var shouldRefreshChildren = this.shouldReInit();
            this.resetToInitialIfRequired();

            if (shouldRefreshChildren) {
                // It's an ugly hack. I really do not know, why child's update lifecycle handler method is not called on spot changed
                // Maybe computed properties are not allowed in v-bind ?

                if (this.editMode) {
                    this.$refs.locationEdit.spot = this.spot;
                    this.$refs.locationEdit.doUpdate();
                    this.$refs.schemasList.spot = this.spot;
                    this.$refs.schemasList.refresh();
                    this.$refs.imagesList.spot = this.spot;
                    this.$refs.imagesList.refresh();
                } else {
                    this.$refs.locationView.spot = this.spot;
                    this.$refs.locationView.doUpdate();
                }
                // end of hack
            }
        },
        computed: {
                images: {
                    get:function() {
                        return store.state.spoteditorstate.images
                    },

                    set:function(newVal) {
                        store.commit('setSpotImages', newVal);
                    }
                },
                schemas: {
                    get:function() {
                        return store.state.spoteditorstate.schemas
                    },

                    set:function(newVal) {
                        store.commit('setSpotSchemas', newVal);
                    }
                },
                videos: {
                    get:function() {
                        return store.state.spoteditorstate.videos
                    },

                    set:function(newVal) {
                        store.commit('setSpotVideos', newVal);
                    }
                },
                mainImageIndex: function() {
                    var imgIdx = this.findMainImgIdx(this.images, this.spotMainUrl)
                    if (imgIdx>-1) {
                        return imgIdx
                    }
                    return null
                },
                mainSchemaIndex: function() {
                    var imgIdx = this.findMainImgIdx(this.schemas, this.spotMainUrl)
                    if (imgIdx>-1) {
                        return imgIdx
                    }
                    return null
                },
                editMode: {
                    get:function() {
                        return store.state.spoteditorstate.editMode
                    },

                    set:function(newVal) {
                        store.commit("setSpotEditorEditMode", newVal);
                    }
                },
        },
        data:function() {
            return {
                // for editor
                map: null,
                label: null,
                canEdit: false,
                askForRemove: false,
                save:function() {
                    if (!this.spot.title || !this.spot.title.replace(/\s/g, '').length) {
                        this.showError("Нельзя сохранять порог без названия");
                        return false
                    }

                    let oldRiver = this.spot.river;
                    let riverChanged = this.river.id != oldRiver.id;
                    this.spot.river = this.river;

                    saveSpot(this.spot).then(updated => {
                        if (riverChanged) {
                            console.log("Move spot between rivers:" + oldRiver.id + "->" + updated.river.id);
                            hideRiverSubentities(oldRiver.region.country_id, nvlReturningId(oldRiver.region), oldRiver.id)
                        }
                        this.spot = updated;
                        this.river =  updated.river;
                        this.editMode=false;
                        this.reloadMainImg();
                        this.reloadImgs();
                        this.hideError();

                        let countryId = this.spot.river.region.country_id;
                        let regionId = nvlReturningId(this.spot.river.region);
                        let riverId = this.spot.river.id;

                        setActiveEntityUrlHash(countryId, regionId, riverId, updated.id);
                        store.commit('setActiveEntityState', {
                            contryId: countryId,
                            regionId: regionId,
                            riverId: riverId,
                            spotId: updated.id,
                        });
                        store.commit('showRiverSubentities', {
                            countryId: countryId,
                            regionId: regionId,
                            riverId: riverId
                        });
                    }, err => this.showError("Не удалось сохранить препятствие. Возможно, недостаточно прав"));
                },
                cancelEditing:function() {
                  if(this.spot && this.spot.id>0) {
                      this.reload();
                  } else {
                      this.closeEditorAndShowRiver();
                  }
                },
                reload:function() {
                    getSpot(this.spot.id).then(spot => {
                        this.spot = spot;
                        this.river = this.spot.river;
                        this.reloadMainImg();
                        this.reloadImgs();
                        this.hideError();
                    });
                },
                reloadMainImg: function() {
                    this.spotMainUrl = null;
                    getSpotMainImageUrl(this.spot.id).then(url => this.spotMainUrl = url);
                },
                reloadImgs: function() {
                    this.images = [];
                    this.schemas = [];
                    this.videos = [];

                    this.imgIndex = null;
                    this.schIndex = null;
                    this.vidIndex = null;

                    getImages(this.initialSpot.id, "image").then(images => this.images = images);
                    getImages(this.initialSpot.id, "schema").then(schemas => this.schemas = schemas);
                    getImages(this.initialSpot.id, "video").then(videos => this.videos = videos);
                },
                remove: function() {
                    this.hideError();
                    removeSpot(this.spot.id).then(
                        ok => this.closeEditorAndShowRiver(),
                        err => this.showError("Не могу удалить: "+err));
                },
                closeEditorAndShowRiver: function() {
                    let countryId = this.river.region.country_id;
                    let regionId = nvlReturningId(this.river.region);
                    let riverId = this.river.id;
                    setActiveEntityUrlHash(countryId, regionId, riverId);
                    store.commit('setActiveEntityState', {
                        contryId: countryId,
                        regionId: regionId,
                        riverId: riverId,
                    });
                    store.commit('showRiverSubentities', {
                        countryId: countryId,
                        regionId: regionId,
                        riverId: riverId
                    });
                    store.commit('hideAll');
                    store.commit('selectRiver', {
                        country: {id: this.river.region.country_id},
                        region: this.river.region,
                        riverId: this.river.id
                    });
                },
                showError: function(errMsg) {
                    store.commit("setErrMsg", errMsg);
                },
                hideError: function() {
                    store.commit("setErrMsg", null);
                },
                // end of editor
                all_categories:all_categories,

                options: [],

                // imgs
                imgIndex: null,
                schIndex: null,
                vidIndex: null,
                spotMainUrl: null,

                spot: null,
                river: null,
                editEndPoint: function () {
                    return Array.isArray(this.spot.point[0])
                },
                endPointBackup: null,
                addSpotEndPoint: function () {
                    if (!Array.isArray(this.spot.point[0])) {
                        if (this.endPointBackup) {
                            this.spot.point = [this.spot.point, this.endPointBackup];
                        } else {
                            this.spot.point = [this.spot.point, this.spot.point];
                        }
                    }
                },
                addSpotAnchorPoint: function () {
                    if (Array.isArray(this.spot.point[0])) {
                        let l = this.spot.point.length;
                        let p1 = this.spot.point[l - 2];
                        let p2 = this.spot.point[l - 1];
                        let p = [(p1[0] + p2[0]) / 2, (p1[1] + p2[1]) / 2];
                        this.spot.point.splice(l - 1, 0, p);
                    }
                },
                removeSpotEndPoint: function () {
                    if (Array.isArray(this.spot.point[0])) {
                        this.endPointBackup = this.spot.point[this.spot.point.length - 1];
                        this.spot.point = this.spot.point[0];
                    }
                },
                spotPoint0: function() {
                    if (Array.isArray(this.spot.point[0])) {
                        return this.spot.point[0]
                    } else {
                        return this.spot.point
                    }
                },
                shouldReInit:function(){
                    return this.spot==null || this.previousSpotId !== this.initialSpot.id && this.initialSpot.id > 0
                },
                resetToInitialIfRequired:function() {
                    if (this.shouldReInit()) {
                        hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
                        getRiver(this.initialSpot.river.id).then(river => this.options = [river]);

                        this.previousSpotId = this.initialSpot.id;
                        this.spot = this.initialSpot;
                        this.river = this.initialSpot.river;

                        this.reloadMainImg();
                        this.reloadImgs();
                    }
                },

                spotMainUrlCached: null,
                previousSpotId: this.initialSpot.id,
                parseAliases:function(strVal) {
                    return strVal.split('\n').map(function(x) {return x.trim()}).filter(function(x){return x.length>0})
                },
                lastAutoOrdering:function() {
                    var lastOrderingDate = new Date(this.spot.last_automatic_ordering)
                    if (lastOrderingDate.getFullYear()<=2017) {
                        return 'Никогда'
                    }
                    return this.spot.last_automatic_ordering
                },
                findMainImgIdx: function (imgs, spotMainUrl) {
                    return imgs.findIndex(function (el) {
                        if (el.preview_url == spotMainUrl) {
                            return true
                        }
                        return false
                    })
                },
                embeddedVideoUrl: function (id) {
                    return "https://www.youtube.com/embed/" + id
                },
                schemaIndex: null,
                imageIndex: null,
        i: [
          'https://dummyimage.com/800/ffffff/000000',
        ],
        index: null
            }
        }
    }

</script>