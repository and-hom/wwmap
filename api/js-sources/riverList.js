import {WWMapPopup} from "./popup";
import {canEdit} from "./util";
import {loadFragment} from "./template-data";

var $ = require("jquery");
import Template7 from "./contrib/template7";

export function RiverList(divId, templateDivId, riversTemplateData) {
    this.divId = divId;
    var t = this;

    if (riversTemplateData) {
        t.template = Template7.compile(riversTemplateData);
    } else {
        loadFragment(templateDivId).then(templateText => {
            t.template = Template7.compile(templateText);
        });
    }

    this.riverInfoPopup = new WWMapPopup("river_desc", 'river_desc_template');
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