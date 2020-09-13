<template>
  <tr :class="rowClass(transfer)">
    <td>{{ transfer.title }}</td>
    <td>
      <ul class="ti-tags">
        <li v-for="station in transfer.stations" class="ti-tag">
          <div class="ti-content">
            <div class="ti-tag-center"><span class="">{{ station }}</span></div>
          </div>
        </li>
      </ul>
    </td>
    <td>
      <ul style="list-style-type: none; padding: 0;">
        <li v-for="river in transfer.rivers">
          <div style="width: 100%; min-height: 30px; display: block;">
            <span>{{ river.title }}</span>
            <a :href="editorLink(river)" target="_blank"
               style="float:right;"><img
                src="https://wwmap.ru/img/edit.png" width="25px" :alt="editorLinkAlt"
                :title="editorLinkAlt"/></a>
            <a v-if="river.bounds" :href="mapLine(river)" target="_blank"
               style="padding-left:10px;float:right;"><img
                src="https://wwmap.ru/img/locate.png" width="25px" alt="Показать на карте"
                title="Показать на карте"/></a>
          </div>
        </li>
      </ul>
    </td>
    <td class="raw-text-content">{{ transfer.description }}</td>
    <slot></slot>
  </tr>
</template>

<style>
.row-selected {
  background: #eeeedd;
}
</style>

<script>
import {frontendBase} from "../../config";
import {calculateZoom} from "wwmap-js-commons/util";

module.exports = {
  props: {
    transfer: {
      type: Object,
      required: true,
    },
    selected: {
      type: Boolean,
      default: false,
    },
    canEdit: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      editorLinkAlt: this.canEdit ? 'Редактор' : 'Посмотреть в каталоге',
    }
  },
  methods: {
    rowClass: function (transfer) {
      return this.selected ? 'row-selected' : '';
    },
    editorLink: function (river) {
      return `${frontendBase}editor.htm#${river.country_id},${river.region_id},${river.id}`
    },
    mapLine: function (river) {
      let bounds = this.centerOf(river.bounds);
      let z = this.zoomOf(river.bounds);
      return `${frontendBase}map.htm#${bounds[0]},${bounds[1]},${z}`
    },
    centerOf: function (point) {
      if (Array.isArray(point[0])) {
        let p = [point[0], point[point.length - 1]];
        return [(p[0][0] + p[1][0]) / 2, (p[0][1] + p[1][1]) / 2,]
      } else {
        return point;
      }
    },
    zoomOf: function (p) {
      if (Array.isArray(p[0])) {
        return calculateZoom(p);
      } else {
        return 15;
      }
    },
  }
}
</script>