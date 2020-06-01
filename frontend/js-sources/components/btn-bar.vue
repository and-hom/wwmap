<template>
    <div class="btn-toolbar justify-content-between">
        <div class="btn-group mr-2" role="group">
            <slot></slot>
        </div>
        <div v-if="showLog()" class="btn-group mr-2">
            <log-dropdown :object-type="logObjectType" :object-id="logObjectId"/>
        </div>
    </div>
</template>

<script type="text/javascript">
    module.exports = {
        props: {
            logObjectType: {
                type: String,
                required: false,
            },
            logObjectId: {
                type: Number,
                required: false,
            },
        },
        data: function () {
            return {}
        },
        methods: {
            showLog() {
                return this.logObjectType != null && this.logObjectId != null
            }
        }
    }
</script>