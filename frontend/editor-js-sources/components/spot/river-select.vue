<template>
    <v-select v-model="river"
              label="title"
              :filterable="false"
              :options="options"
              @search="onSearch">
        <template slot="no-options">
            Начните печатать название реки
        </template>
        <template slot="option" slot-scope="option">
            <div class="d-center">
                {{ option.title }}
            </div>
        </template>
        <template slot="selected-option" slot-scope="option">
            <div class="selected d-center">
                {{ option.title }}
            </div>
        </template>
    </v-select>
</template>

<script>
    import {doGetJson} from "../../api";
    import {backendApiBase} from "../../config";

    module.exports = {
        props: ['value'],
        methods: {
            onSearch: function (search, loading) {
                loading(true);
                var component = this;
                doGetJson(backendApiBase + '/river?q=' + search).then(json => {
                    component.options = json;
                    loading(false);
                }, err => {
                    console.error(err);
                    component.options = [];
                    loading(false);
                });
            },
        },
        computed: {
            river: {
                get() {
                    return this.value
                },
                set(river) {
                    this.$emit('input', river)
                },
            }
        },
        data() {
            return {
                options: [],
            }
        },
    }
</script>