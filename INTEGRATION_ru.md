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
5. При желании задать третий параметр options - объект с набором опций. Несколько опций можно задать одновременно.
    1. **catalogLinkType** - cсылка, ведущая из краткого описания порога на карте.
        Возможные варианты:
        * **none** - никогда не показывать ссылки в кратком описании порога на карте
        * по-умолчанию (когда ничего не указано) - показывать ссылку, указанную автором описания
        * **wwmap** - показывать ссылку в каталог на https://wwmap.ru, где описание можно редактировать
        * **huskytm** - показывать ссылку в каталог, выгруженный на huskytm.ru (только просмотр)

        Как сделать: поменять код из предыдущего пункта на такой (последним параметром указать значение из списка выше):
        
        ```html
        <script type="text/javascript">
           $(document).ready(function() {
               wwmap.initWWMap("wwmap-container", "wwmap-rivers", {
                   catalogLinkType: "wwmap"
               });
           });
        </script>
        ```
       
    2. **riversTemplateData** - свой шаблон списка рек.  Язык шаблонов - https://idangero.us/template7/. Пример:
        
        ```html
        <div id="my-river-template" style="display: none">
           {{#each rivers}}
               <div class="wwmap-river-menu-item"><div class="wwmap-river-menu-title">
                   <a href="" style="padding-left:10px;" onclick="wwmap.show_river_info_popup({{id}}); return false;">{{title}}</a>
               </div><div class="wwmap-river-menu-controls">
                   <a href="" style="padding-left:10px;" onclick="wwmap.show_map_at_and_highlight_river({{bounds}}, {{id}}); return false;"><img src="https://wwmap.ru/img/locate.png" width="25px" alt="Показать на карте" title="Показать на карте"/></a>
               </div></div>
           {{/each}}
        </div> 
        <script type="text/javascript">
           $(document).ready(function() {
               wwmap.initWWMap("wwmap-container", "wwmap-rivers", {
                   riversTemplateData: $('#my-river-template').html()
               });
           });
        </script>
        ```
   
    3. **userInfoFunction** - собственная ф-ция определения логина. Когда пользователь жалуется на неточность по-умолчанию 
    сообщение будет анонимным. С помощью этой опции можно задать логин, чтобы с пользователем можно было связаться 
    для обсуждения неточности
   
        ```html
        <script type="text/javascript">
           function myUserInfoFunction() {
              return new Promise((resolve, reject) => {
                      resolve({
                              login: "ivan_petrov",
                              auth_provier: "your_site.com",
                          });
                  })
           }   
          
           $(document).ready(function() {
               wwmap.initWWMap("wwmap-container", "wwmap-rivers", {
                   userInfoFunction: myUserInfoFunction
               });
           });
        </script>
        ```

    
## Примечание
При использовании vue.js div, в котором должна была отобразиться карта, имел сначала нулевой размер, и карта не отображалась. Пришлось вспользоваться вот таким приёмом:

```javascript
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

