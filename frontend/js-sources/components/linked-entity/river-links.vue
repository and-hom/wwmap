<template>
  <ul v-if="rivers" style="width: 200px; list-style-type: none; padding: 0;">
    <li v-for="river in rivers">
      <div style="width: 100%; min-height: 30px; display: block;">
        <span>{{ river.title }}</span>
        <a :href="editorLink(river)" target="_blank"
           style="float:right;"><img
            src="https://wwmap.ru/img/edit.png" width="25px" :alt="editorLinkAlt()"
            :title="editorLinkAlt()"/></a>
        <a v-if="river.bounds" :href="mapLine(river)" target="_blank"
           style="padding-left:10px;float:right;"><img
            src="https://wwmap.ru/img/locate.png" width="25px" alt="Показать на карте"
            title="Показать на карте"/></a>
      </div>
    </li>
  </ul>
</template>

<script>
import {frontendBase} from "../../config";
import {calculateZoom} from "wwmap-js-commons/util";
import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from "../../auth";

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
      let bounds = this.centerOf(river.bounds);
      let z = this.zoomOf(river.bounds);
      return `${frontendBase}map.htm#${bounds[0]},${bounds[1]},${z}`
    },
    centerOf(point) {
      if (Array.isArray(point[0])) {
        let p = [point[0], point[point.length - 1]];
        return [(p[0][0] + p[1][0]) / 2, (p[0][1] + p[1][1]) / 2,]
      } else {
        return point;
      }
    },
    zoomOf(p) {
      if (Array.isArray(p[0])) {
        return calculateZoom(p);
      } else {
        return 15;
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