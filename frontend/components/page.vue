<template>
<div>
    <nav class="navbar navbar-expand-lg navbar-light bg-light">
        <ul class="navbar-nav mr-auto">
            <li :class="pageClass(link)" v-if="showMenuItem(page)" v-for="(page, link) in pages">
                <a class="nav-link" :href="pageLink(link)">{{pageTitle(page)}}</a>
            </li>
        </ul>
        <auth></auth>
    </nav>
    <slot></slot>
</div>
</template>

<script>
    module.exports = {
        props: ['link'],
        data: function () {
            return {
                pages: {
                    "editor.htm": {
                        title: "Редактор",
                        allow: ['USER', 'EDITOR', 'ADMIN'],
                    },
                    "map.htm": "Карта",
                    "refs.htm": "Сайты",
                    "users.htm": {
                        title: "Пользователи",
                        allow: ['ADMIN'],
                    },
                    "docs.htm": {
                        title: "Прочитай меня",
                        allow: ['EDITOR', 'ADMIN'],
                    },
                },
                showMenuItem: function(page) {
                    if (typeof page == 'string' || page instanceof String) {
                        return true
                    }
                    if (!page.allow) {
                        return true
                    }

                    userInfo = getAuthorizedUserInfoOrNull()
                    if (userInfo==null || userInfo.roles==null) {
                        return false
                    }

                    return page.allow.filter(r => userInfo.roles.includes(r)).length>0
                },
                pageTitle: function(page) {
                    if (typeof page === 'string' || page instanceof String) {
                        return page
                    }
                    return page.title
                },
                pageLink: function(link) {
                    if (link==this.link) {
                        return "#"
                    }
                    return link
                },
                pageClass: function(link) {
                    if (link==this.link) {
                        return "nav-item"
                    }
                    return "nav-item active"
                },
            }
        },
    }
</script>