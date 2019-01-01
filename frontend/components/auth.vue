<template>
    <div class="auth" v-if="userInfo">
        <div>Привет, {{userInfo.first_name}}&nbsp;{{userInfo.last_name}}!</div>
        <a href="javascript:clearToken(); location.reload();">Выход</a>
        <div style="float:right; color: grey; font-size:60%;"><strong>{{userInfo.auth_provider}}/</strong>{{userInfo.login}}</div>
    </div>
    <div class="auth" v-else>
        <div>Здравствуйте! Для редактирования надо</div>
        авторизоваться через <a href="javascript:forceRedirectYndx();">Яндекс</a>, <a href="javascript:forceRedirectGoogle();">Google</a> или <a href="javascript:forceRedirectVk();">ВК</a>
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