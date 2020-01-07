import {initMailtoLinks} from "./util"
import {loadFragment} from "./template-data";

var $ = require("jquery");
import Template7 from "template7";

export function WWMapPopup(templateId, fromTemplates, workingDivId, options) {
    this.divId = workingDivId;
    if (fromTemplates) {
        loadFragment(templateId).then(templateText => {
            t.template = Template7.compile(templateText);
        });
    } else {
        throw "Not implemented"
    }

    this.submitUrl = (options) ? options.submitUrl : null;
    this.okMsg = (options) ? options.okMsg : null;
    this.failMsg = (options) ? options.failMsg : null;

    $('body').prepend('<div id="' + this.divId + '" class="wwmap-popup_area"></div>');
    this.div = $("#" + this.divId);

    var t = this;

    // close on mouse click outside the window
    if (!options || options.closeOnMouseClickOutside !== false) {
        this.div.click(function (source) {
            var classAttr = $(source.target).attr('class');
            if (classAttr && classAttr.indexOf('wwmap-popup_are') > -1) {
                t.hide()
            }
        });
    }

    // close on escape pressed
    if (!options || options.closeOnEscape !== false) {
        $(document).keyup(function (e) {
            if (e.which == 27) { // Escape
                t.hide()
            }
        });
    }
}

WWMapPopup.prototype.show = function (dataObject) {
    var t = this;

    var html = "";
    if (dataObject) {
        html = this.template(dataObject)
    } else {
        html = this.template({})
    }

    this.div.html(html);
    $("#" + this.divId + " input[name=cancel]").click(function () {
        return t.hide()
    });
    $("#" + this.divId + " input[type=submit]").click(function () {
        return t.submit_form()
    });

    initMailtoLinks();
    this.div.addClass("wwmap-show");
};

WWMapPopup.prototype.hide = function () {
    this.div.html('');
    this.div.removeClass("wwmap-show");
};

WWMapPopup.prototype.submit_form = function () {
    var form = $("#" + this.divId + " form");
    var t = this;
    $.post(this.submitUrl, form.serialize())
        .done(function (data) {
            window.alert(t.okMsg);
            t.hide();
            form.trigger('reset')
        }).fail(function () {
        window.alert(t.failMsg);
    });
    return false;
};
