<template>
    <div>
      <btn-bar v-if="canEdit" logObjectType="REGION" :logObjectId="region.id">
        <slot></slot>
      </btn-bar>
      <breadcrumbs :country="country" :region="region"/>
        <h1>{{ region.title }}</h1>
    </div>
</template>

<script>
    import {store} from '../../app-state';
    import {hasRole, ROLE_ADMIN, ROLE_EDITOR} from '../../auth';

    module.exports = {
        props: ['region', 'country'],
        created: function() {
            hasRole(ROLE_ADMIN, ROLE_EDITOR).then(canEdit => this.canEdit = canEdit);
        },
        data:function() {
            return {
              canEdit: false,
            }
        }
    }

</script>