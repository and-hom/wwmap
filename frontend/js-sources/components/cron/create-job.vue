<template>
    <div id="add-job" class="modal fade" tabindex="-1" role="dialog">
        <div class="modal-dialog modal-dialog-centered" role="document">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Добавить задачу</h5>
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                </div>
                <div class="modal-body">
                    <div class="wwmap-settings">
                        <label for="task_title">Название</label><input id="task_title" v-model="job.title"/>
                        <label for="cron_expr">Cron expression</label><input id="cron_expr" v-model="job.expr"/>
                        <label for="enabled">Включен</label><input type="checkbox" id="enabled" v-model="job.enabled"/>
                        <label for="critical">Критически важный</label><input type="checkbox" id="critical" v-model="job.critical"/>
                        <label for="command">Команда</label>
                        <select id="command" v-model="job.command">
                            <option v-for="(command_id, command) in commands" v-bind:value="command">
                               {{command}}
                            </option>
                        </select>
                        <label for="args">Аргументы</label><input id="args" v-model="job.args"/>
                        <label for="logsTtlDays">TTL логов (дней)</label><input id="logsTtlDays" type="number" min="0" step="1" v-model="job.logsTtlDays"/>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-primary" data-dismiss="modal" v-on:click="save()">
                        Сохранить
                    </button>
                    <button type="button" class="btn btn-secondary" data-dismiss="modal" v-on:click="cancel()">
                        Отмена
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import {cronApiBase} from '../../config'
    import {doPostJson} from '../../api'

    module.exports = {
        props: {
            job: {
                type: Object,
                required: true,
            },
            commands: {
                type: Object,
                required: true,
            },
            okFn: {
                type: Function,
                required: true,
            },
            cancelFn: {
                type: Function,
                required: false,
            },
        },
        data: function () {
            return {}
        },
        methods: {
            save: function () {
                doPostJson(cronApiBase + "/job", this.job, true).then(_ => {
                    this.okFn();
                });
            },
            cancel: function () {
                if (this.cancelFn) {
                    this.cancelFn();
                }
            },
        }
    }
</script>

<style>
</style>