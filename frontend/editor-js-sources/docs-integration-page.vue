<template>
    <page link="docs-integration.htm">
        <div class="container-fluid">
            <div class="row">
                <div class="col-12" style="padding-top:30px;">
                    <div id="integration-guide" v-html="integrationGuide">
                    </div>
                </div>
            </div>
        </div>
    </page>
</template>

<script>
    var showdown = require('showdown');
    import {sendRequest} from './api'

    export default {
        created: function () {
            var converter = new showdown.Converter();

            sendRequest('INTEGRATION_ru.md', 'GET', false)
                .then(resp => this.integrationGuide = converter.makeHtml(resp))
                .catch(_ => sendRequest('../INTEGRATION_ru.md', 'GET', false))
                .then(resp => this.integrationGuide = converter.makeHtml(resp))
                .catch(_ => this.integrationGuide = '<span style="color:red">Can not load</span>');
        },
        data() {
            return {
                integrationGuide: '<span style="color:gray">Loading...</span>'
            }
        },
    }


</script>
