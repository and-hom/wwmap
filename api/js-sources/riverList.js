import {WWMapPopup} from "./popup";
import {canEdit} from "./util";
import {loadFragment} from "./template-data";

var $ = require("jquery");
import Template7 from "template7";

export function RiverList(divId, templateDivId, fromTemplates) {
    this.divId = divId;
    var t = this;

    if (fromTemplates) {
        loadFragment(templateDivId).then(templateText => {
            t.template = Template7.compile(templateText);
        });
    } else {
        throw "Not implemented"
    }

    this.riverInfoPopup = new WWMapPopup('river_desc_template', true, "river_desc");
}

RiverList.prototype.update = function (rivers) {
    if (this.template) {
        canEdit().then(canEdit => {
            rivers.canEdit = canEdit;
            var html = this.template(rivers);
            $('#' + this.divId).html(html);
        });
    }
};