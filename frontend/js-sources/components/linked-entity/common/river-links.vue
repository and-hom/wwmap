<template>
  <ul v-if="rivers" style="width: 100%; list-style-type: none; padding: 0;">
    <li v-for="river in rivers">
      <div style="width: 100%; min-height: 30px; display: block;">
        <span>{{ river.title }}</span>
        <a :href="editorLink(river)" target="_blank"
           style="float:left;"><img
            src="https://wwmap.ru/img/edit.png" width="25px" :alt="editorLinkAlt()"
            :title="editorLinkAlt()"/></a>
        <a v-if="river.bounds" :href="mapLine(river)" target="_blank"
           style="padding-right:10px;float:left;"><img
            src="https://wwmap.ru/img/locate.png" width="25px" alt="Показать на карте"
            title="Показать на карте"/></a>
        <img v-else
             style="padding-right:10px; float:left; width: 35px; height: 25px; opacity: 0.3;"
             src="https://wwmap.ru/img/locate.png"
             alt="Нет порогов - нечего показывать на карте"
             width="25px"/>
      </div>
    </li>
  </ul>
</template>

<script>
import {frontendBase} from "../../../config";
import {calculateCenter, calculateZoom, DEFAULT_POINT_ZOOM} from "wwmap-js-commons/util";
import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from "../../../auth";

module.exports = {
  props: ["rivers"],
  created() {
    hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
  },
  methods: {
    editorLinkAlt() {
      return this.canEdit ? 'Редактор' : 'Посмотреть в каталоге'
    },
    editorLink(river) {
      return `${frontendBase}editor.htm#${river.country_id},${river.region_id},${river.id}`
    },
    mapLine(river) {
      let center = this.centerOf(river.bounds);
      let z = this.zoomOf(river.bounds);
      return `${frontendBase}map.htm#${center[0]},${center[1]},${z}`
    },
    centerOf(point) {
      if (Array.isArray(point[0])) {
        return calculateCenter(point)
      } else {
        return point;
      }
    },
    zoomOf(p) {
      if (Array.isArray(p[0])) {
        return calculateZoom(p);
      } else {
        return DEFAULT_POINT_ZOOM;
      }
    },
  },
  data() {
    return {
      canEdit: false,
    }
  }
}
</script>