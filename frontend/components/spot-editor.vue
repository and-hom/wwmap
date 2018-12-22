<template>
    <div>
        <div style="display:none;" :id="initialSpot.id"></div>
        <transition name="fade">
            <div class="alert alert-danger" role="alert" v-if="errMsg">
                {{errMsg}}
            </div>
        </transition>
        <ask id="del-spot" title="Точно?" msg="Совсем удалить?" :okfn="function() { remove(); }"></ask>

        <div v-if="canEdit()" class="btn-toolbar">
            <div class="btn-group mr-2" role="group" aria-label="First group">
                <button type="button" class="btn btn-info" v-if="!editMode" v-on:click="editMode=true; hideError();">
                    Редактирование
                </button>
                <button type="button" class="btn btn-success" v-if="editMode" v-on:click="save()">Сохранить</button>
                <button type="button" class="btn btn-secondary" v-if="editMode" v-on:click="editMode=!editMode; reload()">Отменить</button>
            </div>
            <div class="btn-group">
                <button type="button" class="btn btn-danger" v-if="spot.id>0" data-toggle="modal" data-target="#del-spot">Удалить
                </button>
            </div>
        </div>

        <div v-if="editMode" class="spot-editor-panel" style="padding-top:15px;">
            <b-tabs>
                <b-tab title="Главное" active>
                    <div class="container-fluid" style="margin-top: 20px;">
                        <div class="row">
                            <div class="col-6">
                                <div class="short-div">
                                    <strong>Название: </strong>
                                    <input v-model.trim="spot.title" style="display:block; width: 100%;"/>
                                </div>
                                <div class="short-div">
                                    <strong>Река: </strong>
                                    <v-select v-model="spot.river" label="name" :filterable="false" :options="options"
                                              @search="onSearch">
                                        <template slot="no-options">
                                            Начните печатать название реки
                                        </template>
                                        <template slot="option" slot-scope="option">
                                            <div class="d-center">
                                                {{ option.title }}
                                            </div>
                                        </template>
                                        <template slot="selected-option" scope="option">
                                            <div class="selected d-center">
                                                {{ option.title }}
                                            </div>
                                        </template>
                                    </v-select>
                                </div>
                                <div class='row' style="margin-left:0px">
                                    <div class="col-7">
                                        <strong>Категория сложности: </strong><a target="_blank"
                                                                                 href="https://huskytm.ru/rules2018-2019/#categories_tab"><img
                                            src="img/question_16.png"></a>
                                        <dl style="padding-left:40px;">
                                            <dt>По классификатору</dt>
                                            <dd>
                                                <select v-model="spot.category">
                                                    <option v-for="cat in all_categories" v-bind:value="cat.id">{{cat.title}}</option>
                                                </select>
                                            </dd>
                                            <dt>Низкий уровень воды</dt>
                                            <dd>
                                                <select v-model="spot.lw_category">
                                                    <option v-for="cat in all_categories" v-bind:value="cat.id">{{cat.title}}</option>
                                                </select>
                                            </dd>
                                            <dt>Средний уровень воды</dt>
                                            <dd>
                                                <select v-model="spot.mw_category">
                                                    <option v-for="cat in all_categories" v-bind:value="cat.id">{{cat.title}}</option>
                                                </select>
                                            </dd>
                                            <dt>Высокий уровень воды</dt>
                                            <dd>
                                                <select v-model="spot.hw_category">
                                                    <option v-for="cat in all_categories" v-bind:value="cat.id">{{cat.title}}</option>
                                                </select>
                                            </dd>
                                        </dl>
                                    </div>
                                    <div class="col-5">
                                        <strong>Порядок следования:</strong><a href="#"><img src="img/question_16.png"></a>
                                        <div><label for="auto-order">Автоматически</label>
                                        <input id="auto-order" type="checkbox" v-model="spot.automatic_ordering"></input></div>
                                        <div class="wwmap-system-hint">Если препятствия не отсортированны, не будет сгенерированно
                                                                                    PDF-описание.</div>
                                        <transition name="fade">
                                            <div v-if="spot.automatic_ordering" class="wwmap-system-hint">Автосортировка
                                            на основе треков из <a href="https://www.openstreetmap.org/">OSM</a> срабатывает ежесуточно ночью.</div>
                                            <div v-else><label for="auto-order" style="display:block;"><strong>Индекс препятствия</strong><br/>
                                            <div class="wwmap-system-hint">Чем меньше, тем выше по течению (раньше)</div></label>
                                            <input v-model="spot.order_index" type="number"></input></div>
                                        </transition>
                                    </div>
                                </div>
                            </div>
                            <div class="col-6">
                                <strong>Расположение: </strong>
                                <ya-map-location ref="locationEdit" v-bind:spot="spot" width="100%" :editable="true"></ya-map-location>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-6">
                            </div>
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
                        <div class="col-12"><strong>Другие варианты названия для поиска изображений:</strong><textarea v-if="editMode" v-bind:text-content="spot.aliases"
                                                                      v-on:input="spot.aliases = parseAliases($event.target.value)"
                                                                      rows="10" cols="120">{{ spot.aliases.join('\n') }}</textarea>
                                                                      </div>
                    </div>
                </b-tab>
                <b-tab title="Схемы" :disabled="spot.id>0 ? false : true">
                    <img-upload ref="schemasList" :spot="spot" v-bind:images="schemas" type="schema" :auth="true"></img-upload>
                </b-tab>
                <b-tab title="Фото" :disabled="spot.id>0 ? false : true">
                    <img-upload ref="imagesList" :spot="spot" v-bind:images="images" type="image" :auth="true"></img-upload>
                </b-tab>
                <b-tab title="Видео" disabled>
                </b-tab>
                <b-tab title="Системные параметры">
                    <span class="wwmap-system-hint" style="padding-top: 10px;">Тут собраны настройки разных системных вещей для каждого порога в отдельности</span>
                    <props :p="spot.props"/>
                </b-tab>
            </b-tabs>
        </div>
        <div v-else class="spot-display">
            <div class="container-fluid" style="margin-top: 20px;">
                <div class="row">
                    <div class="col-7">
                        <div class="short-div">
                            <h1>{{ spot.title }}</h1>
                        </div>
                        <div class="short-div">
                            {{ spot.short_description }}
                        </div>
                        <div class="short-div">
                            <strong>Ориентиры:</strong><br/>
                            {{ spot.orient }}
                        </div>
                    </div>
                    <div class="col-5">
                        <img v-if="spotMainUrl" :src="spotMainUrl" style="width:100%; cursor: pointer;"
                        @click="imgIndex = mainImageIndex; schIndex = mainSchemaIndex"/>
                        <img v-else src="img/no-photo.png" style="width:100%"/>
                    </div>
                </div>
                <div class="row">
                    <div class="col-7">
                        <div class="container-fluid border-inside" style="padding-bottom:15px;">
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

                        <div v:if="spot.approach" class="short-div">
                            <strong>Подход/выход</strong>
                            <td colspan="2">{{ spot.approach }}</td>
                        </div>
                        <div v:if="spot.safety" class="short-div">
                            <strong>Страховка</strong>
                            <td colspan="2">{{ spot.safety }}</td>
                        </div>
                    </div>
                    <div class="col-5">
                        <strong>Расположение: </strong>
                        <ya-map-location ref="locationView" v-bind:spot="spot" width="100%" :editable="false"></ya-map-location>
                        <div class="short-div">
                            <strong>Координаты:</strong><br/><strong>Широта:</strong>&nbsp;{{ spot.point[0] }}
                            <br/><strong>Долгота:</strong>&nbsp;{{ spot.point[1] }}
                            <br/>
                            <div><strong>К.с. нв/св/вв:</strong>&nbsp;<category :category="spot.lw_category"></category>/<category :category="spot.mw_category"></category>/<category :category="spot.hw_category"></category></div>
                            <div><strong>К.с. по классификатору:</strong>&nbsp;<category :category="spot.category"></category></div>
                        </div>
                    </div>
                </div>
                <div class="row">
                    <div class="col-6">
                        <strong>Другие варианты названия для поиска картинок роботом:</strong>
                        <ul>
                            <li v-for="alias in spot.aliases">{{alias}}</li>
                        </ul>
                    </div>
                    <div class="col-6">
                        <div v-if="canEdit()">
                            <div style="font-size:60%;">Эти параметры предназначены для определения порядка следования порогов.</div>
                            <div><strong>Порядок следования:</strong> {{ spot.order_index }}</div>
                            <div><strong>Автоматическое упорядочивание:</strong>&nbsp;<span v-if="spot.automatic_ordering">Да</span><span v-else>Нет</span></div>
                            <div><strong>В последний раз автоматическое упорядочивание срабатывало:</strong> {{ lastAutoOrdering() }}</div>
                        </div>
                    </div>
                </div>
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
        </div>
    </div>
