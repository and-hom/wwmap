## Встраивание карты в сайт

Для того, чтобы встроить карту, нужно иметь доступ к редактированию html и возможность подключать свои javascript и css.

_Например, в wordpress есть специальный плагин, чтобы подключить свой css к отдельной странице. Остальное возможно в рамках редактирования html страницы_

1. Подключаем css отсюда: https://wwmap.ru/css/index.css
2. Добавляем div для карты (возможно, задать размеры,поля и др. параметры области стилями):
```
<div id="map"></div>
```
3. Добавляем шаблон краткого описания порога (bubble на карте):
```
<div id="bubble_template" style="display:none">
    <h3 class="popover-title">
        {%if properties.link%}<a target="_blank" href="{{properties.link}}">{{properties.title}}</a>{%else%}{{properties.title}}{%endif%}
    </h3>

    <div class="popover-content">
        {%if properties.category%}<div>Категория сложности: {{properties.category}}</div>{%endif%}
        {%if properties.river_name%}<div>Река: {{properties.river_name}}</div>{%endif%}
        <div>{{properties.short_description}}</div>
        <a target="_top" href="" onclick="show_report_popup({{properties.id}}); return false">Сообщить о неточном местоположении</a>
    </div>
</div>
```
4. **Опционально:** если мы хотим видеть список отображаемых в данный момент на карте рек со ссылками для скачивания gpx, нужно добавить элемент,
в котором будет размещён список, (в нашем случае _ul_) и шаблон списка
```
<div id="rivers" class="riverMenu"></div>
<script type="text/x-jquery-tmpl" id="riversMenuTemplate">
    <table>
    {{each rivers}}
        <tr>
            <td class="title-col"><a href="" style="padding-left:10px;" onclick="show_map_at(${bounds}); return false;">${title}</a></td>
            <td class="gpx-col">
                <a href="${apiUrl}/${id}" style="padding-right:10px;" alt="Скачать GPX с точками порогов">GPX</a>
                <a href="${apiUrl}/${id}?tr=true" alt="Скачать GPX с точками порогов">GPX<sub>en</sub></a>
            </td>
            <td class="report-col">
                <ul class="reports">
                {{each reports}}
                    <li><a target="_blank" href="${url}" title="${title}"><img src="${source_logo_url}"/>{{if year>1}}${year}{{/if}}&nbsp;${title}</a></li>
                {{/each}}
                </ul>
            </td>
        </tr>
    {{/each}}
    </table>
</script>
```
5. Добавить диалог для сообщения о неточностях местоположения (если убрать такую ссылку из шаблона bubble п.3, то можно и не добавлять)
```
<div class="popup_area">
    <div class="popuptext" id="report_popup">
        <form>
            <label for="object_id">ID</label>
            <input type="text" id="object_id" name="object_id" readonly style="margin-bottom:10px;"/><br/>
            <label for="comment" style="width:400px">Исправления происходят в ручном режиме и после проверки. Пожалуйста, расскажите коротко,
                что не так. Будет полезным добавить источники информации (отчёты, например) и указать реальные координаты точки.
                <u>Оставьте контакт для обратной связи</u> на случай, если у меня возникнет вопрос.</label><br/>
            <textarea id="comment" name="comment" rows="20" maxlength="4000" style="margin-top:10px; margin-bottom:20px; "></textarea>
            <input type="submit" value="Отправить" width="600px; align: center;"/>
            <input type="button" name="cancel" value="Отмена" width="600px; align: center;"/>
        </form>
    </div>
</div>


<div id="report_ok" style="display:none">Запрос отправлен. Я прочитаю его по мере наличия свободного времени.</div>
<div id="report_fail" style="display:none">Что-то пошло не так...</div>
```

