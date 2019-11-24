<template>
    <div>
        <window-title v-bind:text="region.title"></window-title>
        <div v-if="canEdit()" class="btn-toolbar">
            <div class="btn-group mr-2" role="group">
                <button type="button" class="btn btn-primary" v-on:click="add_river()">Добавить реку</button>
            </div>
        </div>
        <h1>{{ region.title }}</h1>
    </div>
</template>

<script>
    module.exports = {
        props: ['region', 'country'],
        data:function() {
            return {
                // for editor
                userInfo: getAuthorizedUserInfoOrNull(),
                canEdit: function(){
                 return this.userInfo!=null && (this.userInfo.roles.includes("EDITOR") || this.userInfo.roles.includes("ADMIN"))
                },

                getEditModeButtonTitle: function() {
                    return this.editMode ? 'Просмотр' : 'Редактирование';
                },
                // end of editor

                add_river: function() {return newRiver(this.country, this.region)},
            }
        }
    }

</script>