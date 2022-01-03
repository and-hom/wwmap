<template>
  <div>
    <btn-bar v-if="canEdit" logObjectType="COUNTRY" :logObjectId="country.id">
      <slot></slot>
    </btn-bar>
    <breadcrumbs :country="country"/>
    <h1>{{ country.title }}</h1>
    <span class="props-title-span">
       Двухбуквенный код <a href="https://ru.wikipedia.org/wiki/%D0%9E%D0%B1%D1%89%D0%B5%D1%80%D0%BE%D1%81%D1%81%D0%B8%D0%B9%D1%81%D0%BA%D0%B8%D0%B9_%D0%BA%D0%BB%D0%B0%D1%81%D1%81%D0%B8%D1%84%D0%B8%D0%BA%D0%B0%D1%82%D0%BE%D1%80_%D1%81%D1%82%D1%80%D0%B0%D0%BD_%D0%BC%D0%B8%D1%80%D0%B0" target="_blank">ОКСМ</a>:
    </span><span class="props-value-span">{{ country.code }}</span>
  </div>
</template>

<style>
  .props-title-span {
    font-weight: bold
  }
  .props-value-span {
    padding-left: 15px
  }
</style>

<script>
import {hasRole, ROLE_ADMIN} from '../../auth';

module.exports = {
  props: ['country'],
  created: function () {
    hasRole(ROLE_ADMIN).then(canEdit => this.canEdit = canEdit);
  },
  data: function () {
    return {
      canEdit: false,
    }
  }
}

</script>
