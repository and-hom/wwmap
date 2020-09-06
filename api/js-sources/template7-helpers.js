import Template7 from "./contrib/template7";

export function regiterTemplate7Helpers() {
    let showdown = require('showdown');
    let converter = new showdown.Converter();

    Template7.registerHelper('md', function (md) {
        return converter.makeHtml(md)
    });
}