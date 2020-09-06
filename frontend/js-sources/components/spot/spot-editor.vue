<template>
    <div>
        <btn-bar ref="btnBar" logObjectType="SPOT" :logObjectId="spot.id">
            <button type="button" class="btn btn-success" v-on:click="save()">Сохранить</button>
            <slot></slot>
        </btn-bar>
        <div class="spot-editor-panel" style="padding-top:15px;">
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
                                    <river-select v-model="river"></river-select>
                                </div>
                            </div>
                            <div class='col-3' style="margin-left:0px">
                                <strong>Категория сложности: </strong><a target="_blank"
                                                                         href="https://huskytm.ru/rules2018-2019/#categories_tab"><img
                                    src="img/question_16.png"></a>
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
                                <a href="javascript:void(0);" v-on:click.stop="removeSpotEndPoint()"
                                   style="font-size: 80%;">Убрать конечную точку</a><br/>
                                <a href="javascript:void(0);" v-on:click.stop="addSpotAnchorPoint()"
                                   style="font-size: 80%;">Добавить опорную точку в середину</a><a target="_blank"
                                                                                                   href="javascript:void(0)"
                                                                                                   data-toggle="modal"
                                                                                                   data-target="#mid-point-about"><img
                                    src="img/question_16.png"></a>
                            </div>
                            <a v-else href="javascript:void(0);" v-on:click.stop="addSpotEndPoint()"
                               style="font-size: 80%;">Препятствие протяжённое. Добавить конечную точку</a>
                        </div>
                        <div class="col-10">
                            <ya-map-location-and-coords ref="locationEdit"
                                                        v-bind:spot="spot"
                                                        :zoom="zoom"
                                                        width="100%"
                                                        height="400px"
                                                        :editable="true"
                                                        :ya-search="true"
                                                        :switch-type-hotkeys="true"
                                                        v-bind:refresh-on-change="spot.point"
                                                        :show-map-by-default="true"
                                                        v-on:spotClick="$emit('spotClick', $event)"/>
                        </div>
                    </div>
                    <div class="row" style="padding-top: 12px;">
                        <div class="col-2"><strong>Краткое описание: </strong></div>
                        <div class="col-10">
                            <div class="wwmap-system-hint" style="color: red">Избегайте излишнего форматирования: этот текст показывается во всплывающей подсказке у порога на карте.</div>
                            <editor ref="descEditor"
                                  initialEditType="wysiwyg"
                                  :initialValue="spot.short_description"
                                  :options="editorOptions"
                                  v-on:change="onDescriptionChanged();"/>
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
                            <div class="wwmap-system-hint" style="margin-bottom: 7px;">Каждое альтернативное название на
                                новой строке
                            </div>
                            <textarea v-if="editMode" v-bind:text-content="spot.aliases"
                                      v-on:input="spot.aliases = parseAliases($event.target.value)"
                                      rows="10" cols="120">{{ spot.aliases.join('\n') }}</textarea>
                        </div>
                    </div>
                </b-tab>
                <b-tab title="Схемы" :disabled="spot.id<=0">
                    <img-upload ref="schemasList"
                                :spot="spot"
                                v-model="_schemas"
                                v-on:reloadAllImages="reloadAllImages"
                                type="schema"
                                :auth="true"></img-upload>
                </b-tab>
                <b-tab title="Фото" :disabled="spot.id<=0">
                    <img-upload ref="imagesList"
                                :spot="spot"
                                v-model="_images"
                                v-on:reloadAllImages="reloadAllImages"
                                type="image"
                                :auth="true"></img-upload>
                </b-tab>
                <b-tab title="Видео" :disabled="spot.id<=0">
                    <video-add :spot="spot" v-model="_videos" type="video" :auth="true"></video-add>
                </b-tab>
                <b-tab title="Системные параметры">
                    <span class="wwmap-system-hint" style="padding-top: 10px;">Тут собраны настройки разных системных вещей для каждого порога в отдельности</span>
                    <props v-if="spot.props" :p="spot.props"/>
                </b-tab>
            </b-tabs>
        </div>
    </div>
</template>

