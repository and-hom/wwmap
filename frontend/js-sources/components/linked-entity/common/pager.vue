<template>
  <div>
    <div style="margin-top: 10px; display: flex;">
      <ul class="pagination" style="margin-right: 15px;">
        <li class="page-item" v-if="isNotBegin(pages)"><a class="page-link" href="#" v-on:click.stop="toBegin()">&lt;&lt;</a>
        </li>
        <li class="page-item" v-if="isNotBegin(pages)"><a class="page-link" href="#" v-on:click.stop="left()">&lt;</a>
        </li>
        <li :class="pageBtnClass(page)" v-if="!isSinglePage(pages)" v-for="page in pages"><a class="page-link" href="#"
                                                                                   v-on:click.stop="toPage(page)">{{
            page + 1
          }}</a></li>
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
    customFilter: {
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
        let lower = this.currentPage - 3;
        let upper = this.currentPage + 3;
        let btnCount = upper - lower + 1;

        let start = lower;
        if (start < 0) {
          start = 0;
        } else if (upper >= this.pageCount) {
          start = this.pageCount - btnCount;
        }

        for (let i = start; i < start + btnCount; i++) {
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
        return Math.ceil(this.filteredData.length / this.pageSize);
      }
    },
  },
  watch: {
    filter: function (newVal, _) {
      this.refreshData(this.data, newVal, this.customFilter);
    },
    customFilter: function (newVal, _) {
      this._customFilter = newVal;
      this.refreshData(this.data, this.filter, newVal);
    },
    data: function (newVal, _) {
      this.refreshData(newVal, this.filter, this.customFilter)
    },
  },
  created() {
    this.refreshData(this.data, this.filter, this.customFilter)
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
      } else if (page && page > this.pageCount) {
        this.currentPage = this.pageCount - 1;
      } else {
        this.currentPage = 0;
      }
    },
    isValidPage(page) {
      return page >= 0 && page < this.pageCount;
    },
    isSinglePage(pages) {
      return !this.pages || this.pages.length <= 1;
    },
    isNotBegin(pages) {
      return pages[0] != 0
    },
    isNotEnd(pages) {
      return pages[pages.length - 1] != (this.pageCount - 1);
    },
    refreshData(data, filter, customFilter) {
      if (this.filterFunction && filter) {
        let page = this.currentPage;
        this.filteredData = this.filterFunction(data, filter, customFilter);
        this.toPage(page);
      }
    },
    pageBtnClass(page) {
      return this.currentPage == page ? 'page-item active' : 'page-item';
    },
  }
}
</script>
