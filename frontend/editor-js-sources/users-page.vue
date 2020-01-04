<template>
    <page link="users.htm">
        <ask-role id="ask-role" :roles="['ADMIN', 'EDITOR', 'USER']" :ok-fn="function(userId, role) { setRole(userId, role) }"></ask-role>

        <div class="container-fluid" v-if="admin">
            <div class="row">
                <div class="col-12">
                    <table class="table">
                        <thead>
                        <tr>
                            <td>
                                Id
                            </td>
                            <td>
                                Логин
                            </td>
                            <td>
                                Провайдер авторизации/Удалённый Id
                            </td>
                            <td>
                                Имя
                            </td>
                            <td>
                                Роли
                            </td>
                            <td></td>
                        </tr>
                        </thead>
                        <tr v-for="user in users" :class="rowClass(user)">
                            <td>
                                {{user.id}}
                            </td>
                            <td>
                                {{user.info.login}}
                            </td>
                            <td>
                                {{user.auth_provider}}/{{user.ext_id}}
                            </td>
                            <td>
                                {{user.info.first_name}} {{user.info.last_name}}
                            </td>
                            <td :class="roleClass(user)">
                                {{user.role}}
                            </td>
                            <td style="white-space: nowrap;">
                                <button class="btn btn-primary" data-toggle="modal" data-target="#ask-role"
                                        :data-user-id="user.id" :data-current-role="user.role" :data-label="roleChangeText(user)">Сменить роль</button>
                                <button class="btn btn-primary" style="width: 150px;" v-on:click="toggleExperimental(user)">{{experimentalFeatureSwitchText(user)}}</button>
                                <log-dropdown object-type="USER" :object-id="user.id"/>
                            </td>
                        </tr>
                    </table>
                </div>
            </div>
        </div>
    </page>
</template>

<script>
    import {doGetJson, doPostJson} from "./api";
    import {hasRole} from "./auth";
    import {backendApiBase} from "./config"

    export default {
        created() {
            doGetJson(backendApiBase + "/user", true).then(users => this.users = users);
            hasRole('ADMIN').then(admin => this.admin = admin);
        },
        data() {
            return {
                users: [],
                availableRoles: ["ADMIN", "EDITOR", "USER"],
                admin: false,
                roleClass: function (user) {
                    return "role-" + user.role.toLowerCase()
                },
                setRole: function (userId, role) {
                    doPostJson(backendApiBase + '/user/' + userId + '/role', role, true).then(users => this.users = users);
                },
                toggleExperimental: function (user) {
                    doPostJson(backendApiBase + '/user/' + user.id + '/experimental', !user.experimental_features, true).then(users => this.users = users);
                },
                roleChangeText: function (user) {
                    return 'Сменить роль для ' + user.info.login + '. Текущая роль - ' + user.role
                },
                experimentalFeatureSwitchText: function (user) {
                    return user.experimental_features
                        ? "Выкл эксперимент"
                        : "Вкл эксперимент"
                },
                rowClass: function (user) {
                    let cssClass = "wwmap-user-row";
                    if (user.experimental_features) {
                        cssClass += " wwmap-user-row-experimental"
                    }
                    return cssClass
                },
            }
        }
    }
</script>
