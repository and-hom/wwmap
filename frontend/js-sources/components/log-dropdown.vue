<style scoped>
    .wwmap-log-menu.show {
        width: auto !important;
        left: auto !important;
        right: 0 !important;
        padding: 20px;
    }
</style>

<template>
    <div v-if="admin" class="dropdown" style="display: inline-block;">
        <button class="btn btn-secondary dropdown-toggle" type="button" id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
            Лог изменений
        </button>
        <div class="dropdown-menu wwmap-log-menu">
            <log :object-type="objectType" :object-id="objectId"/>
        </div>
    </div>
</template>

<script>
    import {hasRole, ROLE_ADMIN} from "../auth";

    module.exports = {
        props: {
            objectType: String,
            objectId: Number
        },
        created: function() {
            hasRole(ROLE_ADMIN).then(admin => this.admin = admin);
        },
        data: function () {
            return {
                admin: false,
            }
        }
    }

</script>