export function initLayoutFilters() {
    let showdown = require('showdown');
    let converter = new showdown.Converter();
    ymaps.template.filtersStorage.add('md', (dm, md) => converter.makeHtml(md));
}