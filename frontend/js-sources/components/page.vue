<template>
    <div v-if="rolesLoaded">
        <nav class="navbar navbar-expand-lg navbar-light bg-light">
            <ul class="navbar-nav mr-auto">
                <li :class="itemClass(page)" v-if="showMenuItem(page)" v-for="page in pages">
                    <a v-if="page.sub==null"
                       :style="page.href=='map.htm' ? 'font-size: x-large; padding-top: 1px; padding-bottom: 1px;' : ''"
                       class="nav-link" :href="pageLink(page.href)">{{pageTitle(page)}}</a>
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
        <footer class="footer">
            <div class="wwmap-footer">
                <span v-if="showTechInfo">
                    <span>Версия бэкенда:&nbsp;<a target="_blank" :href="'./package-changelog.htm?module=' + 'backend'">{{backVersion}}</a></span>
                    <span>Версия фронтенда:&nbsp;<a target="_blank" :href="'./package-changelog.htm?module=' + 'frontend'">{{frontVersion}}</a></span>
                    <span>Версия api карты:&nbsp;<a target="_blank" :href="'./package-changelog.htm?module=' + 'api'">{{apiVersion}}</a></span>
                    <span>Версия t-cache:&nbsp;<a target="_blank" :href="'./package-changelog.htm?module=' + 't-cache'">{{tCacheVersion}}</a></span>
                    <span>Версия cron:&nbsp;<a target="_blank" :href="'./package-changelog.htm?module=' + 'cron'">{{cronVersion}}</a></span>
                    <span>Версия базы:&nbsp;<a target="_blank" :href="'./package-changelog.htm?module=' + 'db'">{{dbVersion}}</a></span>
                </span><span>Контакт для связи:&nbsp;<a
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
        margin-right: 0;
    }

    .wwmap-footer span {
        margin-right: 30px;
    }
</style>

<script>
    import {frontendVersion} from '../config'
    import {getBackendVersion, getTCacheVersion, getCronVersion, getDbVersion} from '../api'
    import {getRoles, ROLE_ADMIN, ROLE_EDITOR} from '../auth'


    module.exports = {
        props: ['link'],
        created: function () {
            getDbVersion().then(version => this.dbVersion = version);
            getBackendVersion().then(version => this.backVersion = version);
            getTCacheVersion().then(version => this.tCacheVersion = version);
            getCronVersion().then(version => this.cronVersion = version);
            let x = getRoles();
            x.then(roles => {
                this.roles = roles;
                this.showTechInfo = roles.includes(ROLE_ADMIN) || roles.includes(ROLE_EDITOR);
                this.rolesLoaded = true;
            });
        },
        data: function () {
            return {
                rolesLoaded: false,
                showTechInfo: false,
                roles: [],
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
                        id: "2",
                        title: "Разное",
                        allow: ['EDITOR', 'ADMIN'],
                        sub: [
                            {
                                href: "transfer.htm",
                                title: "Заброски",
                            },
                        ]
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
                        allow: ['ADMIN'],
                        sub: [
                            {
                                href: "users.htm",
                                title: "Пользователи",
                            },
                            {
                                href: "level.htm",
                                title: "Уровни воды",
                            },
                            {
                                href: "sites.htm",
                                title: "Сайты, на которых размещена карта",
                            },
                            {
                                href: "log.htm",
                                title: "История изменений",
                            },
                        ]
                    },
                    {
                        id: "3",
                        title: "Cron",
                        allow: ['ADMIN'],
                        sub: [
                            {
                                href: "jobs.htm",
                                title: "Задачи",
                            },
                            {
                                href: "timeline.htm",
                                title: "Таймлайн и логи",
                            },
                        ]
                    },
                    {
                        id: "1",
                        title: "Информация",
                        sub: [
                            {
                                href: "docs.htm",
                                title: "Прочитай меня",
                                allow: ['EDITOR', 'ADMIN'],
                            },
                            {
                                href: "tech.htm",
                                title: "Технологии, источники данных",
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
                    if (!page.allow || page.allow.includes('ANONYMOUS') && this.roles != null && this.roles.length == 0) {
                        return true
                    }

                    return page.allow
                        .filter(r => this.roles.includes(r))
                        .length > 0
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
                        let t = this
                        return page.sub.filter(s => s.href == t.link).length == 0
                            ? "nav-item dropdown"
                            : "nav-item dropdown active";
                    }
                    return "nav-item"
                },
                subItemClass: function (sub) {
                    if (sub.href === this.link) {
                        return "dropdown-item active"
                    }
                    return "dropdown-item"
                },
                backVersion: '–',
                apiVersion: typeof wwmap != 'undefined' && wwmap && wwmap.version ? wwmap.version : '–',
                dbVersion: '–',
                frontVersion: frontendVersion,
                tCacheVersion: '–',
                cronVersion: '–',
            }
        },
    }
</script>