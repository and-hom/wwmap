<template>
    <div>
        <window-title v-bind:text="country.title"></window-title>
        <div v-if="canEdit" class="btn-toolbar">
            <div class="btn-group mr-2" role="group">
                <button type="button" class="btn btn-primary" v-on:click="add_river()">Добавить реку без региона</button>
            </div>
        </div>
        <h1>{{ country.title }}</h1>
    </div>
</template>

<script>
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../auth'
    import {store} from "../main";

    module.exports = {
        props: ['country'],
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
                        region: {
                            id: 0,
                            title: this.country.title,
                            country_id: this.country.id,
                        }
                    })
                },
            }
        }
    }

</script>