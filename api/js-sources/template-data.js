export function loadFragment(templateId) {
    return new Promise((resolve, reject) => {
        try {
            resolve(require(`./templates/${templateId}.tmpl.htm`).default)
        } catch (e) {
            console.error(e);
            reject(e)
        }
    })
}