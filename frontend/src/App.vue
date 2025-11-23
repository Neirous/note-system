<template>
  <!-- 路由出口：所有路由对应的组件都会渲染在这里 -->
  <router-view />
</template>

<script setup>
// 无需额外逻辑，路由接管所有渲染
</script>

<style>
/* 全局样式重置：去掉默认边距，让布局占满屏幕 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

/* 让 html 和 body 占满屏幕高度，避免布局塌陷 */
html, body, #app {
  height: 100%;
}
</style>