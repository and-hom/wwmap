<template>
    <div>
        <window-title v-bind:text="country.title"></window-title>
        <div v-if="canEdit()" class="btn-toolbar">
            <div class="btn-group mr-2" role="group">
                <button type="button" class="btn btn-primary" v-on:click="add_river()">Добавить реку без региона</button>
            </div>
        </div>
        <h1>{{ country.title }}</h1>
    </div>
</template>

<script>
    import {getAuthorizedUserInfoOrNull} from '../auth'

    module.exports = {
        props: ['country'],
        created: function() {
            getAuthorizedUserInfoOrNull().then(userInfo => this.userInfo = userInfo);
        },
        data:function() {
            return {
                // for editor
                userInfo: null,
                canEdit: function(){
                 return this.userInfo!=null && (this.userInfo.roles.includes("EDITOR") || this.userInfo.roles.includes("ADMIN"))
                },

                getEditModeButtonTitle: function() {
                    return this.editMode ? 'Просмотр' : 'Редактирование';
                },
                // end of editor
                add_river: function() {return newRiver(this.country, {
                    id: 0,
                    title: this.country.title,
                    country_id: this.country.id,
                })},
            }
        }
    }

</script>