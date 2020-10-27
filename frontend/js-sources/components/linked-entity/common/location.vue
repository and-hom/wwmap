<template>
  <a :href="href" target="_blank"><img class="location-medium" :src="icon" alt="Показать на карте" title="Показать на карте"/></a>
</template>

<style>
.location-medium {
  width: 48px;
  height: 48px;
}
</style>

<script>
import {frontendBase} from "../../../config";
import {calculateCenter, calculateZoom, DEFAULT_POINT_ZOOM} from "wwmap-js-commons/util";

module.exports = {
  props: {
    pointOrLine: {
      required: true,
    }
  },
  computed: {
    href() {
      let line = Array.isArray(this.pointOrLine[0]);
      let z = line ? calculateZoom(this.pointOrLine) : DEFAULT_POINT_ZOOM;
      let center = line ? calculateCenter(this.pointOrLine) : this.pointOrLine;
      return `${frontendBase}map.htm#${center[0]},${center[1]},${z}`;
    }
  },
  data() {
    return {
      icon: `${frontendBase}/img/locate.png`,
    }
  }
}
</script>