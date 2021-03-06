import Vue from 'vue'
import {store} from './app-state'

import EditorPage from './pages/editor-page.vue'
import DocsIntegrationPage from './pages/docs-integration-page.vue'
import MapPage from './pages/map-page.vue'
import LogPage from './pages/log-page.vue'
import UsersPage from './pages/users-page.vue'
import TechPage from './pages/tech-page.vue'
import AboutPage from './pages/about-page.vue'
import DocsPage from './pages/docs-page.vue'
import LevelPage from './pages/level-page.vue'
import SitesPage from './pages/sites-page.vue'
import JobsPage from './pages/jobs-page.vue'
import TimelinePage from './pages/timeline-page.vue'
import TransferPage from './pages/transfer-page.vue'
import CampPage from './pages/camp-page.vue'
import VoyageReportPage from './pages/voyage-report-page.vue'
import PackageChangelog from './pages/package-changelog-page.vue'

import vSelect from 'vue-select'
import VueGallery from 'vue-gallery';
import FileUpload from 'vue-upload-component';
import Datepicker from 'vuejs-datepicker';
import {ImageRating} from 'vue-rate-it';
import VueGoogleCharts from 'vue-google-charts'
import VueTagsInput from '@johmun/vue-tags-input';
import LoadScript from 'vue-plugin-load-script';

import VueSlider from 'vue-slider-component'
import 'vue-slider-component/theme/default.css'

import upperFirst from 'lodash/upperFirst'
import camelCase from 'lodash/camelCase'

import {TabsPlugin} from 'bootstrap-vue'
import {getCountries} from './editor'
import {acquireTokenVk, getTokenFromRequestAndStartWwmapSession, startWwmapSession} from './auth'
import {parseParams} from './api'
import 'bootstrap/dist/css/bootstrap.min.css';
import 'vue-select/dist/vue-select.css';
import 'codemirror/lib/codemirror.css';
import '@toast-ui/editor/dist/toastui-editor.css';
import '@toast-ui/editor/dist/toastui-editor-viewer.css';

import './style/main.css'
import './style/editor.css'
import {mapJsApiUrl} from "./config";

const moment = require('moment');

require("bootstrap");

require('babel-polyfill');

const VueScrollTo = require('vue-scrollto');

export var app;

export function initEditor() {
    return init(EditorPage)
}

export function initDocsIntegration() {
    return init(DocsIntegrationPage)
}

export function initMap() {
    return init(MapPage)
}

export function initLog() {
    return init(LogPage)
}

export function initUsers() {
    return init(UsersPage)
}

export function initTech() {
    return init(TechPage)
}

export function initAbout() {
    return init(AboutPage)
}

export function initDocs() {
    return init(DocsPage)
}

export function initLevel() {
    return init(LevelPage)
}

export function initSites() {
    return init(SitesPage)
}

export function initJobs() {
    return init(JobsPage)
}

export function initTimeline() {
    return init(TimelinePage)
}

export function initTransfer() {
    return init(TransferPage)
}

export function initCamp() {
    return init(CampPage)
}

export function initVoyageReport() {
    return init(VoyageReportPage)
}

export function initPackageChangelog() {
    return init(PackageChangelog)
}

function init(page) {
    Vue.component('v-select', vSelect);
    Vue.component('gallery', VueGallery);
    Vue.component('file-upload', FileUpload);
    Vue.component('datepicker', Datepicker);
    Vue.component('image-rating', ImageRating);
    Vue.use(TabsPlugin);
    Vue.use(VueGoogleCharts);
    Vue.use(VueTagsInput);
    Vue.use(VueScrollTo);
    Vue.use(LoadScript);
    Vue.component('VueSlider', VueSlider);

    Vue.filter('formatDateTimeStr', function (value) {
        if (value) {
            return moment(String(value)).format('YYYY-MM-DD HH:mm:ss')
        }
    });

    Vue.filter('formatDateStr', function (value) {
        if (value) {
            return moment(String(value)).format('YYYY-MM-DD')
        }
    });

    Vue.filter('yearOnly0102', function (value) {
        if (value.endsWith('-01-02')) {
            return value.substring(0, 4);
        } else {
            return value;
        }
    });

    Vue.filter('emptyDate00010101', function (value) {
        if (value && value.startsWith('0001-01-01')) {
            return null;
        } else {
            return value;
        }
    });

    Vue.filter('orElse', function (value, emptyPlaceholder) {
        if (value) {
            return value;
        } else {
            return emptyPlaceholder
        }
    });

    const requireComponent = require.context('./components', true, /.+?\.(vue|js)$/);

    requireComponent.keys().forEach(fileName => {
        const componentConfig = requireComponent(fileName);
        const componentName = upperFirst(
            camelCase(fileName
                .split('/')
                .pop()
                .replace(/\.\w+$/, '')
            )
        );

        Vue.component(componentName, componentConfig.default || componentConfig)
    });

    getCountries().then(countries => {
        store.commit('setTreePath', countries)
    });

    Vue.loadScript(mapJsApiUrl)
        .catch(err => console.error(err))
        .finally(() => {
            app = new Vue({
                el: '#vue-app',
                render: h => h(page),
            });
        });
}

export function getApp() {
    return app
}


// Auth

export function auth_parseParams(paramsStr) {
    return parseParams(paramsStr)
}

export function auth_getTokenFromRequestAndStartWwmapSession(authSource, callback) {
    return getTokenFromRequestAndStartWwmapSession(authSource).then(_ => callback())
}

export function auth_acquireTokenVk(code, callback) {
    acquireTokenVk(code, callback)
}

export function auth_startWwmapSession(source, token, callback) {
    startWwmapSession(source, token).then(_ => callback())
}