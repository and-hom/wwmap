<template>
    <page link="transfer.htm">
        <div v-if="canEdit">
            <create-transfer :transfer="transferForEdit" :stations="stations" :okFn="editOk"
                             :cancelFn="resetTransferForEdit"/>
            <h2 style="display: inline;">Трансферы</h2>
            <button data-toggle="modal" data-target="#add-transfer" style="margin-left:30px;">+</button>
        </div>
        <table class="table">
            <thead>
            <tr>
                <th width="200px">Название</th>
                <th width="250px">Откуда</th>
                <th width="250px">Реки</th>
                <th>Контакты / Описание</th>
                <th v-if="canEdit"></th>
            </tr>
            </thead>
            <tbody>
            <tr :id="`row${transfer.id}`" :ref="`row${transfer.id}`" v-for="transfer in transfers"
                :class="rowClass(transfer)">
                <td>{{transfer.title}}</td>
                <td>
                    <ul class="ti-tags">
                        <li v-for="station in transfer.stations" class="ti-tag">
                            <div class="ti-content">
                                <div class="ti-tag-center"><span class="">{{station}}</span></div>
                            </div>
                        </li>
                    </ul>
                </td>
                <td>
                    <ul style="list-style-type: none; padding: 0;">
                        <li v-for="river in transfer.rivers">
                            <div style="width: 100%; min-height: 30px; display: block;">
                                <span>{{river.title}}</span>
                                <a :href="editorLink(river)" target="_blank"
                                   style="float:right;"><img
                                        src="https://wwmap.ru/img/edit.png" width="25px" :alt="editorLinkAlt"
                                        :title="editorLinkAlt"/></a>
                                <a v-if="river.bounds" :href="mapLine(river)" target="_blank"
                                   style="padding-left:10px;float:right;"><img
                                        src="https://wwmap.ru/img/locate.png" width="25px" alt="Показать на карте"
                                        title="Показать на карте"/></a>
                            </div>
                        </li>
                    </ul>
                </td>
                <td class="raw-text-content">{{transfer.description}}</td>
                <td v-if="canEdit">
                    <button v-on:click="transferForEdit={ ...transfer }"
                            data-toggle="modal" data-target="#add-transfer">Правка
                    </button>
                    <ask :id="'del-transfer-'+transfer.id" title="Удалить задачу?"
                         msg="Вся история и логи будут также удалены. Если удаление всех связанных данных не требуется, лучше отредактировать задачу, сняв флаг enabled"
                         :ok-fn="function() {
                          remove(transfer.id)
                        }"></ask>
                    <button data-toggle="modal" :data-target="'#del-transfer-'+transfer.id">Удалить</button>
                </td>
            </tr>
            </tbody>
        </table>
    </page>
</template>

<style>
    .row-selected {
        background: #eeeedd;
    }
</style>

<script>
    import {doDelete, doGetJson} from "../api";
    import {backendApiBase, frontendBase} from "../config";
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from "../auth";
    import Vue from "vue";

    const moment = require('moment');
    const VueScrollTo = require('vue-scrollto');

    export default {
        created: function () {
            Vue.use(VueScrollTo)
            this.refresh();
            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
        },
        mounted: function () {
        },
        beforeDestroy: function () {

        },
        updated: function () {
        },
        data() {
            return {
                transfers: [],
                canEdit: false,
                transferForEdit: {},
                stations: [],
                editorLinkAlt: this.canEdit ? 'Редактор' : 'Посмотреть в каталоге',
                selected: window.location.hash ? window.location.hash.substr(1) : '',
                scrolled: false,
            }
        },
        methods: {
            refresh: function () {
                doGetJson(backendApiBase + "/transfer-full", false).then(transfers => {
                    this.transfers = transfers;
                    this.stations = Array.from(new Set(transfers.flatMap(t => t.stations)));
                }).then(this.scrollToActive)

            },
            scrollToActive: function () {
                if (this.selected) {
                    let t = this;
                    this.$nextTick().then(_ => {
                        let id = `row${t.selected}`;
                        let elt = document.getElementById(id);
                        if (elt) {
                            VueScrollTo.scrollTo(`#row${t.selected}`, 1000, {});
                        } else {
                            console.log("Can't scroll to active")
                            setTimeout(t.scrollToActive, 100);
                        }
                    });
                }
            },

    resetTransferForEdit: function () {
                this.transferForEdit = {
                    title: "",
                    stations: [],
                    description: "",
                    rivers: [],
                }
            },
            editOk: function () {
                this.refresh();
                this.resetTransferForEdit();
            },
            remove: function (id) {
                doDelete(backendApiBase + "/transfer/" + id, true).then(this.refresh)
            },
            mapLine: function (river) {
                let bounds = this.centerOf(river.bounds);
                let z = this.zoomOf(river.bounds);
                return `${frontendBase}map.htm#${bounds[0]},${bounds[1]},${z}`
            },
            editorLink: function (river) {
                return `${frontendBase}editor.htm#${river.country_id},${river.region_id},${river.id}`
            },
            centerOf: function (point) {
                if (Array.isArray(point[0])) {
                    let p = [point[0], point[point.length - 1]];
                    return [(p[0][0] + p[1][0]) / 2, (p[0][1] + p[1][1]) / 2,]
                } else {
                    return point;
                }
            },
            zoomOf: function (p) {
                if (Array.isArray(p[0])) {
                    let dx = Math.abs(p[0][0] - p[1][0]);
                    let dy = Math.abs(p[0][1] - p[1][1]);
                    if (dx < 0.001 && dy < 0.001) {
                        return 15;
                    }
                    let d = Math.max(dx, dy);
                    let z = Math.log(180 / d) / Math.log(2) + 2;
                    return Math.min(Math.round(z), 19);
                } else {
                    return 15;
                }
            },
            rowClass: function (transfer) {
                return `${transfer.id}` == this.selected ? 'row-selected' : '';
            }
        }
    }
</script>
