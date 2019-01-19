<template>
    <div class="auth" v-if="userInfo">
        <div>Привет, {{userName()}}!</div>
        <a href="javascript:clearSessionId(); location.reload();">Выход</a>
        <div style="float:right; color: grey; font-size:60%; padding-left:10px;"><strong>{{userInfo.auth_provider}}/</strong>{{userInfo.login}}</div>
    </div>
    <div class="auth" v-else>
        <div>Здравствуйте! Для редактирования надо</div>
        авторизоваться через <a href="javascript:YANDEX_AUTH.authRedirect();">Яндекс</a>, <a href="javascript:GOOGLE_AUTH.authRedirect();">Google</a> или <a href="javascript:VK_AUTH.authRedirect();">ВК</a>
    </div>
</template>

<script>
    module.exports = {
        props: ['id', 'msg', 'title', 'okfn'],
        data: function () {
            return {
                userInfo: getAuthorizedUserInfoOrNull(),
                userName: function () {
                    if (this.userInfo.first_name || this.userInfo.last_name) {
                        return [this.userInfo.first_name, this.userInfo.last_name].join('\xa0')
                    }
                    return this.userInfo.login
                }
            }
        }
    }
</script>

<style>
</style>