6. Добавить туториал
```
<div class="popup_area">
    <div class="popuptext" style="width:1000px;" id="info_popup">
        <h3>Что это такое?</h3>
        <p>Это карта порогов для водного туризма с описаниями, категориями, ссылками на отчёты и выгрузкой GPX-файлов в навигатор. </p>
        <h3>Как добавить новые препятствия на карту?</h3>
        <p>Сейчас это делается не самым простым способом:
        </p>
            <ol>
                <li>В любом удобном редакторе геоданных нужно добавить точки в те места, где находятся пороги. Кроме того
                    можно отметить совершенно непроходимые опасные участки, например, многолетние завалы. Точка порога должна
                    находиться там, где непосредственно начинается препятствие, либо на месте заходной шиверы, если её сила
                    существенна. В качестве редактора геоданных можно использовать, например, <a target="_blank" href="http://sasgis.org">SAS Planet</a>.</li>
                <li>У каждой точки должно быть краткое описание: не более 1000 символов. В начале описания нужно упомянуть категорию препятствия
                и по возможности ссылку на отчёт в интернете. При добавлении я внесу их в отдельные поля, удалив из описания.</li>
                <li>Затем из редактора нужно выгрузить полученные точки в формате
                    <a target="_blank" href="https://ru.wikipedia.org/wiki/GPX">GPX</a> так, чтобы в каждом файле были препятствия одной реки.
                    В случае крайней необходимости можно выгрузить их и в формате KML, но это будет не так удобно.</li>
                <li>Полученные файлы отправить по электронной почте на <a class="email-link" href=""></a>. Иногда я бываю в походе, поэтому
                не обещаю обрабатывать их быстро. Иногда это может происходить в течении месяца. Возможно у меня будут уточнения и вопросы,
                так как я проверяю добавляемые точки.</li>
            </ol>
        <h3>Как ещё помочь развитию карты?</h3>
        <ol>
            <li>Рассказать про открытые библиотеки отчётов по водному туризму и способы автомтизированного анализа их отчётов</li>
            <li>Разместить карту у себя на сайте. Как это сделать, описано <a target="_blank" href="https://github.com/and-hom/wwmap/blob/master/INTEGRATION_ru.md">тут</a></li>
            <li>Помочь в разработке</li>
        </ol>
        <h3>Как связаться с разработчиком?</h3>
        <p>По всем свящанным с картой вопросам нужно писать по электронной почте <a class="email-link" href=""></a>
        Также можно ознакомиться с проектом на <a target="_blank" href="https://github.com/and-hom/wwmap">github</a>, узнать о текущи задачах,
            предложить улучшения и предложить пулл-реквест.</p>
        <input type="button" name="ok" value="Понятно" width="600px; align: center;"/>
    </div>
</div>
```

7. Подключить кучу скриптов (Вполне вероятно, что среди них могут быть и неиспользуемые.
Если вы найдёте такой, сообщите, и я его уберу):
```
<script type="text/javascript" src="https://wwmap.ru/js/jquery-3.1.1.min.js"></script>
<script type="text/javascript" src="https://wwmap.ru/js/jquery.tmpl.js"></script>
<script type="text/javascript" src="https://wwmap.ru/js/tether.min.js"></script>
<script type="text/javascript" src="https://wwmap.ru/js/jquery.cookie.js"></script>
<script type="text/javascript" src="https://wwmap.ru/js/jquery.json.js"></script>
<script type="text/javascript" src="https://wwmap.ru/js/index.js"></script>
<script type="text/javascript" src="https://api-maps.yandex.ru/2.1/?lang=ru_RU"></script>
<script type="text/javascript" src="https://wwmap.ru/js/config.js"></script>
<script type="text/javascript" src="https://wwmap.ru/js/map.js"></script>
```


8. Добавить несколько javascript-ф-ций
```
<script type="text/javascript">
  function show_map_at(bounds) {
    myMap.setBounds(bounds, {
            checkZoomRange: true,
            duration: 200,
     })
  }

  function show_report_popup(id){
      $("#report_popup #object_id").val(id)
      $("#report_popup").addClass("show");
  }

  function show_info_popup(){
      $("#info_popup").addClass("show");
  }

  $(document).ready(function() {
    $("#info_popup input[name=ok]").click(function(){
        $("#info_popup").removeClass("show")
    });
    $("#report_popup input[name=cancel]").click(function(){
        $("#report_popup").removeClass("show")
    });
    $("#report_popup input[type=submit]").click(function(){
        $.post(apiBase + "/report", $( "#report_popup form" ).serialize() )
        .done(function( data ) {
            window.alert($("#report_ok").html());
            $("#report_popup").removeClass("show")
            $('#comment').val('')
        })  .fail(function() {
            window.alert($("#report_fail").html());
        });
        return false;
    });

    user='info'
    domain='wwmap.ru'
    $('.email-link').attr('href','mailto:'+user+'@'+domain)
    $('.email-link').text(user+'@'+domain)
  });
</script>
```