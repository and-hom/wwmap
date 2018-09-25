## Встраивание карты в сайт (простой вариант)

Для того, чтобы встроить карту, нужно иметь доступ к редактированию html и возможность подключать свои javascript и css.

_Например, в wordpress есть специальный плагин, чтобы подключить свой css к отдельной странице. Остальное возможно в рамках редактирования html страницы_

1. Подключаем css отсюда: https://wwmap.ru/css/map.v2.css
2. Подключаем javascript
```
    <script type="text/javascript" src="https://wwmap.ru/js/jquery-3.1.1.min.js"></script>
    <script type="text/javascript" src="https://wwmap.ru/js/jquery.tmpl.js"></script>
    <script type="text/javascript" src="https://wwmap.ru/js/jquery.cookie.js"></script>
    <script type="text/javascript" src="https://api-maps.yandex.ru/2.1/?lang=ru_RU"></script>

    <script type="text/javascript" src="https://wwmap.ru/js/config.js"></script>
    <script type="text/javascript" src="https://wwmap.ru/js/map.v2.js"></script>
```
3. Добавляем div для карты. id произвольный. Возможно, задать размеры, поля и др. параметры области стилями:
```
<div style="width:100%;height 600px;" id="wwmap-container"></div>
```
4. Если нам нужна таблица с реками, ссылками на отчёты и GPX-файлы с точками порогов, то добавляем div и для неё в удобном месте страницы (я расположу его сразу под картой):
```
<div id="wwmap-rivers" class="wwmap-river-menu"></div>
```
5. Добавить javascript для инициализации и работы карты. В случае, если шаг 4 пропущен, второй параметр ф-ции initWWMap также можно опустить.
```
<script type="text/javascript">
  $(document).ready(function() {
    initWWMap("wwmap-container", "wwmap-rivers");
  });
</script>
```

## Кастомизация
### Изменение шаблона списка рек
1. Открываем https://wwmap.ru/map-components/map-html-components.htm. Ищем там элемент с ``id="rivers_template"`` - это стандартный шаблон списка рек.
Копируем его себе в html (id можно не менять, этот файл не загружается напрямую). Подробнее про шаблонизатор можно почитать тут https://github.com/KanbanSolutions/jquery-tmpl
2. Открываем https://wwmap.ru/js/map.v2.js. Ищем ф-цию ``function initWWMap(mapId, riversListId)``. Копируем её себе так, чтобы она перекрыла оригинал,
ну или хотя бы так, чтобы именно она вызывалась на 5-м шаге из предыдущего абзаца.
3. В скопированной на предыдущем шаге ф-ции ищем инициализацию списка рек: ``riverList = new RiverList($('#' + riversListId), 'rivers_template', true)``.
Меняем ``true`` на ``false``. ``true`` - загружать шаблон из файла https://wwmap.ru/map-components/map-html-components.htm, ``false`` - брать с текущей страницы. Также
при необходимости можно поправить id шаблона (второй параметр)
4. Теперь можно редактировать шаблон по-своему. Язык шаблонов - jquery-template. Поля отчёта:
    1. **title"** - заголовок отчёта (сильно зависит правил оформления от сайта, на котором он размещён)
    1. **author"** - автор отчёта (чаще всего руководитель, но нельзя гарантировать)
    1. **year"** - год, в который совершён поход (не год публикации)
    1. **url"** - ссылка на отчёт
    1. **source_logo_url"** - логотип сайта, на котором размещён отчёт

