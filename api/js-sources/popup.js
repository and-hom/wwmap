import {initMailtoLinks} from "./util"
import {loadFragment} from "./template-data";

var $ = require("jquery");
import Template7 from "./contrib/template7";
import {frontendBase} from "./config";

export function WWMapPopup(divId, templateId, options) {
    this.divId = divId;
    let optionsIsObject = typeof options == 'object';
    if (optionsIsObject && options.templateData) {
        t.template = Template7.compile(options.templateData);
    } else {
        loadFragment(templateId).then(templateText => {
            t.template = Template7.compile(templateText);
        });
    }

    this.submitUrl = optionsIsObject ? options.submitUrl : null;
    this.okMsg = optionsIsObject ? options.okMsg : null;
    this.failMsg = optionsIsObject ? options.failMsg : null;

    $('body').prepend('<div id="' + this.divId + '" class="wwmap-popup_area"></div>');
    this.div = $("#" + this.divId);

    var t = this;

    // close on mouse click outside the window
    if (!optionsIsObject || options.closeOnMouseClickOutside !== false) {
        this.div.click(function (source) {
            var classAttr = $(source.target).attr('class');
            if (classAttr && classAttr.indexOf('wwmap-popup_are') > -1) {
                t.hide()
            }
        });
    }

    // close on escape pressed
    if (!optionsIsObject || options.closeOnEscape !== false) {
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
        dataObject.frontendBase = frontendBase;
        html = this.template(dataObject)
    } else {
        html = this.template({frontendBase: frontendBase})
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
