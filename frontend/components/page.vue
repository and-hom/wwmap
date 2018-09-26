<template>
<div>
    <nav class="navbar navbar-expand-lg navbar-light bg-light">
        <ul class="navbar-nav mr-auto">
            <li :class="pageClass(page.href)" v-if="showMenuItem(page)" v-for="page in pages">
                <a class="nav-link" :href="pageLink(page.href)">{{pageTitle(page)}}</a>
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
                pages: [
                    {
                        href:  "editor.htm",
                        title: "Редактор",
                        allow: ['EDITOR', 'ADMIN'],
                    },
                    {
                        href:  "editor.htm",
                        title: "Каталог",
                        allow: ['USER', 'ANONYMOUS'],
                    },
                    {
                        href: "map.htm",
                        title: "Как карта видна окружающим",
                        allow: ['EDITOR', 'ADMIN'],
                    },
                    {
                        href: "map.htm",
                        title: "Карта",
                        allow: ['USER', 'ANONYMOUS'],
                    },
                    {
                        href: "refs.htm",
                        title: "Сайты"
                    },
                    {
                        href: "users.htm",
                        title: "Пользователи",
                        allow: ['ADMIN'],
                    },
                    {
                        href: "docs.htm",
                        title: "Прочитай меня",
                        allow: ['EDITOR', 'ADMIN'],
                    },
                ],
                showMenuItem: function(page) {
                    if (typeof page == 'string' || page instanceof String) {
                        return true
                    }
                    if (!page.allow) {
                        return true
                    }

                    userInfo = getAuthorizedUserInfoOrNull()
                    if (userInfo==null || userInfo.roles==null) {
                        return page.allow.filter(r => r=='ANONYMOUS').length>0
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