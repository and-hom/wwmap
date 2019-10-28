<template>
    <div style="display: none;">{{text}}</div>
</template>

<script>
    module.exports = {
        props: {
            text: {
                type: String,
            },
        },
        created: function() {
            document.title = this.text;
        },
        updated: function() {
            document.title = this.text;
        }
    }
</script>

<style>
</style>