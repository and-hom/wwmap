<template>
    <div>
        <div v-if="canEdit()" class="btn-toolbar">
            <div class="btn-group mr-2" role="group">
                <button type="button" class="btn btn-primary" v-on:click="add_river()">Добавить реку</button>
            </div>
        </div>
    </div>
</template>

<script>
    module.exports = {
        props: ['region'],
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

                add_river: function() {
                    app.spoteditorstate.visible = false
                    app.rivereditorstate.visible = false;
                    app.regioneditorstate.visible = false;

                    app.rivereditorstate.visible = true;
                    app.rivereditorstate.editMode = true;
                    app.rivereditorstate.river = {
                        id: 0,
                        region: this.region,
                        aliases: [],
                    }
                },

                regions: getAllRegions(),
                parseAliases:function(strVal) {
                    return strVal.split('\n').map(function(x) {return x.trim()}).filter(function(x){return x.length>0})
                },
            }
        }
    }

</script>