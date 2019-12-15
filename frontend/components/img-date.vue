<template>
    <datepicker format="yyyy-MM-dd" v-model="date" @selected="function (d) {
                    setImageDate(spotId, imgId, d);
                    $emit('input', d);
                }" :clear-button="true"/>
</template>

<style type="text/css">

</style>


<script>
    module.exports = {
        props: {
            spotId: {
                type: Number,
                required: true,
            },
            imgId: {
                type: Number,
                required: true,
            },
            value: {
                validator: prop => typeof prop === 'object' || prop === null,
                required: true,
            }
        },
        data: function () {
            return {
                date: this.value,
                setImageDate: setImageDate,
            }
        },
    }
</script>