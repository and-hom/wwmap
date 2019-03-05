<template>
    <div>
        <nav class="navbar navbar-expand-lg navbar-light bg-light">
            <ul class="navbar-nav mr-auto">
                <li :class="itemClass(page)" v-if="showMenuItem(page)" v-for="page in pages">
                    <a v-if="page.sub==null" class="nav-link" :href="pageLink(page.href)">{{pageTitle(page)}}</a>
                    <a v-else class="nav-link dropdown-toggle" :id="page.id" role="button" data-toggle="dropdown"
                       aria-haspopup="true" aria-expanded="false">{{page.title}}</a>
                    <div v-if="page.sub" class="dropdown-menu" :aria-labelledby="page.id">
                        <a v-for="sub in page.sub" v-if="showMenuItem(sub)" :class="subItemClass(sub)"
                           :href="pageLink(sub.href)">{{sub.title}}</a>
                    </div>
                </li>
            </ul>
            <auth></auth>
        </nav>
        <slot></slot>
        <footer v-if="showFooter" class="footer">
            <div class="container wwmap-footer">
                <span>Версия бэкенда:&nbsp;{{backVersion}}</span><span>Версия фронтенда:&nbsp;{{frontVersion}}</span><span>Контакт для связи:&nbsp;<a
                    href="mailto:info@wwmap.ru">info@wwmap.ru</a></span>
            </div>
        </footer>
    </div>

</template>
<style type="text/css">
    html {
        position: relative;
        min-height: 100%;
    }

    body {
        /* Margin bottom by footer height */
        margin-bottom: 60px;
    }

    .footer {
        position: absolute;
        bottom: 0;
        width: 100%;
        height: 60px;
        line-height: 60px; /* Vertically center the text there */
        background-color: #f5f5f5;
    }

    body > .wwmap-footer {
        padding: 60px 15px 0;
    }

    .footer > .wwmap-footer {
        padding-right: 15px;
        padding-left: 15px;
        margin-left: 0;
    }

    .wwmap-footer span {
        margin-right: 30px;
    }
</style>

<script>
    module.exports = {
        props: ['link'],
        computed: {
            showFooter: function () {
                var userInfo = getAuthorizedUserInfoOrNull();
                return userInfo && ['EDITOR', 'ADMIN'].filter(function (r) {
                    return userInfo.roles.includes(r)
                }).length > 0
            }
        },
        data: function () {
            return {
                pages: [
                    {
                        href: "editor.htm",
                        title: "Редактор",
                        allow: ['EDITOR', 'ADMIN'],
                    },
                    {
                        href: "editor.htm",
                        title: "Каталог",
                        allow: ['USER', 'ANONYMOUS'],
                    },
                    {
                        href: "map.htm",
                        title: "Карта",
                    },
                    {
                        href: "docs-integration.htm",
                        title: "Карта на свой сайт"
                    },
                    {
                        id: "2",
                        title: "Администрирование",
                        allow: ['EDITOR', 'ADMIN'],
                        sub: [
                            {
                                href: "users.htm",
                                title: "Пользователи",
                            },
                            {
                                href: "dashboard.htm",
                                title: "Панель показателей",
                            },
                        ]
                    },
                    {
                        href: "about.htm",
                        title: "О проекте",
                        allow: ['USER', 'ANONYMOUS'],
                    },
                    {
                        id: "1",
                        title: "Информация",
                        allow: ['EDITOR', 'ADMIN'],
                        sub: [
                            {
                                href: "docs.htm",
                                title: "Прочитай меня",
                            },
                            {
                                href: "about.htm",
                                title: "О проекте",
                            },
                        ]
                    },
                ],
                showMenuItem: function (page) {
                    if (typeof page == 'string' || page instanceof String) {
                        return true
                    }
                    if (!page.allow) {
                        return true
                    }

                    var userInfo = getAuthorizedUserInfoOrNull()
                    if (userInfo == null || userInfo.roles == null) {
                        return page.allow.filter(function (r) {
                            return r == 'ANONYMOUS'
                        }).length > 0
                    }

                    return page.allow.filter(function (r) {
                        return userInfo.roles.includes(r)
                    }).length > 0
                },
                pageTitle: function (page) {
                    if (typeof page === 'string' || page instanceof String) {
                        return page
                    }
                    return page.title
                },
                pageLink: function (link) {
                    if (link == this.link) {
                        return "#"
                    }
                    return link
                },
                itemClass: function (page) {
                    if (page.href == this.link) {
                        return "nav-item active"
                    }
                    if (page.sub) {
                        return "nav-item dropdown"
                    }
                    return "nav-item"
                },
                subItemClass: function (sub) {
                    if (sub.href === this.link) {
                        return "dropdown-item active"
                    }
                    return "dropdown-item"
                },
                backVersion: getBackendVersion(),
                frontVersion: frontendVersion
            }
        },
    }
</script>