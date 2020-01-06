<template>
    <page link="tech.htm">
        <div class="container-fluid">
            <div class="row">
                <div class="col-12">
                    <dl>
                        <dt>Яндекс-карты</dt>
                        <dd>Карта порогов отображаются с помощью <a target="_blank" href="https://tech.yandex.ru/maps%20/">api Яндекс-карт</a>. Используется слой спутниковых снимков.</dd>
                        <dt>OSM</dt>
                        <dd>Один из слоёв карты - <a target="_blank" href="https://www.openstreetmap.org">OpenStreetMap</a></dd>
                        <dt>Спутниковые снимк Google</dt>
                        <dd>Иногда важно увидеть порог на нескольких вариантах спутниковой съёмки, потому дополнительно подключены снимки от <a target="_blank" href="https://www.google.ru/maps">Google</a></dd>
                        <dt>Данные об уровнях воды Центра Регистра и Кадастра</dt>
                        <dd>Графики уровня воды взяты с <a target="_blank" href="http://gis.vodinfo.ru/">http://gis.vodinfo.ru/informer/</a></dd>
                        <dt>Источники отчётов</dt>
                        <dd><ul>
                            <li><a target="_blank" href="http://tlib.ru">tlib.ru</a></li>
                            <li><a target="_blank" href="https://huskytm.ru">huskytm.ru</a></li>
                            <li><a target="_blank" href="http://skitalets.ru">skitalets.ru</a></li>
                        </ul></dd>
                        <dt>Кроме того</dt>
                        <dd>Используются ссылки на Youtube и Яндекс-погоду</dd>
                    </dl>
                </div>
            </div>
        </div>
    </page>
</template>

<script>
    import {doGetJson, doPostJson} from "../api";
    import {hasRole} from "../auth";
    import {backendApiBase} from "../config"

    export default {
        created() {
            doGetJson(backendApiBase + "/user", true).then(users => this.users = users);
            hasRole('ADMIN').then(admin => this.admin = admin);
        },
        data() {
            return {
                users: [],
                availableRoles: ["ADMIN", "EDITOR", "USER"],
                admin: false,
                roleClass: function (user) {
                    return "role-" + user.role.toLowerCase()
                },
                setRole: function (userId, role) {
                    doPostJson(backendApiBase + '/user/' + userId + '/role', role, true).then(users => this.users = users);
                },
                toggleExperimental: function (user) {
                    doPostJson(backendApiBase + '/user/' + user.id + '/experimental', !user.experimental_features, true).then(users => this.users = users);
                },
                roleChangeText: function (user) {
                    return 'Сменить роль для ' + user.info.login + '. Текущая роль - ' + user.role
                },
                experimentalFeatureSwitchText: function (user) {
                    return user.experimental_features
                        ? "Выкл эксперимент"
                        : "Вкл эксперимент"
                },
                rowClass: function (user) {
                    let cssClass = "wwmap-user-row";
                    if (user.experimental_features) {
                        cssClass += " wwmap-user-row-experimental"
                    }
                    return cssClass
                },
            };
        }
    }
</script>
