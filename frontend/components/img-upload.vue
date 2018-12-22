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
            <thead>
                <tr>
                    <th></th>
                    <th>Ссылка на источник</th>
                    <th></th>
                </tr>
            </thead>
            <tr v-for="image in images">
                <td><img :class="imageClass(image)" :src="image.preview_url"/></td>
                <td>
                    <div v-if="image.source=='wwmap'">Загружено пользователем</div>
                    <a v-else target="_blank" :href="image.report_url">{{image.report_title}}</a>
                </td>
                <td>
                    <button v-if="image.enabled==false" v-on:click="setImgEnabled(true, image.id)" class="btn btn-success">Показывать</button>
                    <button v-if="image.enabled==true" v-on:click="setImgEnabled(false, image.id)"class="btn btn-secondary">Не показывать</button>
                    <button v-if="image.source=='wwmap'" v-on:click="removeImage(image.id)" class="btn btn-danger">Удалить</button>
                    <button v-if="!image.main_image" v-on:click="setSpotPreview(image.id)" class="btn btn-info">Сделать главным изображением</button>
                </td>
            </tr>
        </table>
    </div>
</template>

<script>
    module.exports = {
        props: ['spot', 'images', 'type', 'auth'],
        components: {
          FileUpload: VueUploadComponent
        },
        updated: function() {
            if(this.$refs[this.type] && this.$refs[this.type].value.length && this.$refs[this.type].uploaded) {
                var t = this;
                setTimeout(function() {
                    t.refresh()
                }, 700);
            }
        },
        computed: {
            headers:function(){
                if (this.auth) {
                    return {
                        Authorization: authHeaderVal(getSourceAndToken()),
                    }
                }
                return {}
            },
        },
        data:function() {
            return {
                files: [],
                uploadPath: function() { return backendApiBase + "/spot/" + this.spot.id +"/img?type=" + this.type },
                removeImage: function(imgId) {
                    this.images = removeImage(this.spot.id, imgId, this.type);
                },
                setImgEnabled: function(enabled, imgId) {
                    this.images = setImageEnabled(this.spot.id, imgId, enabled, this.type);
                },
                setSpotPreview: function(imgId) {
                    imgs = setSpotPreview(this.spot.id, imgId, this.type)
                    if (this.type=="image") {
                        Vue.set(app.spoteditorstate, "schemas", getImages(this.spot.id, "schema"))
                        Vue.set(app.spoteditorstate, "images", imgs)
                    } else {
                        Vue.set(app.spoteditorstate, "schemas", imgs)
                        Vue.set(app.spoteditorstate, "images", getImages(this.spot.id, "image"))
                    }
                },
                refresh:function() {
                    this.images = getImages(this.spot.id, this.type)
                },
                imageClass:function(image) {
                    if(image.enabled==false) {
                        return "wwmap-img-disabled"
                    }
                    return ""
                },
            }
        },
    }

</script>