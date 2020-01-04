<template>

    <page link="log.htm">
        <div style="margin-left:10px; margin-top: 10px;">
            <h2>История изменений</h2>
            <table class="table">
                <thead>
                <tr>
                    <td>Время</td>
                    <td>Пользователь</td>
                    <td>Действие</td>
                    <td>Тип/Id объекта</td>
                    <td>Описание</td>
                </tr>
                </thead>
                <tbody>
                <tr v-for="entry in logs" :class="rowStyle(entry.type)">
                    <td>{{entry.time}}</td>

                    <td v-if="entry.auth_provider=='yandex'">
                        {{entry.login}}@yandex.ru
                    </td>
                    <td v-else-if="entry.auth_provider=='google'">
                        Google: {{entry.login}}
                    </td>
                    <td v-else-if="entry.auth_provider=='vk'">
                        <a target="_blank" :href="vkLink(entry.ext_id)">vk/{{entry.login}}</a>
                    </td>
                    <td v-else>
                        {{entry.auth_provider}}/{{entry.ext_id}}
                    </td>

                    <td v-if="entry.type=='CREATE'">
                        Создан
                    </td>
                    <td v-else-if="entry.type=='MODIFY'">
                        Изменён
                    </td>
                    <td v-else-if="entry.type=='DELETE'">
                        Удалён
                    </td>
                    <td v-else>
                        {{entry.type}}
                    </td>

                    <td v-if="isRiver(entry)">
                        <a :href="getRiverUrl(entry.object_id)">{{getRiverTitle(entry.object_id)}}</a>
                    </td>
                    <td v-else-if="isSpot(entry)">
                        <a :href="getSpotUrl(entry.object_id)">{{getSpotTitle(entry.object_id)}}</a>
                    </td>
                    <td v-else-if="isImage(entry)">
                        <a :href="getImageUrl(entry.object_id)">{{getImageTitle(entry.object_id)}}</a>
                    </td>
                    <td v-else-if="isUser(entry)">
                        Пользователь {{entry.object_id}}
                    </td>
                    <td v-else>
                        {{entry.object_type}}/{{entry.object_id}}
                    </td>
                    <td>{{entry.description}}</td>
                </tr>
                </tbody>
            </table>
        </div>
    </page>
</template>

<script>
    import {doGetJson, doPostJson} from './api'
    import {backendApiBase} from './config'

    export default {
        created: function () {
            doGetJson(backendApiBase + "/log", true).then(logs => {
                this.logs = logs;

                this.riverIds = this.logs.filter(this.isRiver).map(this.getObjId).filter(this.distinct);
                doPostJson(backendApiBase + "/river_base_ids", this.riverIds, true).then(riverById => this.riverById = riverById);

                this.imageIds = this.logs.filter(this.isImage).map(this.getObjId).filter(this.distinct);
                doPostJson(backendApiBase + "/image_base_ids", this.imageIds, true).then(imgById => this.imgById = imgById);

                this.spotIds = this.logs
                    .filter(this.isSpot)
                    .map(this.getObjId)
                    .concat(Object.values(this.imageIds))
                    .filter(this.distinct);
                doPostJson(backendApiBase + "/spot_base_ids", this.spotIds, true).then(spotById => this.spotById = spotById);
            });
        },
        data() {
            return {
                logs: [],
                riverById: {},
                spotById: {},
                imgById: {},
                vkLink: function (vkId) {
                    return "https://vk.com/id" + vkId;
                },

                getRiverUrl: function (id) {
                    let riverData = this.riverById[id];
                    if (!riverData) {
                        return null
                    }
                    return "https://wwmap.ru/editor.htm#" + riverData.country_id + "," + riverData.region_id + "," + id;
                },

                getRiverTitle: function (id) {
                    let riverData = this.riverById[id];
                    return !riverData ? ("Река " + id) : riverData.river_title;
                },

                getSpotUrl: function (id) {
                    let spotData = this.spotById[id];
                    if (!spotData) {
                        return null
                    }
                    return "https://wwmap.ru/editor.htm#" + spotData.country_id + "," + spotData.region_id + "," + spotData.river_id + "," + id;
                },

                getSpotTitle: function (id) {
                    let spotData = this.spotById[id];
                    return !spotData ? ("Порог " + id) : spotData.river_title + " / " + spotData.spot_title;
                },

                getImageUrl: function (id) {
                    let imgData = this.imgById[id];
                    if (!imgData || !imgData.spot_id) {
                        return null
                    }
                    return this.getSpotUrl(imgData.spot_id);
                },

                getImageTitle: function (id) {
                    let imgData = this.imgById[id];
                    return "Изображение " + id
                        + (!imgData || !imgData.spot_id ? "" : " для порога " + this.getSpotTitle(imgData.spot_id));
                },

                distinct: function (value, index, self) {
                    return self.indexOf(value) === index;
                },
                getObjId: function (logEntry) {
                    return logEntry.object_id
                },
                isRiver: function (logEntry) {
                    return logEntry.object_type == 'RIVER'
                },
                isSpot: function (logEntry) {
                    return logEntry.object_type == 'SPOT'
                },
                isImage: function (logEntry) {
                    return logEntry.object_type == 'IMAGE'
                },
                isUser: function (logEntry) {
                    return logEntry.object_type == 'USER'
                },

                rowStyle: function (entryType) {
                    switch (entryType) {
                        case 'CREATE':
                            return 'wwmap-changes-log-create';
                        case 'MODIFY':
                            return 'wwmap-changes-log-change';
                        case 'DELETE':
                            return 'wwmap-changes-log-delete';
                    }
                    return ""
                },
            }
        }
    }
</script>
