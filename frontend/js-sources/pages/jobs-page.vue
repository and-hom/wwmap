<template>
    <page link="jobs.htm">
        <create-job :job="jobForEdit" :okFn="editOk" :cancelFn="resetJobForEdit"/>
        <div v-if="logs!=null" style="margin-left:10px; margin-top: 10px;">
            <h2>Логи</h2>
            <a href="#" v-on:click="logs=null">Назад к задачам</a>
            <div v-for="log in logs">{{log}}</div>
        </div>
        <div v-else style="margin-left:10px; margin-top: 10px;">
            <h2 style="display: inline;">Задачи</h2>
            <button data-toggle="modal" data-target="#add-job" style="margin-left:30px;">+</button>
            <table class="table">
                <thead>
                <tr>
                    <td width="50px">Id</td>
                    <td width="250px">Название</td>
                    <td width="150px">Cron-expr</td>
                    <td width="250px"></td>
                    <td>Команда</td>
                    <td></td>
                </tr>
                </thead>
                <tbody>
                <tr v-for="job in jobs" :class="rowClass(job)">
                    <td>{{job.id}}</td>
                    <td>{{job.title}}</td>
                    <td>{{job.expr}}</td>
                    <td><a href="#" v-on:click="showLogs(job.id)">Логи</a></td>
                    <td>{{job.command}}</td>
                    <td>
                        <button v-on:click="jobForEdit={ ...job }" data-toggle="modal" data-target="#add-job">Правка
                        </button>
                    </td>
                </tr>
                </tbody>
            </table>
        </div>
    </page>
</template>

<style type="text/css">
    .job-enabled {

    }

    .job-disabled {
        color: gray;
        text-decoration: line-through;
    }
</style>

<script>
    import {doGetJson} from '../api'
    import {cronApiBase} from '../config'

    export default {
        created: function () {
            this.refresh();
            this.resetJobForEdit();
        },
        data() {
            return {
                jobs: [],
                jobForEdit: {},
                logs: null,
            }
        },
        methods: {
            editOk: function () {
                this.refresh();
                this.resetJobForEdit();
            },
            refresh: function () {
                doGetJson(cronApiBase + "/job", true).then(jobs => {
                    this.jobs = jobs;
                })
            },
            rowClass: function (job) {
                return job.enabled ? 'job-enabled' : 'job-disabled'
            },
            resetJobForEdit: function () {
                this.jobForEdit = {
                    title: "",
                    expr: "",
                    enabled: false,
                    command: "",
                }
            },
            showLogs: function (jobId) {
                doGetJson(cronApiBase + "/logs/" + jobId, true).then(logs => {
                    this.logs = logs;
                })
            }
        }
    }
</script>
