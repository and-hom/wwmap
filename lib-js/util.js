export const DEFAULT_POINT_ZOOM = 15;

export function calculateCenter(point) {
    let p = [point[0], point[point.length - 1]];
    return [(p[0][0] + p[1][0]) / 2, (p[0][1] + p[1][1]) / 2,]
}

export function calculateZoom(p) {
    let dx = Math.abs(p[0][0] - p[1][0]);
    let dy = Math.abs(p[0][1] - p[1][1]);
    if (dx < 0.001 && dy < 0.001) {
        return 15;
    }
    let d = Math.max(dx, dy);
    let z = Math.log(180 / d) / Math.log(2) + 2;
    return Math.min(Math.round(z), 19);
}

export function format() {
    var s = arguments[0];
    for (var i = 0; i < arguments.length - 1; i++) {
        var reg = new RegExp("\\{" + i + "\\}", "gm");
        s = s.replace(reg, arguments[i + 1]);
    }
    return s;
}

export function arrays_intersects(a1, a2) {
    return a1.filter(el => a2.indexOf(el) >= 0).length > 0
}

export function randomStringId(len) {
    let result = Math.random().toString(36).substring(2, 15);
    while (result.length < len) {
        result += Math.random().toString(36).substring(2, 15);
    }
    return result.substring(0, len);
}
