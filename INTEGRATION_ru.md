## Встраивание карты в сайт (простой вариант)

Для того, чтобы встроить карту, нужно иметь доступ к редактированию html и возможность подключать свои javascript и css.

_Например, в wordpress есть специальный плагин, чтобы подключить свой css к отдельной странице. Остальное возможно в рамках редактирования html страницы_

1. Подключаем javascript

    ```html
        <script type="text/javascript" src="https://api-maps.yandex.ru/2.1/?lang=ru_RU"></script>
        <script type="text/javascript" src="https://wwmap.ru/js/map.v2.1.js"></script>
    ```
 
        Потребуется только для вызова document.ready в п.4

    ```html
        <script type="text/javascript" src="https://wwmap.ru/js/jquery-3.1.1.min.js"></script>
    ```
2. Добавляем div для карты. id произвольный. Возможно, задать размеры, поля и др. параметры области стилями:

    ```html
    <div style="width:100%;height 600px;" id="wwmap-container"></div>
    ```
3. Если нам нужна таблица с реками, ссылками на отчёты и GPX-файлы с точками порогов, то добавляем div и для неё в удобном месте страницы (я расположу его сразу под картой):

    ```html
    <div id="wwmap-rivers" class="wwmap-river-menu"></div>
    ```
4. Добавить javascript для инициализации и работы карты.

    ```html
    <script type="text/javascript">
      $(document).ready(function() {
        wwmap.initWWMap("wwmap-container", "wwmap-rivers");
      });
    </script>
    ```
    В случае, если шаг 3 пропущен, второй параметр ф-ции initWWMap также можно опустить:
    ```html
    <script type="text/javascript">
      $(document).ready(function() {
        wwmap.initWWMap("wwmap-container");
      });
    </script>
    ```
5. При желании поменять ссылку, ведущую из краткого описания порога на карте.
Возможные варианты:
    * **none** - никогда не показывать ссылки в кратком описании порога на карте
    * по-умолчанию (когда ничего не указано) - показывать ссылку, указанную автором описания
    * **wwmap** - показывать ссылку в каталог на https://wwmap.ru, где описание можно редактировать
    * **huskytm** - показывать ссылку в каталог, выгруженный на huskytm.ru (только просмотр)

    Как сделать: поменять код из предыдущего пункта на такой (последним параметром указать значение из списка выше):
    ```html
    <script type="text/javascript">
    $(document).ready(function() {
        wwmap.initWWMap("wwmap-container", "wwmap-rivers", "wwmap");
    });
    </script>
    ```

## Кастомизация
### Изменение шаблона списка рек
1. Открываем https://wwmap.ru/map-components/map-html-components-2.1.htm. Ищем там элемент с ``id="rivers_template"`` - это стандартный шаблон списка рек.
Копируем его себе в html (id можно не менять, этот файл не загружается напрямую). Подробнее про шаблонизатор можно почитать тут https://github.com/KanbanSolutions/jquery-tmpl . Выглядит он, например, так:

    ```html
    <script type="text/x-jquery-tmpl" id="rivers_template">
    <div class="wwmap-river-menu">
        {%each rivers%}
            <div class="wwmap-river-menu-item"><div class="wwmap-river-menu-title">
                <a href="" style="padding-left:10px;" onclick="wwmap.show_river_info_popup(${id}); return false;">${title}</a>
            </div><div class="wwmap-river-menu-controls">
                <a href="" style="padding-left:10px;" onclick="wwmap.show_map_at(${bounds}); return false;"><img src="https://wwmap.ru/img/locate.png" width="25px" alt="Показать на карте" title="Показать на карте"/></a>
                {%if canEdit%}
                    <a href="https://wwmap.ru/editor.htm#${region.country_id},${region.id},${id}" target="_blank" style="margin-left:-6px;"><img src="https://wwmap.ru/img/edit.png" width="25px" alt="Редактор" title="Редактор"/></a>
                {%/if%}
            </div></div>
        {%/each%}
    </div>
    </script>

    ```
2. Вместо wwmap.initWWap используем ф-цию `initWWMapCustomRiverList(mapId, riversListTemplateElement, catalogLinkType)`, например, так

    ```html
    <script type="text/javascript">
    $(document).ready(function() {
        let templateContainer = $("#wwmap-rivers");
        wwmap.initWWMap("wwmap-container", templateContainer, "wwmap");
    });
    </script>
    ```
3. В скопированной на предыдущем шаге ф-ции ищем инициализацию списка рек: ``riverList = new RiverList($('#' + riversListId), 'rivers_template', true)``.
Меняем ``true`` на ``false``. ``true`` - загружать шаблон из файла https://wwmap.ru/map-components/map-html-components.htm, ``false`` - брать с текущей страницы. Также
при необходимости можно поправить id шаблона (второй параметр)
4. Теперь можно редактировать шаблон по-своему. Язык шаблонов - jquery-template. Поля отчёта:
    1. **title"** - заголовок отчёта (сильно зависит правил оформления от сайта, на котором он размещён)
    1. **author"** - автор отчёта (чаще всего руководитель, но нельзя гарантировать)
    1. **year"** - год, в который совершён поход (не год публикации)
    1. **url"** - ссылка на отчёт
    1. **source_logo_url"** - логотип сайта, на котором размещён отчёт
    
    
## Примечание
При использовании vue.js div, в котором должна была отобразиться карта, имел сначала нулевой размер, и карта не отображалась. Пришлось вспользоваться вот таким приёмом:

```javascript
    $(document).ready(function() {

     Vue.component('auth', httpVueLoader('components/auth.vue'));
     Vue.component('page', httpVueLoader('components/page.vue'));

     var app = new Vue({
        el: '#vue-app',
        data: {
        },
        mounted: loadMapWhenDivIsReady
    });
  });

  function loadMapWhenDivIsReady() {
      // Если div отображён, инициализируем карту
      if($('#wwmap-container').outerWidth()) {
          wwmap.initWWMap("wwmap-container", "wwmap-rivers", "wwmap")
      } else {
          // Через 100мс пытаемся ещё раз сделать то же самое
          console.log("#div-container is not ready yet: has no offsetWidth");
          setTimeout(loadMapWhenDivIsReady, 100);
      }
  }
```

