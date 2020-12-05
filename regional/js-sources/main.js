import 'materialize-css/dist/css/materialize.min.css'
import 'materialize-css/dist/js/materialize.min.js'
import './style/index.css'

import $ from 'jquery';
import {mapJsApiUrl} from "../../frontend/js-sources/config";

global.jQuery = $;
global.$ = $;


jQuery.loadScript = function (url, callback) {
    jQuery.ajax({
        url: url,
        dataType: 'script',
        success: callback,
        async: true
    });
}

export function init(successCallback) {
    $.loadScript(mapJsApiUrl, successCallback)
}