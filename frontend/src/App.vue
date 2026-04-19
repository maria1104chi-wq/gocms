<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup>
// 根组件
// 文件名: App.vue
// 路径: /workspace/frontend/src/App.vue
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: #f5f7fa;
  color: #333;
}

#app {
  min-height: 100vh;
}

a {
  text-decoration: none;
  color: inherit;
}
</style>
