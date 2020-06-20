<template>
    <div :id="uniqueId" class="btn-toolbar justify-content-between">
        <div class="btn-group mr-2" role="group">
            <slot></slot>
        </div>
        <div v-if="showLog()" class="btn-group mr-2">
            <log-dropdown :object-type="logObjectType" :object-id="logObjectId"/>
        </div>
    </div>
</template>

<script type="text/javascript">
    var $ = require("jquery");
    require("jquery.cookie");
    const uuidv4 = require('uuid/v4');
    const DISABLED_CLASS_NAME = 'disabled';

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
            return {
                uniqueId: uuidv4(),
            }
        },
        methods: {
            showLog() {
                return this.logObjectType != null && this.logObjectId != null
            },
            disable() {
                let id = this.uniqueId;
                $(`#${id} button`).addClass(DISABLED_CLASS_NAME);
            },
            enable() {
                let id = this.uniqueId;
                $(`#${id} button`).removeClass(DISABLED_CLASS_NAME);
            },
        }
    }
</script>