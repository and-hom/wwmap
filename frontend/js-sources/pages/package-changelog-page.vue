<template>
    <page link="package-changelog.htm">
        <div class="container-fluid">
            <div class="row">
                <div class="col-12">
                    <h1>{{module}}</h1>
                    <div style="white-space:pre-wrap;">{{content}}</div>
                </div>
            </div>
        </div>
    </page>
</template>

<script>
    import {changelogPathTemplate} from '../config'
    import {sendRequest} from '../api'
    import {format} from 'wwmap-js-commons/util'

    export default {
        created: function () {
            let urlParams = new URLSearchParams(window.location.search);
            this.module = urlParams.get('module');
            let url = format(changelogPathTemplate, this.module);
            sendRequest(url, 'GET', false)
                .then(resp => this.content = resp);
        },
        data() {
            return {
                module: 'None',
                content: 'Loading...',
            };
        }
    }
</script>
