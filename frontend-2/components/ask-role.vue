<template>
    <div :id="id" class="modal fade" tabindex="-1" role="dialog">
        <div class="modal-dialog modal-dialog-centered" role="document">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Выберите роль</h5>
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                </div>
                <div class="modal-body">
                    <p>{{msg}}</p>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-primary" v-for="role in roles" v-if="role!=currentRole" data-dismiss="modal" v-on:click="okfn(userId, role)">{{role}}</button>
                    <button type="button" class="btn btn-secondary" data-dismiss="modal">Отмена</button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>

    module.exports = {
        props: ['id', 'roles', 'okfn'],
        created: function() {
            component = this
            $(document).ready(function(){
                $(document).on('show.bs.modal','#'+component.id, function (event) {
                    component.userId =  $(event.relatedTarget).data('user-id')
                    component.msg =  $(event.relatedTarget).data('label')
                    component.currentRole = $(event.relatedTarget).data('current-role')
                })
            })
        },
        data: function() {
            return {
                userId: 0,
                msg: 'Выберите роль',
                currentRole: 'ANONYMOUS',
            }
        }
    }
</script>

<style>
</style>