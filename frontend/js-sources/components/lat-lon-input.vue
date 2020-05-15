<template>
    <div class="coord-input-line">
        <label :for="'lat_'+uniqueId">Широта</label> <input :id="'lat_'+uniqueId" v-bind:value="lat"/>
        <label :for="'lon_'+uniqueId">Долгота</label> <input :id="'lon_'+uniqueId" v-bind:value="lon"/>
    </div>
</template>

<style type="text/css">
    .coord-input-line {
        display: flex;
        flex-direction: row;
        flex-wrap: nowrap;
    }

    .coord-input-line label {
        display: block;
        float: left;
        white-space: nowrap;
        margin-right: 6px;
    }

    .coord-input-line input {
        display: inline-block;
        float: left;
        white-space: nowrap;
        margin-right: 12px;
        width: 100%;
    }
</style>


<script>
    const uuidv4 = require('uuid/v4');
    import Inputmask from "inputmask";

    import {DegMinCoordMarshaller, DegMinSecCoordMarshaller, NumberCoordMarshaller, norm} from '../coord-marshaller'

    let DEG_MIN_SEC = 'deg-min-sec';
    let DEG_MIN = 'deg-min';
    let RAW = 'raw';

    module.exports = {
        props: {
            value: {
                type: Array,
                required: true,
            },
            mode: {
                type: String,
                default: DEG_MIN_SEC,
            }
        },
        created: function () {
            this.latMarshaller = this.getLatMarshaller(this.mode);
            this.lonMarshaller = this.getLonMarshaller(this.mode);
        },
        mounted: function () {
            this.onModeChanged(this.mode)
        },
        watch: {
            mode: function (newVal) {
                this.onModeChanged(newVal);
            },
        },
        computed: {
            lat: {
                get: function () {
                    return this.latMarshaller.marshal(this.point[0]);
                },
            },
            lon: {
                get: function () {
                    return this.lonMarshaller.marshal(this.point[1]);
                },
            },
        },
        data: function () {
            return {
                point: [...this.value],
                uniqueId: uuidv4(),
                showMap: this.showMapByDefault,

                onModeChanged: function (newVal) {
                    this.latMarshaller = this.getLatMarshaller(newVal);
                    this.lonMarshaller = this.getLonMarshaller(newVal);

                    let latInput = $('#lat_' + this.uniqueId);
                    let lonInput = $('#lon_' + this.uniqueId);
                    let t = this;

                    latInput.change(function (evt) {
                        t.point[0] = norm(t.latMarshaller.unmarshal(latInput.val()), -75, 75);
                        t.$emit('input', t.point)
                    });
                    lonInput.change(function (evt) {
                        t.point[1] = norm(t.lonMarshaller.unmarshal(lonInput.val()), -180, 180);
                        t.$emit('input', t.point)
                    });

                    Inputmask.remove(latInput);
                    Inputmask.remove(lonInput);

                    t.setInputmask(latInput, newVal, 2, "[NSns]");
                    t.setInputmask(lonInput, newVal, 3, "[EWew]");
                },
                setInputmask: function (input, newVal, maxDeg, regex) {
                    if (newVal == DEG_MIN_SEC) {
                        Inputmask({
                            mask: "A9{" + maxDeg + "}° 99' 99.9{3}''",
                            autoUnmask: true,
                            greedy: false,
                            definitions: {
                                'A': {
                                    validator: regex,
                                    casing: "upper"
                                }
                            }
                        }).mask(input);
                    } else if (newVal == DEG_MIN) {
                        Inputmask({
                            mask: "A9{" + maxDeg + "}° 99.9{5}'",
                            autoUnmask: true,
                            greedy: false,
                            definitions: {
                                'A': {
                                    validator: regex,
                                    casing: "upper"
                                }
                            }
                        }).mask(input);
                    } else if (newVal == RAW) {
                        Inputmask({
                            mask: "[A]9{" + maxDeg + "}.9{12}°",
                            autoUnmask: true,
                            greedy: false,
                            definitions: {
                                'A': {
                                    validator: "-",
                                    casing: "upper"
                                }
                            }
                        }).mask(input);
                    }
                },

                numberMarshallerLat: new NumberCoordMarshaller(2, 12),
                numberMarshallerLon: new NumberCoordMarshaller(3, 12),
                degMinMarshallerLat: new DegMinCoordMarshaller(2, "N", "S"),
                degMinMarshallerLon: new DegMinCoordMarshaller(3, "E", "W"),
                degMinSecMarshallerLat: new DegMinSecCoordMarshaller(2, "N", "S"),
                degMinSecMarshallerLon: new DegMinSecCoordMarshaller(3, "E", "W"),

                getLatMarshaller: function (mode) {
                    switch (mode) {
                        case RAW:
                            return this.numberMarshallerLat;
                        case DEG_MIN:
                            return this.degMinMarshallerLat;
                        case DEG_MIN_SEC:
                            return this.degMinSecMarshallerLat;
                    }
                },
                getLonMarshaller: function (mode) {
                    switch (mode) {
                        case RAW:
                            return this.numberMarshallerLon;
                        case DEG_MIN:
                            return this.degMinMarshallerLon;
                        case DEG_MIN_SEC:
                            return this.degMinSecMarshallerLon;
                    }
                },

                latMarshaller: null,
                lonMarshaller: null,

            }
        }
    }

</script>