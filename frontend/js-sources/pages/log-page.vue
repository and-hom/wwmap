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

                  <td class="user-info">
                    {{ entry.user_info.first_name }}&nbsp;{{ entry.user_info.last_name }}
                    <div v-if="entry.auth_provider=='yandex'">
                      {{ entry.login }}@yandex.ru
                    </div>
                    <div v-else-if="entry.auth_provider=='google'">
                      Google: {{ entry.login }}
                    </div>
                    <div v-else-if="entry.auth_provider=='vk'">
                      <a target="_blank" :href="vkLink(entry.ext_id)">vk/{{ entry.login }}</a>
                    </div>
                    <div v-else>
                      {{ entry.auth_provider }}/{{ entry.ext_id }}
                    </div>
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

                    <td v-if="isCountry(entry)">
                        <a :href="getCountryUrl(entry.object_id)">{{getCountryTitle(entry.object_id)}}</a>
                    </td>
                    <td v-else-if="isRegion(entry)">
                        <a :href="getRegionUrl(entry.object_id)">{{getRegionTitle(entry.object_id)}}</a>
                    </td>
                    <td v-else-if ="isRiver(entry)">
                        <a :href="getRiverUrl(entry.object_id)">{{getRiverTitle(entry.object_id)}}</a>
                    </td>
                    <td v-else-if="isSpot(entry)">
                        <a :href="getSpotUrl(entry.object_id)">{{getSpotTitle(entry.object_id)}}</a>
                    </td>
                    <td v-else-if="isImage(entry)">
                        <a :href="getImageUrl(entry.object_id)">{{getImageTitle(entry.object_id)}}</a>
                    </td>
                    <td v-else-if="isCamp(entry)">
                        Стоянка {{entry.object_id}} {{entry.description}}
                    </td>
                    <td v-else-if="isVoyageReport(entry)">
                        <a :href="entry.description">Отчёт {{entry.object_id}} {{entry.description}}</a>
                    </td>
                    <td v-else-if="isTransfer(entry)">
                        Заброска {{entry.object_id}} {{entry.description}}
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

<style>
.user-info div {
  font-size: 60%;
  color: gray;
}
</style>

<script>
import {doGetJson, doPostJson} from '../api'
import {backendApiBase, frontendBase} from '../config'

export default {
        created: function () {
            doGetJson(backendApiBase + "/log", true).then(logs => {
                this.logs = logs;

                this.countryIds = this.logs.filter(this.isCountry).map(this.getObjId).filter(this.distinct);
                doGetJson(backendApiBase + "/country", false)
                    .then(countries => this.countryById = countries.reduce((map, country) => {
                      map[country.id] = country;
                      return map;
                    }, {}));

                this.regionIds = this.logs.filter(this.isRegion).map(this.getObjId).filter(this.distinct);
                doPostJson(backendApiBase + "/region_base_ids", this.regionIds, true).then(regionById => this.regionById = regionById);

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
                countryById: {},
                regionById: {},
                riverById: {},
                spotById: {},
                imgById: {},
                vkLink: function (vkId) {
                    return "https://vk.com/id" + vkId;
                },

                getCountryUrl: function (id) {
                    return this.countryById[id] ? `${frontendBase}/editor.htm#${id}` : null;
                },

                getCountryTitle: function (id) {
                    let countryData = this.countryById[id];
                    return !countryData ? ("Страна " + id) : countryData.title;
                },

                getRegionUrl: function (id) {
                    let regionData = this.regionById[id];
                    if (!regionData) {
                        return null
                    }
                    return `${frontendBase}/editor.htm#${regionData.country_id},${id}`;
                },

                getRegionTitle: function (id) {
                    let regionData = this.regionById[id];
                    return !regionData ? ("Регион " + id) : regionData.region_title;
                },

                getRiverUrl: function (id) {
                    let riverData = this.riverById[id];
                    if (!riverData) {
                        return null
                    }
                    return `${frontendBase}editor.htm#${riverData.country_id},${riverData.region_id},${id}`;
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
                    return `${frontendBase}/editor.htm#${spotData.country_id},${spotData.region_id},${spotData.river_id},${id}`;
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
                isCountry: function (logEntry) {
                    return logEntry.object_type == 'COUNTRY'
                },
                isRegion: function (logEntry) {
                    return logEntry.object_type == 'REGION'
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
                isCamp: function (logEntry) {
                    return logEntry.object_type == 'CAMP'
                },
                isVoyageReport: function (logEntry) {
                    return logEntry.object_type == 'REPORT'
                },
                isTransfer: function (logEntry) {
                    return logEntry.object_type == 'TRANSFER'
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