<script>
    import {getImages, getRiver, nvlReturningId, saveSpot, setActiveEntityUrlHash} from '../../editor'
    import {store} from '../../app-state';
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth';
    import {Editor} from '@toast-ui/vue-editor';
    import {markdownEditorConfig} from "../../toast-editor-config";

    const NEW_POINT_POSITION_KOEFF = 0.04;

    module.exports = {
        props: ['spot', 'country', 'region', 'images', 'schemas', 'videos', 'zoom'],

        components: {
          editor: Editor,
        },

        mounted: function () {
            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
            getRiver(this.spot.river.id).then(river => this.options = [river]);
            this.river = this.spot.river;
        },
        computed: {
            _images: {
                get() {
                    return this.images
                },
                set(images) {
                    this.$emit("images", images);
                }
            },
            _schemas: {
                get() {
                    return this.schemas
                },
                set(schemas) {
                    this.$emit("schemas", schemas);
                }
            },
            _videos: {
                get() {
                    return this.videos
                },
                set(videos) {
                    this.$emit("videos", videos)
                }
            },
            editMode: {
                get: function () {
                    return store.state.spoteditorstate.editMode
                },

                set: function (newVal) {
                    store.commit("setSpotEditorEditMode", newVal);
                }
            },
        },
        data: function () {
            return {
                map: null,
                canEdit: false,
                river: null,
                endPointBackup: null,
                editorOptions: markdownEditorConfig,
            }
        },

        methods: {
            onDescriptionChanged: function () {
              this.spot.short_description  = this.$refs.descEditor.invoke('getMarkdown');
            },
            save: function () {
                if (!this.spot.title || !this.spot.title.replace(/\s/g, '').length) {
                    this.showError("Нельзя сохранять порог без названия");
                    return
                }

                let oldRiver = this.spot.river;
                let riverChanged = this.river.id != oldRiver.id;
                this.spot.river = this.river;
                let t = this;

                this.$refs.btnBar.disable()
                saveSpot(this.spot).then(updated => {
                    this.hideError();

                    if (riverChanged) {
                        store.commit('hideRiverSubentities', {
                            countryId: oldRiver.region.country_id,
                            regionId: nvlReturningId(oldRiver.region),
                            riverId: oldRiver.id,
                        });
                    }

                    let countryId = this.spot.river.region.country_id;
                    let regionId = nvlReturningId(this.spot.river.region);
                    let riverId = this.spot.river.id;

                    setActiveEntityUrlHash(countryId, regionId, riverId, updated.id);
                    store.commit('setTreeSelection', {
                        contryId: countryId,
                        regionId: regionId,
                        riverId: riverId,
                        spotId: updated.id,
                    });
                    store.dispatch(riverChanged ? 'showRiverSubentities' : 'reloadRiverSubentities', {
                        countryId: countryId,
                        regionId: regionId,
                        riverId: riverId
                    });
                    this.editMode = false;

                    this.$emit('spot', updated)
                }, err => this.showError("Не удалось сохранить препятствие: " + err))
                    .finally(() => {
                        if (this.$refs.btnBar) {
                            this.$refs.btnBar.enable()
                        }
                    });
            },
            reloadAllImages() {
                getImages(this.spot.id, "image").then(images => this._images = images);
                getImages(this.spot.id, "schema").then(schemas => this._schemas = schemas);
            },
            closeEditorAndShowRiver: function () {
                let countryId = this.river.region.country_id;
                let regionId = nvlReturningId(this.river.region);
                let riverId = this.river.id;
                setActiveEntityUrlHash(countryId, regionId, riverId);
                store.commit('setTreeSelection', {
                    contryId: countryId,
                    regionId: regionId,
                    riverId: riverId,
                });
                store.dispatch('reloadRiverSubentities', {
                    countryId: countryId,
                    regionId: regionId,
                    riverId: riverId
                });
                store.commit('showRiverPage', {
                    country: {id: this.river.region.country_id},
                    region: this.river.region,
                    riverId: this.river.id
                });
            },

            editEndPoint: function () {
                return Array.isArray(this.spot.point[0])
            },
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
                    let p = [p1[0] + (p2[0] - p1[0]) * NEW_POINT_POSITION_KOEFF, p1[1] + (p2[1] - p1[1]) * NEW_POINT_POSITION_KOEFF];
                    this.spot.point.splice(l - 1, 0, p);
                }
            },
            removeSpotEndPoint: function () {
                if (Array.isArray(this.spot.point[0])) {
                    this.endPointBackup = this.spot.point[this.spot.point.length - 1];
                    this.spot.point = this.spot.point[0];
                }
            },

            parseAliases: function (strVal) {
                return strVal.split('\n')
                    .map(x => x.trim())
                    .filter(x => x.length > 0)
            },

            showError: function (errMsg) {
                store.commit("setErrMsg", errMsg);
            },
            hideError: function () {
                store.commit("setErrMsg", null);
            },
        },
    }

</script>