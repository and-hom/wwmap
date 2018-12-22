<template>
<div>
    <nav class="navbar navbar-expand-lg navbar-light bg-light">
        <ul class="navbar-nav mr-auto">
            <li :class="itemClass(page)" v-if="showMenuItem(page)" v-for="page in pages">
                <a v-if="page.sub==null" class="nav-link" :href="pageLink(page.href)">{{pageTitle(page)}}</a>
                <a v-else class="nav-link dropdown-toggle" :id="page.id" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">{{page.title}}</a>
                <div v-if="page.sub" class="dropdown-menu" :aria-labelledby="page.id">
                  <a v-for="sub in page.sub" v-if="showMenuItem(sub)" :class="subItemClass(sub)" :href="pageLink(sub.href)">{{sub.title}}</a>
                </div>
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
                showMenuItem: function(page) {
                    if (typeof page == 'string' || page instanceof String) {
                        return true
                    }
                    if (!page.allow) {
                        return true
                    }

                    userInfo = getAuthorizedUserInfoOrNull()
                    if (userInfo==null || userInfo.roles==null) {
                        return page.allow.filter(function(r) {return r=='ANONYMOUS'}).length>0
                    }

                    return page.allow.filter(function(r) {return userInfo.roles.includes(r)}).length>0
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
                itemClass: function(page) {
                    if (page.href==this.link) {
                        return "nav-item active"
                    }
                    if (page.sub) {
                        return "nav-item dropdown"
                    }
                    return "nav-item"
                },
                subItemClass: function(sub) {
                    if (sub.href==this.link) {
                        return "dropdown-item active"
                    }
                    return "dropdown-item"
                },
            }
        },
    }
</script>