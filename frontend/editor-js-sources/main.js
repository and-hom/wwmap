import Vue from 'vue'
import EditorPage from './editor-page.vue'

import VueSelect from 'vue-select'
import VueGallery from 'vue-gallery';
import FileUpload from 'vue-upload-component';

import upperFirst from 'lodash/upperFirst'
import camelCase from 'lodash/camelCase'

export function init() {
    function getById(arr, id) {
        var filtered = arr.filter(function (x) {
            return x.id === id
        });
        if (filtered.length > 0) {
            return filtered[0]
        }
        return null
    }

    function showCountrySubentities(id) {
        let country = {
            rivers: getRiversByCountry(id),
            regions: getRegions(id)
        };
        Vue.set(app.treePath, id, country);
        return country;
    }

    function showRegionTree(countryId, id) {
        var region = getById(app.treePath[countryId].regions, id);
        Vue.set(region, "rivers", getRiversByRegion(countryId, id));
    }

    function getRiverFromTree(countryId, regionId, id) {
        let river;
        let country = app.treePath[countryId];
        if (!country) {
            country = showCountrySubentities(countryId)
        }

        if (regionId && regionId > 0) {
            var region = getById(country.regions, regionId);
            let rivers = region.rivers;
            if (!rivers) {
                rivers = getRiversByRegion(countryId, region.id);
                Vue.set(region, "rivers", rivers);
            }
            river = getById(rivers, id)
        } else {
            river = getById(country.rivers, id)
        }
        return river;
    }

    function showRiverTree(countryId, regionId, id) {
        var river = getRiverFromTree(countryId, regionId, id);
        Vue.set(river, "spots", getSpots(id))
    }

    function hideRiverTree(countryId, regionId, id) {
        var river = getRiverFromTree(countryId, regionId, id);
        Vue.delete(river, "spots")
    }

    function getSpotsFromTree(countryId, regionId, riverId) {
        var river;
        if (regionId && regionId > 0) {
            var region = getById(app.treePath[countryId].regions, regionId);
            river = getById(region.rivers, riverId)
        } else {
            river = getById(app.treePath[countryId].rivers, riverId)
        }
        return river.spots
    }

    function setActiveEntityState(countryId, regionId, riverId, spotId) {
        app.selectedSpot = spotId;
        app.selectedRiver = riverId;
        app.selectedRegion = regionId;
        app.selectedCountry = countryId;
    }

    function newRiver(country, region) {
        app.spoteditorstate.visible = false;
        app.rivereditorstate.visible = false;
        app.regioneditorstate.visible = false;
        app.countryeditorstate.visible = false;

        app.rivereditorstate.visible = true;
        app.rivereditorstate.pageMode = 'edit';
        app.rivereditorstate.river = {
            id: 0,
            region: region,
            aliases: [],
            props: {}
        };
        app.rivereditorstate.country = country;
        app.rivereditorstate.region = region;
    }

    function selectRiver(country, region, id) {
        app.rivereditorstate.river = getRiver(id);
        app.rivereditorstate.pageMode = 'view';
        app.rivereditorstate.reports = getReports(id);
        app.rivereditorstate.country = country;
        app.rivereditorstate.region = region;
        app.rivereditorstate.visible = true;
    }

    function selectCountry(country) {
        app.countryeditorstate.country = country;
        app.countryeditorstate.editMode = false;
        app.countryeditorstate.visible = true
    }

    function selectRegion(country, id) {
        app.regioneditorstate.region = getRegion(id);
        app.regioneditorstate.country = country;
        app.regioneditorstate.editMode = false;
        app.regioneditorstate.visible = true
    }

    const requireComponent = require.context(
        // Относительный путь до каталога компонентов
        './components',
        // Обрабатывать или нет подкаталоги
        true,
        // Регулярное выражение для определения файлов базовых компонентов
        /.+?\.(vue|js)$/
    );

    Vue.component('v-select', VueSelect.VueSelect);
    Vue.component('gallery', VueGallery);
    Vue.component('file-upload', FileUpload);

    requireComponent.keys().forEach(fileName => {
        // Получение конфигурации компонента
        const componentConfig = requireComponent(fileName);

        // Получение имени компонента в PascalCase
        const componentName = upperFirst(
            camelCase(
                // Получаем имя файла независимо от глубины вложенности
                fileName
                    .split('/')
                    .pop()
                    .replace(/\.\w+$/, '')
            )
        );

        // Глобальная регистрация компонента
        Vue.component(
            componentName,
            // Поиск опций компонента в `.default`, который будет существовать,
            // если компонент экспортирован с помощью `export default`,
            // иначе будет использован корневой уровень модуля.
            componentConfig.default || componentConfig
        )
    });

    var app = new Vue({
        el: '#vue-app',
        render: h => h(EditorPage),
    });
}