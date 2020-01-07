import {WWMapPopup} from "./popup";
import {canEdit} from "./util";
import {loadFragment} from "./template-data";

var $ = require("jquery");
require("./contrib/jquery.tmpl");

export function RiverList(divId, templateDivId, fromTemplates) {
    this.divId = divId;
    var t = this;

    loadFragment(templateDivId).then(templateText => {
        $('body').prepend(`<div id="${templateDivId}" style="display: none">${templateText}</div>`);
        t.templateDiv = $('#' + templateDivId);
    });

    this.riverInfoPopup = new WWMapPopup('river_desc_template', true, "river_desc");
}

RiverList.prototype.update = function (rivers) {
    if (this.templateDiv) {
        rivers.canEdit = canEdit();
        var html = this.templateDiv.tmpl(rivers).html();
        $('#' + this.divId).html(html)
    }
};