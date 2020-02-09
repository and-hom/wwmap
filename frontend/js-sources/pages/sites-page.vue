<template>
    <page link="sites.htm">

        <div style="margin-left:10px; margin-top: 10px;">
            <h2>Сайты, на которых размещена карта</h2>
            <ul>
                <li v-for="ref in refs"><a target="_blank" :href="ref.page">{{ ref.page }}</a></li>
            </ul>
        </div>
    </page>
</template>

<script>
    import {backendApiBase} from '../config'
    import {doGetJson} from '../api'


    export default {
        data() {
            return {
                refs: [],
            }
        },
        created() {
            doGetJson(backendApiBase + "/dashboard/ref-sites").then(refs => this.refs = refs)
        }
    }
</script>
