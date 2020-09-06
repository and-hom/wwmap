export const backendApiBase = "http://localhost:7007";
export const tCacheApiBase = "http://localhost:7008";
export const tCacheVersionUrl = "http://localhost:7008/maps/version";
export const cronApiBase = "http://localhost:7009";
export const frontendBase = "http://localhost:63342/wwmap/frontend/";
export const frontendVersion = "development";
export const changelogPathTemplate = "../{0}/debian/changelog"
export const mapJsApiUrl = "../api/js/map.v2.1.js"

export const markdownEditorConfig = {
    minHeight: '200px',
    language: 'ru-RU',
    useCommandShortcut: true,
    useDefaultHTMLSanitizer: true,
    usageStatistics: false,
    hideModeSwitch: false,
    toolbarItems: [
        'heading',
        'bold',
        'italic',
        // 'strike', // table strike: not supported by shutdown converter
        'divider',
        'hr',
        // 'quote', // table strike: not supported by shutdown converter
        'divider',
        'ul',
        'ol',
        // 'task',
        // 'indent',
        // 'outdent',
        'divider',
        // 'table', // table disabled: not supported by shutdown converter
        // 'image',
        'link',
        // 'divider',
        // 'code',
        // 'codeblock'
    ]
}
