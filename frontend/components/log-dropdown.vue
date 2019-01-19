<style scoped>
    .wwmap-log-menu.show {
        width: auto !important;
        left: auto !important;
        right: 0 !important;
        padding: 20px;
    }
</style>

<template>
    <div v-if="admin()" class="dropdown" style="display: inline-block;">
        <button class="btn btn-secondary dropdown-toggle" type="button" id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
            Лог изменений
        </button>
        <div class="dropdown-menu wwmap-log-menu">
            <log :object-type="objectType" :object-id="objectId"/>
        </div>
    </div>
</template>

<script>
    module.exports = {
        props: {
            objectType: String,
            objectId: Number
        },
        data: function () {
            return {
                admin: function () {
                    var userInfo = getAuthorizedUserInfoOrNull();
                    return userInfo && userInfo.roles && userInfo.roles.indexOf("ADMIN") > -1;
                }
            }
        }
    }

</script>