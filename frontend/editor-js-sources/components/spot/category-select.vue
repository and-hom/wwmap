<template>
    <select v-model="category">
        <option v-for="cat in all_categories" v-bind:value="cat.id">{{cat.title}}
        </option>
    </select>
</template>
<script>
    import {all_categories} from '../../editor'

    module.exports = {
        props: ['value'],
        computed: {
            category: {
                get() {
                    return this.value;
                },
                set(category) {
                    this.$emit('input', category)
                }
            }
        },
        data: function () {
            return {
                all_categories: all_categories,
            };
        }
    }
</script>