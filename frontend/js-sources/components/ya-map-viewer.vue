<template>
  <div id="map" :style="style"></div>
</template>

<script>
import {backendApiBase} from "../config";
import {addMapLayers, registerMapSwitchLayersHotkeys} from "../map-common";
import {navigateToSpot} from "../app-state";

module.exports = {
  props: {
    bounds: {
      type: Array,
      required: true,
    },
    style: {
      type: String,
      required: false,
    },
  },
  created() {
    this.objectManager.setUrlTemplate(backendApiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z&river=' + this.river.id);
    this.map.setBounds(this.bounds);
  },
  watch: {
    bounds(newValue, _) {
      this.map.setBounds(this.bounds);
    }
  },
  data() {
    return {
      map: null,
      objectManager: null,
    }
  },
  methods: {
    showMap: function () {
      if (this.bounds == null) {
        return;
      }
      let t = this;
      ymaps.ready(function () {
        ymaps.modules.require(['overlay.BiPlacemark'], function (BiPlacemarkOverlay) {
          if (ymaps.overlay.storage.get("BiPlacemrakOverlay") == null) {
            ymaps.overlay.storage.add("BiPlacemrakOverlay", BiPlacemarkOverlay);
          }
          addMapLayers();

          let mapParams = t.mapParamsStorage.getLastPositionZoomType();
          let map = new ymaps.Map("map", {
            bounds: expandIfTooSmall(t.bounds),
            type: mapParams.type,
            controls: ["zoomControl"]
          });

          t.mapParamsStorage.setLastPositionZoomType(map.getCenter(), map.getZoom(), map.getType())
          map.events.add('typechange', function () {
            t.mapParamsStorage.setLastPositionZoomType(map.getCenter(), map.getZoom(), map.getType())
          });
          map.events.add('boundschange', function () {
            t.mapParamsStorage.setLastPositionZoomType(map.getCenter(), map.getZoom(), map.getType())
          });
          map.controls.add(
              new ymaps.control.TypeSelector([
                    'osm#standard',
                    'ggc#standard',
                    'topomapper#genshtab',
                    'marshruty.ru#genshtab',
                    'yandex#satellite',
                    'google#satellite',
                    'bing#satellite',
                  ]
              )
          );
          registerMapSwitchLayersHotkeys(map);
          var objectManager = new ymaps.RemoteObjectManager(backendApiBase + '/ymaps-tile-ww?bbox=%b&zoom=%z&river=' + t.river.id, {
            clusterHasBalloon: false,
            geoObjectOpenBalloonOnClick: false,
            geoObjectStrokeWidth: 3,
            splitRequests: true
          });

          objectManager.objects.events.add(['click'], function (e) {
            let id = e.get('objectId');
            navigateToSpot(id, false);
          });

          map.geoObjects.add(objectManager);
          t.map = map;
          t.objectManager = objectManager;
        });
      });
    },
    reloadMap: function () {
      if (this.map) {
        this.map.destroy();
      }
      this.showMap();
    },
  },
}
</script>