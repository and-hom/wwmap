<template>
    <div :id="id" class="modal fade" tabindex="-1" role="dialog">
        <div class="modal-dialog modal-dialog-centered" role="document">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">{{ title }}</h5>
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                </div>
                <div class="modal-body">
                    <p>{{ msg }}</p>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-primary" data-dismiss="modal" v-on:click="okFn()">{{okBtnTitle}}</button>
                    <button v-if="noBtn" type="button" class="btn btn-secondary" data-dismiss="modal" v-on:click="noFn()">{{noBtnTitle}}</button>
                    <button v-if="cancelBtn" type="button" class="btn btn-secondary" data-dismiss="modal">{{cancelBtnTitle}}</button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    module.exports = {
        props: {
            id: {
                type: String,
            },
            msg: {
                type: String,
            },
            title: {
                type: String,
            },
            okBtnTitle: {
                type: String,
                default: "Да"
            },
            noBtnTitle: {
                type: String,
                default: "Нет"
            },
            cancelBtnTitle: {
                type: String,
                default: "Отмена"
            },
            okFn: {
                type: Function,
                default: function () {

                }
            },
            noFn: {
                type: Function,
                default: function () {

                }
            },
            cancelBtn: {
                type: Boolean,
                default: false
            },
            noBtn: {
                type: Boolean,
                default: true
            },
        },
        data: function() {
            return {

            }
        }
    }
</script>

<style>
</style>