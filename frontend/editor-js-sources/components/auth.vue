<template>
    <div class="auth" v-if="userInfo">
        <div>Привет, {{userName}}!</div>
        <a href="javascript:clearSessionId(); location.reload();">Выход</a>
        <div style="float:right; color: grey; font-size:60%; padding-left:10px;">
            <strong>{{userInfo.auth_provider}}/</strong>{{userInfo.login}}
        </div>
    </div>
    <div class="auth" v-else>
        <div>Здравствуйте! Для редактирования надо</div>
        авторизоваться через <a href="javascript:YANDEX_AUTH.authRedirect();">Яндекс</a>, <a
            href="javascript:GOOGLE_AUTH.authRedirect();">Google</a> или <a
            href="javascript:VK_AUTH.authRedirect();">ВК</a>
    </div>
</template>

<script>
    import {getAuthorizedUserInfoOrNull} from '../auth'

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
            }
        }
    }
</script>

<style>
</style>