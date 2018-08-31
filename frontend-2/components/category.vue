<templatg>
    <div style="display:inline">{{ categoryString() }}</div>
</template>

<script type="text/javascript">
    module.exports = {
        props: ['category'],
        data: function () {
            return {
                categoryString: function() {
                    if (this.category=="0") {
                        return "-"
                    }
                    if (this.category=="-1") {
                        return "Непроход."
                    }
                    return this.category
                }
            }
        },
    }
</script>