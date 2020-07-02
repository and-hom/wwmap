const WWMAP_SESSION_ID = 'wwmap_token';

export function getWwmapSessionId() {
    return window.localStorage[WWMAP_SESSION_ID]
}

export function clearWwmapSessionId() {
    window.localStorage.removeItem(WWMAP_SESSION_ID);
}

export function setWwmapSessionId(sessionId) {
    window.localStorage[WWMAP_SESSION_ID] = sessionId;
}