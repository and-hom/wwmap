<template>
    <div>
        <div>
            <label for="video_url" style="padding-right: 10px;">Ссылка на видео с youtube (как в браузере)</label><input
                id="video_url" type="text" style="width:400px;"/>
            <input type="button" class="btn btn-success" value="Добавить" v-on:click.prevent="onAddVideo"/>
        </div>

        <table class="table">
            <tr v-for="image in images">
                <td style="width: 1px;">
                    <youtube-embedded-video :image="image" width="560" height="315"/>
                </td>
                <td>
                    <button v-if="image.enabled===false" v-on:click="setImgEnabled(true, image.id)"
                            class="btn btn-success">Показывать
                    </button>
                    <button v-if="image.enabled===true" v-on:click="setImgEnabled(false, image.id)"
                            class="btn btn-secondary">Не показывать
                    </button>
                    <ask :id="'del-video-' + image.id" title="Точно?" msg="Удалить видео?"
                         :ok-fn="function() { removeImage(image.id); }"></ask>
                    <button data-toggle="modal" :data-target="'#del-video-' + image.id" class="btn btn-danger">
                        Удалить
                    </button>
                    <div style="margin-top: 20px;">
                        <log-dropdown object-type="IMAGE" :object-id="image.id"/>
                    </div>
                </td>
            </tr>
        </table>
    </div>
</template>

<script>
    import {store} from '../app-state';
    import {getImages, removeImage, setImageEnabled} from "../editor"
    import {parseParams, doPostJson} from "../api"
    import {getWwmapSessionId} from "wwmap-js-commons/auth";
    import {backendApiBase} from '../config'

    module.exports = {
        props: ['spot', 'value', 'type', 'auth'],
        computed: {
            headers: function () {
                if (this.auth) {
                    return {
                        Authorization: getWwmapSessionId()
                    }
                }
                return {}
            },
            images: {
                get() {
                    return this.value;
                },
                set(images) {
                    this.$emit('input', images);
                }
            },
        },
        data: function () {
            return {
                removeImage: function (imgId) {
                    removeImage(this.spot.id, imgId, this.type).then(images => this.images = images);
                },
                setImgEnabled: function (enabled, imgId) {
                    setImageEnabled(this.spot.id, imgId, enabled, this.type).then(images => this.images = images);
                },
                imageClass: function (image) {
                    if (image.enabled === false) {
                        return "wwmap-img-disabled"
                    }
                    return ""
                },
                uploadPath: function () {
                    return backendApiBase + "/spot/" + this.spot.id + "/img_ext?type=" + this.type
                },
                onAddVideo: function () {
                    try {
                        var videoUrl = $("#video_url").val();
                        if (!videoUrl) {
                            throw "Пустая ссылка на видео"
                        }
                        var url = document.createElement('a');
                        url.href = videoUrl;
                        if (!url.search) {
                            throw "В ссылке отсутствуют GET-параметры. Попробуйте ещё раз скопировать ссылку из адресной браузера"
                        }
                        var params = parseParams(url.search.substr(1));
                        var videoId = params["v"];
                        if (!videoId) {
                            throw "В ссылке отсутствует параметр v. Попробуйте ещё раз скопировать ссылку из адресной браузера"
                        }

                        let tStart = params['t']
                        let tStartInt;
                        try {
                            tStartInt = tStart ? parseInt(tStart) : 0;
                        } catch (e) {
                            console.error(e)
                            tStartInt = 0;
                        }

                        var requestData = {
                            id: videoId,
                            type: "video",
                            source: "youtube",
                            props: {
                              t: tStartInt
                            }
                        };
                        var t = this;
                        doPostJson(this.uploadPath(), requestData, true).then(resp => {
                            if (resp) {
                               getImages(this.spot.id, "video").then(images => t.images = images);
                                t.hideError();
                            } else {
                                t.showError("Не удалось добавить видео")
                            }
                        });
                    } catch (e) {
                        this.showError("Не удалось добавить видео: " + e)
                    }
                },
                showError: function (errMsg) {
                    store.commit("setErrMsg", errMsg);
                },
                hideError: function () {
                    store.commit("setErrMsg", null);
                },
            }
        }
    }

</script>