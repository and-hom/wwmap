<template>
    <div>
        <div v-if="canEdit()" class="btn-toolbar">
            <div class="btn-group mr-2" role="group">
                <button type="button" class="btn btn-primary" v-on:click="add_river()">Добавить реку без региона</button>
            </div>
        </div>
    </div>
</template>

<script>
    module.exports = {
        props: ['country'],
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
                add_river: function() {return newRiver(this.country, {
                    id: 0,
                    title: this.country.title,
                    country_id: this.country.id,
                })},
            }
        }
    }

</script>