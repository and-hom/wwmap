<template>
    <page link="jobs.htm">
        <create-job :job="jobForEdit" :commands="commands" :okFn="editOk" :cancelFn="resetJobForEdit"/>
        <ask id="run-job-failed" title="Не получилось запустить" :noBtn="false" :msg="cantRunMessage" :okFn="function(){
            cantRunMessage=''
        }"></ask>

        <div style="margin-left:10px; margin-top: 10px;">
            <h2 style="display: inline;">Задачи</h2>
            <button data-toggle="modal" data-target="#add-job" style="margin-left:30px;">+</button>
            <table class="table">
                <thead>
                <tr>
                    <th width="50px">Id</th>
                    <th width="250px">Название</th>
                    <th width="150px">Cron-expr<a target="_blank" href="https://ru.wikipedia.org/wiki/Cron"><img src="img/question_16.png"></a></th>
                    <th>Флаги</th>
                    <th>Команда</th>
                    <th>Аргументы</th>
                    <th></th>
                </tr>
                </thead>
                <tbody>
                <tr v-for="job in jobs" :class="rowClass(job)">
                    <td class="wwmap_tooltip">{{job.id}}<span v-if="!job.registered && job.enabled" class="wwmap_tooltiptext">Не зарегистрирован: {{job.unregistered_reason}}</span></td>
                    <td>{{job.title}}</td>
                    <td>{{job.expr}}</td>
                    <td>
                        <div v-if="job.critical">Критически важный!</div>
                        <div v-if="!job.enabled" style="color:gray;">Отключен</div>
                        <div v-if="job.enabled && !job.registered" style="color:red;">Ошибка регистрации</div>
                    </td>
                    <td>{{job.command}}</td>
                    <td>{{job.args}}</td>
                    <td>
                        <ask :id="'run-job-'+job.id" title="Запустить сейчас?" msg="Запустить задачу вне расписания прямо сейчас" :ok-fn="function() {
                          runNow(job.id)
                        }"></ask>
                        <ask :id="'del-job-'+job.id" title="Удалить задачу?" msg="Вся история и логи будут также удалены. Если удаление всех связанных данных не требуется, лучше отредактировать задачу, сняв флаг enabled" :ok-fn="function() {
                          remove(job.id)
                        }"></ask>


                        <button data-toggle="modal" :data-target="'#run-job-'+job.id">Запустить сейчас</button>
                        <button v-on:click="jobForEdit={ ...job }" data-toggle="modal" data-target="#add-job">Правка
                        </button>
                        <button data-toggle="modal" :data-target="'#del-job-'+job.id">Удалить</button>
                    </td>
                </tr>
                </tbody>
            </table>
        </div>
    </page>
</template>

<style type="text/css">
    .job-normal {

    }

    .job-disabled {
        color: gray;
    }

    .job-unregistered td:first-child {
        color: #ff6622;
        text-decoration: line-through;
    }

    .wwmap_tooltip .wwmap_tooltiptext {
        visibility: hidden;
        background-color: black;
        color: #fff;
        text-align: center;
        border-radius: 6px;
        padding: 5px 7px;

        /* Position the wwmap_tooltip */
        position: absolute;
        z-index: 1;
    }

    .wwmap_tooltip:hover .wwmap_tooltiptext {
        visibility: visible;
    }
</style>

<script>
    import {doDelete, doGetJson, doPost} from '../api'
    import {cronApiBase} from '../config'

    export default {
        created: function () {
            this.getCommands();
            this.refresh();
            this.resetJobForEdit();
        },
        data() {
            return {
                jobs: [],
                commands: [],
                jobForEdit: {},
                cantRunMessage: "",
            }
        },
        methods: {
            editOk: function () {
                this.refresh();
                this.resetJobForEdit();
            },
            getCommands: function () {
                doGetJson(cronApiBase + "/commands", true).then(commands => {
                    this.commands = commands;
                })
            },
            refresh: function () {
                doGetJson(cronApiBase + "/job", true).then(jobs => {
                    this.jobs = jobs;
                })
            },
            runNow: function (id) {
                let t = this;
                doPost(cronApiBase + "/job/" + id + "/run", null, true).catch(err => {
                    t.cantRunMessage = err;
                    let dialog = $('#run-job-failed');
                    dialog.on('hidden.bs.modal', e => {
                        // state.closeCallback = function () {
                        //
                        // };
                    });
                    dialog.modal();
                })
            },
            remove: function (id) {
                doDelete(cronApiBase + "/job/" + id, true).then(this.refresh)
            },
            rowClass: function (job) {
                return job.enabled ? (job.registered ? 'job-normal' : 'job-unregistered') : 'job-disabled';
            },
            resetJobForEdit: function () {
                this.jobForEdit = {
                    title: "",
                    expr: "",
                    enabled: false,
                    critical: false,
                    command: "",
                    args: "",
                }
            },
        }
    }
</script>
