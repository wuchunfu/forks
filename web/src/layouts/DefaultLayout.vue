<template>
  <n-layout>
    <NavBar />
    <n-layout-content class="main-content" content-style="padding-top: 64px;">
      <div class="content-wrapper">
        <router-view v-slot="{ Component }">
          <transition
            mode="out-in"
            @before-enter="beforeEnter"
            @enter="enter"
            @leave="leave"
          >
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </n-layout-content>
  </n-layout>
</template>
<script setup>
import { NLayout, NLayoutContent } from 'naive-ui'
import NavBar from './NavBar.vue'
function beforeEnter(el) {
  el.style.opacity = '0'
  el.style.transform = 'translateY(30px)'
}
function enter(el, done) {
  el.style.transition = 'all 0.4s ease'
  el.style.opacity = '1'
  el.style.transform = 'translateY(0)'
  setTimeout(done, 400)
}
function leave(el, done) {
  el.style.transition = 'all 0.3s ease'
  el.style.opacity = '0'
  el.style.transform = 'translateY(-30px)'
  setTimeout(done, 300)
}
</script>

<style scoped>
.content-wrapper {
  width: 100%;
  padding: 0 24px;
}

.main-content {
  min-height: calc(100vh - 64px);
}
</style> 