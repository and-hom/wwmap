export function HashTool(delimiter, blankCell) {
    this.delimiter = delimiter ? delimiter : '/'
    this.blankCell = blankCell ? blankCell : ''
}

HashTool.prototype.getHashAtPos = function (pos, defaultValOrFunc, mapperFunc) {
    let value = null;
    let hash = window.location.hash;

    if (hash && hash.length > 1) {
        hash = hash.substr(1);
        let parts = hash.split(this.delimiter);
        if (parts.length > pos) {
            value = parts[pos]
        }
    }

    if (value) {
        return mapperFunc ? mapperFunc(value) : value
    } else if (defaultValOrFunc && typeof defaultValOrFunc === "function") {
        return defaultValOrFunc()
    } else {
        return defaultValOrFunc;
    }
}

HashTool.prototype.setHashAtPos = function (pos, value) {
    let hash = window.location.hash;
    let data = []
    if (hash && hash.length > 1) {
        hash = hash.substr(1);
        data = hash.split(this.delimiter);
    }

    let l = data.length;
    for (let i = l; i <= pos; i++) {
        data.push(this.blankCell)
    }
    data[pos] = value

    window.location.hash = data.join(this.delimiter)
}
