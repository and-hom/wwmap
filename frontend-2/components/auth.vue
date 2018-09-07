<template>
    <div class="auth" v-if="userInfo">
        <div>Привет, {{userInfo.first_name}}&nbsp;{{userInfo.last_name}}!</div>
        <a href="javascript:clearToken(); location.reload();">Выход</a>
        <div style="float:right; color: grey; font-size:60%;">{{userInfo.login}}</div>
    </div>
    <div class="auth" v-else>
        <div>Здравствуйте! Для редактирования надо</div>
        <a href="javascript:forceRedirect();">авторизоваться через Яндекс</a>
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