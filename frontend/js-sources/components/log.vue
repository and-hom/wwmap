<style scoped>
    .wwmap-changes-log td {
        margin-right: 5px;
        white-space: nowrap;
        vertical-align: top;
        font-size: 85%;
    }

    .wwmap-changes-log tr:nth-child(even) {
        background-color: #fcf8e3;
    }
    .wwmap-changes-log tr:nth-child(odd) {
        background-color: #f0fcf0;
    }
</style>

<template>
    <table v-if="admin" class="wwmap-changes-log">
        <tr style="vertical-align: top;" v-for="entry in logEntries">
            <td style="white-space: nowrap">{{entry.time}}</td>
            <td>
              <div>{{ entry.user_info.first_name }}</div>
              <div>{{ entry.user_info.last_name }}</div>
              <div class="wwmap-system-hint">{{entry.auth_provider}}/{{entry.ext_id}}</div>
            </td>
            <td>{{entry.type}}</td>
            <td style="max-width: 200px;">{{entry.description}}</td>
        </tr>
    </table>
</template>

<script>
    import {hasRole, ROLE_ADMIN} from "../auth";
    import {getLogEntries} from "../editor";

    module.exports = {
        props: {
            objectType: String,
            objectId: Number
        },
        created: function() {
            hasRole(ROLE_ADMIN).then(admin => {
                this.admin = admin;
                getLogEntries(this.objectType, this.objectId).then(entries => this.logEntries = entries);
            });
        },
        data: function () {
            return {
                logEntries: [],
                admin: false,
            }
        }
    }

</script>