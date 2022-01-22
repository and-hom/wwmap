<template>
    <div>
        <template>
            <div class="example-drag">
                <div class="upload" style="margin-top:15px;">
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

                    <div class="example-btn">
                        <file-upload
                                class="btn btn-primary"
                                :post-action="uploadPath()"
                                :headers="headers"
                                extensions="gif,jpg,jpeg,png"
                                accept="image/png,image/gif,image/jpeg"
                                :multiple="true"
                                :drop="true"
                                :drop-directory="true"
                                v-model="files"
                                :ref="type"
                                :input-id="type">
                            <i class="fa fa-plus"></i>
                            Выбрать изображения
                        </file-upload>
                        <button type="button" class="btn btn-success" v-if="!$refs[type] || !$refs[type].active"
                                @click.prevent="$refs[type].active = true">
                            <i class="fa fa-arrow-up" aria-hidden="true"></i>
                            Начать загрузку
                        </button>
                        <button type="button" class="btn btn-danger" v-else
                                @click.prevent="$refs[type].active = false">
                            <i class="fa fa-stop" aria-hidden="true"></i>
                            Остановить загрузку
                        </button>
                    </div>
                </div>
            </div>
        </template>
        <table class="table">
            <tr v-for="image in images">
                <td><img :class="imageClass(image)" :src="image.preview_url"/></td>
                <td>
                    <div v-if="image.source=='wwmap'">Загружено пользователем</div>
                    <div v-else>Из отчёта <a target="_blank" :href="image.report_url">{{image.report_title}}</a></div>
                    Опубликовано {{image.date_published|formatDateTimeStr}}
                    <div v-if="image.type=='image'"><b>Фактическая</b> дата снимка:
                        <img-date-and-level :spot-id="spot.id"
                                            :img-id="image.id"
                                            :date="image.date"
                                            :level="image.level"
                                            v-on:level="image.level = $event"
                                            v-on:date="image.date = $event"/>
                    </div>
                </td>
                <td>
                    <button v-if="image.enabled==false" v-on:click="setImgEnabled(true, image.id)"
                            class="btn btn-success">Показывать
                    </button>
                    <button v-if="image.enabled==true" v-on:click="setImgEnabled(false, image.id)"
                            class="btn btn-secondary">Не показывать
                    </button>
                    <ask :id="'del-img-' + image.id" title="Точно?" msg="Удалить изображение?"
                         :ok-fn="function() { removeImage(image.id); }"></ask>
                    <button data-toggle="modal" :data-target="'#del-img-' + image.id" class="btn btn-danger">Удалить
                    </button>
                    <button v-if="!image.main_image" v-on:click="setSpotPreview(image.id)" class="btn btn-info">Сделать
                        главным изображением
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
    import FileUpload from 'vue-upload-component';
    import {backendApiBase} from '../config'
    import {getWwmapSessionId} from "wwmap-js-commons/auth";
    import {getImages, removeImage, setImageEnabled, setSpotPreview} from "../editor";

    module.exports = {
        props: ['spot', 'value', 'type', 'auth'],
        components: {
            FileUpload: FileUpload,
        },
        updated: function () {
            if (this.$refs[this.type] && this.$refs[this.type].value.length && this.$refs[this.type].uploaded) {

                var t = this;
                setTimeout(function () {
                    t.refresh()
                }, 700);
            }
        },
        computed: {
            headers: function () {
                if (this.auth) {
                    return {
                        Authorization: getWwmapSessionId(),
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
            }
        },
        data: function () {
            return {
                files: [],
                filesSum: "",
                uploadPath: function () {
                    return backendApiBase + "/spot/" + this.spot.id + "/img?type=" + this.type
                },
                removeImage: function (imgId) {
                    removeImage(this.spot.id, imgId, this.type)
                        .then(r => getImages("spot", this.spot.id, this.type))
                        .then(images => this.images = images);
                },
                setImgEnabled: function (enabled, imgId) {
                    setImageEnabled(imgId, enabled, this.type)
                        .then(r => getImages("spot", this.spot.id, this.type))
                        .then(images => this.images = images);
                },
                setSpotPreview: function (imgId) {
                    setSpotPreview(this.spot.id, imgId, this.type).then(images => {
                        this.$emit("reloadAllImages");
                    });
                },
                refresh: function () {
                    getImages("spot", this.spot.id, this.type).then(images => {
                        // Workaround #140 do not refresh the same images list. It produces update event and then refresh and then update and then.....
                        var filesSum = images.map(function (x) {
                            return x.id
                        }).join("#");
                        if (this.filesSum !== filesSum) {
                            this.filesSum = filesSum;
                            this.images = images;
                        }
                    });
                },
                imageClass: function (image) {
                    return image.enabled ? "" : "wwmap-img-disabled";
                },
            }
        },
    }

</script>