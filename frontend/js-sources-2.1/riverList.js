import {WWMapPopup} from "./popup";
import {canEdit, loadFragment} from "./util";
import {MAP_FRAGMENTS_URL} from './config';

var $ = require("jquery");
require("./contrib/jquery.tmpl");

export function RiverList(divId, templateDivId, fromTemplates) {
    this.divId = divId;
    var t = this;

    if (fromTemplates) {
        loadFragment(MAP_FRAGMENTS_URL, templateDivId, function (templateText) {
            $('body').prepend(templateText);
            t.templateDiv = $('#' + templateDivId)
        })
    } else {
        t.templateDiv = $('#' + templateDivId)
    }

    this.riverInfoPopup = new WWMapPopup('river_desc_template', true, "river_desc");
}

RiverList.prototype.update = function (rivers) {
    if (this.templateDiv) {
        rivers.canEdit = canEdit();
        var html = this.templateDiv.tmpl(rivers).html();
        $('#' + this.divId).html(html)
    }
};