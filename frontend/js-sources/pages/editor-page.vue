<template>
    <div>
        <ask id="close-spot-editor" title="Отменить редактирование?"
             msg='Открыт редактор порога. Там могут быть ваши изменения. Для сохранения воспользуйтесь кнопкой "Сохранить" сверху. Закрыть редактор и сбросить изменения?'
             :ok-fn="function() { spoteditorstate.editMode = false; if(closeCallback) {closeCallback()}}"></ask>
        <ask id="close-river-editor" title="Отменить редактирование?"
             msg='Открыт редактор реки. Там могут быть ваши изменения. Для сохранения воспользуйтесь кнопкой "Сохранить" сверху. Закрыть редактор и сбросить изменения?'
             :ok-fn="function() { rivereditorstate.pageMode = 'view'; if(closeCallback) {closeCallback()} }"></ask>

        <page link="editor.htm">
            <div class="container-fluid" style="margin-top: 20px;">
                <div class="row">
                    <div class="col-3" id="left-menu">
                        <div v-if="canAddCountry" style="padding-left: 40px; padding-bottom: 15px;">
                            <button type="button" class="btn btn-primary" v-on:click="add_country()">Добавить страну</button>
                        </div>
                        <ul class="menu-items">
                            <country v-bind:key="country.id" v-bind:country="country" v-for="country in countries"/>
                        </ul>
                    </div>
                    <div id="editor-pane" class="col-9" style="bgcolor:red;">
                        <transition name="fade">
                            <div class="alert alert-danger" role="alert" v-if="errMsg">
                                {{errMsg}}
                            </div>
                        </transition>
                        <div>
                            <country-page v-if="countryeditorstate.visible"
                                          v-bind:initial-country="countryeditorstate.country"/>
                        </div>
                        <div>
                            <region-page v-if="regioneditorstate.visible"
                                         v-bind:initial-region="regioneditorstate.region"
                                         v-bind:country="regioneditorstate.country"/>
                        </div>
                        <div>
                            <river-page v-if="rivereditorstate.visible"
                                        v-bind:initial-river="rivereditorstate.river"
                                        v-bind:reports="rivereditorstate.reports"
                                        v-bind:transfers="rivereditorstate.transfers"
                                        v-bind:country="rivereditorstate.country"
                                        v-bind:region="rivereditorstate.region"
                                        v:sensors="sensors"/>
                        </div>
                        <div>
                            <spot-page v-if="spoteditorstate.visible"
                                       v-bind:initial-spot="spoteditorstate.spot"
                                       v-bind:country="spoteditorstate.country"
                                       v-bind:region="spoteditorstate.region"
                                       v-bind:river="spoteditorstate.river"
                                       :zoom="spoteditorstate.zoom"/>
                        </div>
                    </div>
                </div>
            </div>
        </page>
    </div>
</template>

<script>
    import {store, getById} from '../app-state'
    import {hasRole, ROLE_ADMIN} from "../auth";


    export default {
        data() {
            return {
              canAddCountry: false
            };
        },
        created: function () {
          hasRole(ROLE_ADMIN).then(canAddCountry => this.canAddCountry = canAddCountry);
        },
        computed: {
            countries() {
                return store.state.treePath;
            },
            closeCallback() {
                return store.state.closeCallback
            },
            errMsg() {
                return store.state.errMsg
            },
            countryeditorstate() {
                return store.state.countryeditorstate
            },
            regioneditorstate() {
                return store.state.regioneditorstate
            },
            rivereditorstate() {
                return store.state.rivereditorstate
            },
            spoteditorstate() {
                return store.state.spoteditorstate
            },
        },
        methods: {
          add_country: function () {
            store.commit('newCountry', {
            })
          },
        }
    }
</script>
