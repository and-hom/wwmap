<template>
    <div class="spot-display">
        <btn-bar v-if="canEdit" logObjectType="SPOT" :logObjectId="spot.id">
            <slot></slot>
        </btn-bar>
        <breadcrumbs :country="country" :region="region" :river="river"/>
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
                <div class="wwmap-system-hint">Эти параметры предназначены для определения порядка следования
                    порогов. Автоматическое упорядочивание проходит раз в сутки ночью.
                </div>
                <div><strong>Порядок следования:</strong> {{ spot.order_index }}</div>
                <div><strong>Автоматическое упорядочивание:</strong>&nbsp;<span
                        v-if="spot.automatic_ordering">Да</span><span v-else>Нет</span></div>
                <div><strong>В последний раз автоматическое упорядочивание срабатывало:</strong> {{
                    lastAutoOrdering() }}
                </div>
            </div>
        </div>
        <div class="wwmap-desc-section-div">
            <div v-if="spot.lw_category!=='0' ||spot.mw_category!=='0' || spot.hw_category!=='0'"><strong>К.с.
                нв/св/вв:</strong>&nbsp;<category :category="spot.lw_category"></category>
                /
                <category :category="spot.mw_category"></category>
                /
                <category :category="spot.hw_category"></category>
            </div>
            <div><strong>К.с. по классификатору:</strong>&nbsp;<category :category="spot.category"></category>
            </div>
        </div>
        <div class="wwmap-desc-section-div" v-if="spot.short_description">
          <div style="padding-left:40px;">
            <viewer :initialValue="spot.short_description"/>
          </div>
        </div>
        <div class="wwmap-desc-section-div">
            <ya-map-location ref="locationView"
                             v-bind:spot="spot"
                             width="70%" height="600px"
                             :editable="false"
                             :zoom="15"
                             v-on:spotClick = "$emit('spotClick', $event)"/>
            <div style="padding-top:4px;">
                <strong>Широта:</strong>&nbsp;{{ spotPoint0()[0] }}&nbsp;&nbsp;&nbsp;<strong>Долгота:</strong>&nbsp;{{
                spotPoint0()[1] }}
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
                <gallery id="schemas-gallery" :images="schemas.map(function(x) {return x.url})" :index="schIndex"
                         @close="schIndex = null"></gallery>
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
                <gallery id="image-gallery" :images="images.map(function(x) {return x.url})" :index="imgIndex"
                         @close="imgIndex = null"></gallery>
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
    import {getRiver} from '../../editor'
    import {store} from '../../app-state';
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth';
    import {backendApiBase} from '../../config'
    import { Viewer } from '@toast-ui/vue-editor';

    module.exports = {
        props: ['spot', 'country', 'region', 'images', 'schemas', 'videos'],

        components: {
          viewer: Viewer,
        },

        mounted: function () {
            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
            getRiver(this.spot.river.id).then(river => this.options = [river]);
            this.river = this.spot.river;
        },

        computed: {
            spotMainUrl: function() {
                let mainImgs = this.images.filter(i => i.main_image);
                if (mainImgs.length > 0) {
                    return mainImgs[0].preview_url;
                }
                mainImgs = this.schemas.filter(i => i.main_image);
                if (mainImgs.length > 0) {
                    return mainImgs[0].preview_url;
                }
                return null
            },
            mainImageIndex: function () {
                var imgIdx = this.findMainImgIdx(this.images, this.spotMainUrl);
                if (imgIdx > -1) {
                    return imgIdx;
                }
                return null;
            },
            mainSchemaIndex: function () {
                var imgIdx = this.findMainImgIdx(this.schemas, this.spotMainUrl);
                if (imgIdx > -1) {
                    return imgIdx;
                }
                return null;
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
                river: null,
                map: null,
                label: null,
                canEdit: false,
                askForRemove: false,

                options: [],

                // imgs
                imgIndex: null,
                schIndex: null,
                vidIndex: null,
            }
        },

        methods: {
            onSearch: function (search, loading) {
                loading(true);
                var component = this
                fetch(
                    backendApiBase + '/river?q=' + search
                ).then(function (res) {
                    res.json().then(function (json) {
                        component.options = json
                    });
                    loading(false);
                });
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
            lastAutoOrdering: function () {
                var lastOrderingDate = new Date(this.spot.last_automatic_ordering)
                if (lastOrderingDate.getFullYear() <= 2017) {
                    return 'Никогда'
                }
                return this.spot.last_automatic_ordering
            },
            spotPoint0: function () {
                if (Array.isArray(this.spot.point[0])) {
                    return this.spot.point[0]
                } else {
                    return this.spot.point
                }
            },
        },
    }

</script>