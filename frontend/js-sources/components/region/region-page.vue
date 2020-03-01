<template>
    <div>
        <window-title v-bind:text="region.title"></window-title>
        <div v-if="canEdit" class="btn-toolbar">
            <div class="btn-group mr-2" role="group">
                <button type="button" class="btn btn-primary" v-on:click="add_river()">Добавить реку</button>
            </div>
        </div>
        <breadcrumbs :country="country"/>
        <h1>{{ region.title }}</h1>
    </div>
</template>

<script>
    import {store} from '../../app-state';
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth';

    module.exports = {
        props: ['region', 'country'],
        created: function() {
            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
        },
        data:function() {
            return {
                // for editor
                canEdit: false,

                getEditModeButtonTitle: function() {
                    return this.editMode ? 'Просмотр' : 'Редактирование';
                },
                // end of editor

                add_river: function () {
                    store.commit('newRiver', {
                        country: this.country,
                        region: this.region})
                },
            }
        }
    }

</script>