</template>

<script>
    module.exports = {
        props: ['initialSpot', 'country', 'region'],

        methods: {
            onSearch: function(search, loading) {
                loading(true);
                var component = this
                fetch(
                  `${backendApiBase}/river?q=${search}`
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
            shouldRefreshChildren = this.shouldReInit()
            this.resetToInitialIfRequired()

            if (shouldRefreshChildren) {
                // It's an ugly hack. I really do not know, why child's update lifecycle handler method is not called on spot changed
                // Maybe computed properties are not allowed in v-bind ?

                if (this.editMode) {
                    this.$refs.locationEdit.spot = this.spot
                    this.$refs.locationEdit.doUpdate()
                    this.$refs.schemasList.spot = this.spot
                    this.$refs.schemasList.refresh()
                    this.$refs.imagesList.spot = this.spot
                    this.$refs.imagesList.refresh()
                } else {
                    this.$refs.locationView.spot = this.spot
                    this.$refs.locationView.doUpdate()
                }
                // end of hack
            }
        },
        computed: {
                spotMainUrl: {
                    get:function() {
                        if (!this.spotMainUrlCached) {
                            this.spotMainUrlCached = getSpotMainImageUrl(this.initialSpot.id)
                        }
                        return this.spotMainUrlCached
                    },

                    set:function(newVal) {
                        this.spotMainUrlCached = newVal
                    },
                },
                images: {
                    get:function() {
                        return app.spoteditorstate.images
                    },

                    set:function(newVal) {
                        app.spoteditorstate.images = newVal
                    },
                },
                schemas: {
                    get:function() {
                        return app.spoteditorstate.schemas
                    },

                    set:function(newVal) {
                        app.spoteditorstate.schemas = newVal
                    },
                },
                mainImageIndex: function() {
                    var imgIdx = this.findMainImgIdx(this.images, this.spotMainUrl)
                    if (imgIdx>-1) {
                        return imgIdx
                    }
                    console.log('img-null')
                    return null
                },
                mainSchemaIndex: function() {
                    imgIdx = this.findMainImgIdx(this.schemas, this.spotMainUrl)
                    if (imgIdx>-1) {
                        return imgIdx
                    }
                    console.log('sch-null')
                    return null
                },
        },
        data:function() {
            return {
                // for editor
                userInfo: getAuthorizedUserInfoOrNull(),
                map: null,
                label: null,
                canEdit: function(){
                 return this.userInfo!=null && (this.userInfo.roles.includes("EDITOR") || this.userInfo.roles.includes("ADMIN"))
                },
                editMode: app.spoteditorstate.editMode,
                errMsg:null,
                askForRemove: false,
                save:function() {
                    updated = saveSpot(this.spot)
                    if (updated) {
                        this.spot = updated
                        this.editMode=false
                        this.reloadMainImg()
                        this.reloadImgs()
                        this.hideError()

                        setActiveEntity(this.country.id, nvlReturningId(this.region), this.spot.river.id, updated.id)
                        setActiveEntityState(this.country.id, nvlReturningId(this.region), this.spot.river.id, updated.id)
                        showRiverTree(this.country.id, nvlReturningId(this.region), this.spot.river.id)
                    } else {
                        this.showError("Не удалось сохранить препятствие. Возможно, недостаточно прав")
                    }
                },
                reload:function() {
                    this.spot = getSpot(this.spot.id)
                    this.reloadMainImg()
                    this.reloadImgs()
                    this.hideError()
                },
                reloadMainImg: function() {
                    this.spotMainUrl = getSpotMainImageUrl(this.spot.id)
                },
                reloadImgs: function() {
                    this.images = getImages(this.initialSpot.id, "image")
                    this.imgIndex = null
                    this.schemas = getImages(this.initialSpot.id, "schema")
                    this.schIndex = null
                },
                remove: function() {
                    this.hideError()
                    if (!removeSpot(this.spot.id)) {
                        this.showError("Can not delete")
                    } else {
                        setActiveEntity(this.country.id, nvlReturningId(this.region), this.spot.river.id)
                        setActiveEntityState(this.country.id, nvlReturningId(this.region), this.spot.river.id)
                        showRiverTree(this.country.id, nvlReturningId(this.region), this.spot.river.id)
                        app.spoteditorstate.visible = false;
                    }
                },
                showError: function(errMsg) {
                    this.errMsg = errMsg
                },
                hideError: function(errMsg) {
                    this.errMsg = null
                },
                // end of editor
                all_categories:all_categories,

                options: [getRiver(this.initialSpot.river.id)],

                // imgs
                imgIndex: null,
                schIndex: null,

                spot: null,
                shouldReInit:function(){
                    return this.spot==null || this.previousSpotId != this.initialSpot.id && this.initialSpot.id > 0
                },
                resetToInitialIfRequired:function() {
                    if (this.shouldReInit()) {
                        this.previousSpotId = this.initialSpot.id
                        this.spot = this.initialSpot
                        this.spotMainUrl = getSpotMainImageUrl(this.initialSpot.id)
                        this.images = getImages(this.initialSpot.id, "image")
                        this.schemas = getImages(this.initialSpot.id, "schema")
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
                findMainImgIdx:function(imgs, spotMainUrl) {
                    return imgs.findIndex(function(el) {
                                                   if (el.preview_url == spotMainUrl) {
                                                       return true
                                                   }
                                                   return false
                                               })
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