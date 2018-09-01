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
                <button type="button" class="btn btn-primary" v-on:click="add_spot()">Добавить препятствие</button>
            </div>
            <div class="btn-group mr-2" role="group" aria-label="First group">
                <button type="button" class="btn btn-info" v-on:click="editMode=!editMode; hideError();">
                    {{getEditModeButtonTitle()}}
                </button>
                <button type="button" class="btn btn-success" v-if="editMode" v-on:click="save()">Сохранить</button>
                <button type="button" class="btn btn-secondary" v-if="editMode" v-on:click="reload()">Отменить</button>
            </div>
            <div class="btn-group">
                <button type="button" class="btn btn-danger" data-toggle="modal" data-target="#del-river">Удалить
                </button>
            </div>
        </div>

        <input v-if="editMode" v-model.trim="river.title" style="display:block"/>
        <h1 v-else>{{ river.title }}</h1>
        <dl>
            <dt>Регион:</dt>
            <dd>
                <div v-if="editMode">
                    <select v-model="river.region.id">
                        <option v-for="region in regions" v-bind:value="region.id">{{region.title}}</option>
                    </select>
                </div>
                <div v-else style="padding-left:40px;">
                    {{river.region.title}}
                </div>
            </dd>
            <dt>Алиасы:</dt>
            <dd>
                <textarea v-if="editMode" v-bind:text-content="river.aliases"
                          v-on:input="river.aliases = parseAliases($event.target.value)"
                          rows="10" cols="120"
                          style="resize: none; margin-left:40px;">{{ river.aliases.join('\n') }}</textarea>
                <ul v-else>
                    <li v-for="alias in river.aliases">{{alias}}</li>
                </ul>

            </dd>
            <dt>Отчёты:</dt>
            <dd>
                <ul>
                    <li v-for="report in reports"><a target="_blank" :href="report.url">{{report.title}}</a></li>
                </ul>
            </dd>
        </dl>
        <div v-if="canEdit() && river.id && !editMode">
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
</template>

<script>
    module.exports = {
        props: ['river', 'reports'],
        components: {
          FileUpload: VueUploadComponent
        },
        updated: function() {
            if(this.$refs.uploadGpx && this.$refs.uploadGpx.value.length && this.$refs.uploadGpx.uploaded) {
                if (this.imagesOutOfDate) {
                    setTimeout(() => {
                        this.refresh()
                    }, 700);
                    this.imagesOutOfDate = false
                } else {
                    this.imagesOutOfDate = true
                }
            }
        },
        data:function() {
            return {
                // for editor
                userInfo: getAuthorizedUserInfoOrNull(),
                canEdit: function(){
                 return this.userInfo!=null && this.userInfo.roles.includes("ADMIN")
                },
                editMode: app.rivereditorstate.editMode,
                errMsg:null,
                askForRemove: true,
                save:function() {
                    updated = saveRiver(this.river)
                    if (updated) {
                        this.river = updated
                        this.editMode=false
                        this.hideError()
                    } else {
                        this.showError("Не удалось сохранить реку. Возможно, недостаточно прав")
                    }
                },
                reload:function() {
                    this.river = getRiver(this.river.id)
                    this.hideError()
                },
                remove: function() {
                    this.hideError()
                    if (!removeRiver(this.river.id)) {
                        this.showError("Can not delete")
                    }
                },
                showError: function(errMsg) {
                    this.errMsg = errMsg
                },
                hideError: function(errMsg) {
                    this.errMsg = null
                },
                getEditModeButtonTitle: function() {
                    return this.editMode ? 'Просмотр' : 'Редактирование';
                },
                // end of editor

                files: [],
                uploadPath: apiBase + "/river/" + this.river.id +"/gpx",
                add_spot: function() {
                    app.spoteditorstate.visible = false
                    app.rivereditorstate.visible = false;

                    app.spoteditorstate.visible = true;
                    app.spoteditorstate.editMode = true;
                    app.spoteditorstate.spot={
                        id: 0,
                        river: this.river,
                        point:[0.0, 0.0, ],
                    }
                },

                regions: getAllRegions(),
                parseAliases:function(strVal) {
                    return strVal.split('\n').map(function(x) {return x.trim()}).filter(function(x){return x.length>0})
                },
            }
        }
    }

</script>