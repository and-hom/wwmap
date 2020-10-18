export function selectModelToValue(selected, multiselect, bindId) {
    if (multiselect) {
        if (bindId) {
            return selected.map(s => s.id)
        } else {
            return selected
        }
    } else {
        if (bindId) {
            return selected.length > 0 && selected[0] ? selected[0].id : null
        } else {
            return selected.length > 0 ? selected[0] : null
        }
    }
}

export function valueToSelectModel(value, entities, multiselect, bindId) {
    if (multiselect) {
        if (bindId) {
            return value ? entities.filter(e => value.includes(e.id)) : []
        } else {
            return value ? value : []
        }
    } else {
        if (bindId) {
            return entities.filter(e => e.id == value)
        } else {
            return value ? [value] : [];
        }
    }
}