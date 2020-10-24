<template>
  <div>
    <div style="margin-top: 10px; display: flex;">
      <ul class="pagination" style="margin-right: 15px;">
        <li class="page-item" v-if="isNotBegin(pages)"><a class="page-link" href="#" v-on:click.stop="toBegin()">&lt;&lt;</a>
        </li>
        <li class="page-item" v-if="isNotBegin(pages)"><a class="page-link" href="#" v-on:click.stop="left()">&lt;</a>
        </li>
        <li class="page-item" v-for="page in pages"><a class="page-link" href="#"
                                                       v-on:click.stop="toPage(page)">{{ page + 1 }}</a></li>
        <li class="page-item" v-if="isNotEnd(pages)"><a class="page-link" href="#" v-on:click.stop="right()">&gt;</a>
        </li>
        <li class="page-item" v-if="isNotEnd(pages)"><a class="page-link" href="#"
                                                        v-on:click.stop="toEnd()">&gt;&gt;</a></li>
      </ul>
      <div style="width: 100%">
        <slot name="filter"></slot>
      </div>
    </div>
    <slot v-bind:data="pageData"></slot>
  </div>
</template>


<script>

module.exports = {
  props: {
    data: {
      type: Array,
      required: true,
    },
    pageSize: {
      type: Number,
      required: true,
    },
    filter: {
      required: false,
    },
    filterFunction: {
      type: Function,
      required: false,
    }
  },
  computed: {
    pages: {
      get() {
        let pages = Array();
        for (let i = this.currentPage - 3; i <= this.currentPage + 3; i++) {
          if (this.isValidPage(i)) {
            pages.push(i)
          }
        }
        return pages;
      }
    },
    pageData: {
      get() {
        let start = this.currentPage * this.pageSize;
        return this.filteredData.slice(start, start + this.pageSize);
      }
    },
    pageCount: {
      get() {
        return Math.floor(this.filteredData.length / this.pageSize);
      }
    },
  },
  watch: {
    filter: function (newVal, _) {
      this.refreshData(this.data, newVal)
    },
    data: function (newVal, _) {
      this.refreshData(newVal, this.filter)
    },
  },
  created() {
    this.refreshData(this.data, this.filter)
  },
  data() {
    return {
      currentPage: 0,
      filteredData: 0,
    }
  },
  methods: {
    toEnd() {
      this.currentPage = this.pageCount - 1;
    },
    toBegin() {
      this.currentPage = 0;
    },
    left() {
      if (this.currentPage > 0) {
        this.currentPage -= 1;
      }
    },
    right() {
      if (this.currentPage + 1 < this.pageCount) {
        this.currentPage += 1;
      }
    },
    toPage(page) {
      if (this.isValidPage(page)) {
        this.currentPage = page;
      }
    },
    isValidPage(page) {
      return page >= 0 && page < this.pageCount;
    },
    isNotBegin(pages) {
      return pages[0] != 0
    },
    isNotEnd(pages) {
      return pages[pages.length - 1] != (this.pageCount - 1);
    },
    refreshData(data, filter) {
      if (this.filterFunction && filter) {
        let page = this.currentPage;
        this.filteredData = this.filterFunction(data, filter);
        this.toPage(page);
      }
    },
  }
}
</script>