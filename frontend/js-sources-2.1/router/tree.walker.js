export function RiverTreeWalker(trackStorage, id, toId, maxSearchDepth) {
    this.toId = toId;
    this.trackStorage = trackStorage;
    this.maxSearchDepth = maxSearchDepth;
    this.elements = [{
        path: [],
        id: id,
    }];
}

RiverTreeWalker.prototype.searchRiver = function (lvl) {
    return this.searchRiverInternal(0);
};

//
RiverTreeWalker.prototype.searchRiverInternal = function (lvl) {
    if (lvl >= this.maxSearchDepth) {
        return null;
    }
    let result = [];
    for (let i = 0; i < this.elements.length; i++) {
        let element = this.elements[i];
        let river = this.trackStorage.getRiver(element.id);
        if (!river) {
            continue
        }

        if (river.id == this.toId) {
            return {
                river: river,
                path: element.path,
            }
        }

        let path = element.path.concat([element.id]);
        let refs = river.refs;
        let keys = Object.keys(refs);

        for (let j = 0; j < keys.length; j++) {
            let nId = keys[j];
            if (path.includes(nId)) {
                continue;
            }
            result.push({
                id: nId,
                path: path,
            })
        }
    }
    this.elements = result;

    return this.searchRiverInternal(lvl + 1);
};