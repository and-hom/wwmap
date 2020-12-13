<template>
  <iframe width="100%" height="100%"
          :src="embeddedVideoUrl(image.remote_id, image.props.t)"
          frameborder="0"
          allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
          allowfullscreen></iframe>
</template>


<script>

module.exports = {
  props: {
    image: {
      type: Object,
      required: true,
    }
  },
  data() {
    return {}
  },
  methods: {
    embeddedVideoUrl: function (id, t) {
      let url = "https://www.youtube.com/embed/" + id + "?fs=1";
      if (t) {
        url += `&start=${t}`
      }
      return url
    }
  }
}
</script>