
export function sendRequest(url, _type) {
    return new Promise((resolve, reject) => {
        var xhr = new XMLHttpRequest();
        xhr.open(_type, url, true);


        xhr.onload = () => onLoad(xhr, resolve, reject);
        xhr.onerror = () => reject(xhr.statusText);

        try {
            xhr.send();
        } catch (err) {
            console.log(err);
            reject(err);
        }
    });
}

function onLoad(xhr, resolve, reject) {
    if (xhr.status / 100 != 2) {
        reject(xhr.responseText);
        return;
    }
    resolve(xhr.responseText);
}

export function doGet(url) {
    return sendRequest(url, "GET");
}
