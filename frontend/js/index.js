function throttle(func, interval) {
    var lastCall = 0;
    return function (args) {
        var now = Date.now();
        if (lastCall + interval < now) {
            lastCall = now;
            return func(args);
        }
    };
}

function addMouseOverOutHighliterListeners(itemSelector, geodataSelector, selectedClass, geoObjectCreator, highlighter) {
    $(document).on('mouseover', itemSelector, function (obj) {
        var targetLi = $(obj.currentTarget);
        targetLi.addClass(selectedClass);
        var geodataDiv = targetLi.find(geodataSelector);
        if (geodataDiv.length) {
            var geodataStr = geodataDiv.html();
            highlighter.val = geoObjectCreator($.parseJSON(geodataStr));
            myMap.geoObjects.add(highlighter.val);
        }
    });
    $(document).on('mouseout', itemSelector, function (obj) {
        $(obj.currentTarget).removeClass(selectedClass);
        if (highlighter.val) {
            myMap.geoObjects.remove(highlighter.val);
            highlighter.val = null;
        }
    });
}
