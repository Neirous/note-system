<template>
  <el-menu class="icon-menu" :default-active="menuActive" router collapse>
    <el-menu-item index="/notes">
      <el-icon><Notebook /></el-icon>
    </el-menu-item>
    <el-menu-item index="/rag-qa">
      <el-icon><ChatLineSquare /></el-icon>
    </el-menu-item>
    <el-menu-item index="/trash">
      <el-icon><Delete /></el-icon>
    </el-menu-item>
  </el-menu>
</template>

<script setup>
import { useRoute } from 'vue-router'
import { computed } from 'vue'
import { Notebook, ChatLineSquare, Delete } from '@element-plus/icons-vue'
const route = useRoute()
const menuActive = computed(() => (route.path === '/' ? '' : route.path))
</script>

<style scoped>
.icon-menu { height: 100%; border-right: none; }
.icon-menu :deep(.el-menu-item) { justify-content: center; }
.icon-menu :deep(.el-icon) { font-size: 20px; }
</style>
