<template>
    <div class="auth" v-if="userInfo">
        <div>Привет, {{userInfo.first_name}}&nbsp;{{userInfo.last_name}}!</div>
        <a href="javascript:clearSessionId(); location.reload();">Выход</a>
        <div style="float:right; color: grey; font-size:60%;"><strong>{{userInfo.auth_provider}}/</strong>{{userInfo.login}}</div>
    </div>
    <div class="auth" v-else>
        <div>Здравствуйте! Для редактирования надо</div>
        авторизоваться через <a href="javascript:YANDEX_AUTH.authRedirect();">Яндекс</a>, <a href="javascript:GOOGLE_AUTH.authRedirect();">Google</a> или <a href="javascript:VK_AUTH.authRedirect();">ВК</a>
    </div>
</template>

<script>
    module.exports = {
        props: ['id', 'msg', 'title', 'okfn'],
        data: function() {
            return {
                    userInfo: getAuthorizedUserInfoOrNull()
            }
        }
    }
</script>

<style>
</style>