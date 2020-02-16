<template>
    <div class="auth" v-if="userInfo">
        <div>Привет, {{userName}}!</div>
        <a href="javascript:void(0);" v-on:click="clearSessionId">Выход</a>
        <div style="float:right; color: grey; font-size:60%; padding-left:10px;">
            <strong>{{userInfo.auth_provider}}/</strong>{{userInfo.login}}
        </div>
    </div>
    <div class="auth" v-else>
        <div>Здравствуйте! Для редактирования надо</div>
        авторизоваться через <a href="javascript:void(0);" v-on:click="yndxRedirect">Яндекс</a>, <a
            href="javascript:void(0);" v-on:click="googleRedirect">Google</a> или <a
            href="javascript:void(0);" v-on:click="vkRedirect">ВК</a>
    </div>
</template>

<script>
    import {getAuthorizedUserInfoOrNull, clearSessionId, YANDEX_AUTH, GOOGLE_AUTH, VK_AUTH} from '../auth'


    module.exports = {
        created: function() {
            let p = getAuthorizedUserInfoOrNull();
            if (p) {
                p.then(userInfo => this.userInfo = userInfo)
            }
        },
        computed: {
            userName: {
                get: function() {
                    if (this.userInfo.first_name || this.userInfo.last_name) {
                        return [this.userInfo.first_name, this.userInfo.last_name].join('\xa0')
                    }
                    return this.userInfo.login
                }
            }
        },
        data: function () {
            return {
                userInfo: null,
                clearSessionId: function () {
                    clearSessionId();
                    location.reload();
                },
                yndxRedirect: function () {
                    YANDEX_AUTH.authRedirect();
                },
                googleRedirect: function () {
                    GOOGLE_AUTH.authRedirect();
                },
                vkRedirect: function () {
                    VK_AUTH.authRedirect();
                },
            }
        }
    }
</script>

<style>
</style>