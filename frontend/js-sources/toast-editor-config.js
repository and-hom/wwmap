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