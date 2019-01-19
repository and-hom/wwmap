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
    <table v-if="admin()" class="wwmap-changes-log">
        <tr style="vertical-align: top;" v-for="entry in logEntries()">
            <td style="white-space: nowrap">{{entry.time}}</td>
            <td>{{entry.login}}<div class="wwmap-system-hint">{{entry.auth_provider}}/{{entry.ext_id}}</div></td>
            <td>{{entry.type}}</td>
            <td style="max-width: 200px;">{{entry.description}}</td>
        </tr>
    </table>
</template>

<script>
    module.exports = {
        props: {
            objectType: String,
            objectId: Number
        },
        data: function () {
            return {
                logEntries: function () {
                    return getLogEntries(this.objectType, this.objectId);
                },
                admin: function () {
                    var userInfo = getAuthorizedUserInfoOrNull();
                    return userInfo && userInfo.roles && userInfo.roles.indexOf("ADMIN") > -1;
                }
            }
        }
    }

</script>