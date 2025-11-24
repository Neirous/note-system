<template>
  <div class="trash-container">
    <div class="trash-header">
      <div class="title">回收站</div>
      <el-input v-model="keyword" size="small" placeholder="搜索标题" clearable class="search"/>
    </div>
    <el-table :data="filtered" height="100%" stripe>
      <el-table-column prop="title" label="标题" min-width="200"/>
      <el-table-column prop="updated_at" label="删除时间" width="180"/>
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="onRestore(row)">恢复</el-button>
          <el-popconfirm title="永久删除后不可恢复" confirm-button-text="删除" cancel-button-text="取消" @confirm="onHardDelete(row)">
            <template #reference>
              <el-button size="small" type="danger">永久删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { getTrashList, restoreNote, hardDeleteNote } from '../api/note'
import { ElMessage } from 'element-plus'

const list = ref([])
const keyword = ref('')

const load = async () => {
  const res = await getTrashList(1, 200)
  list.value = res.data.data.list || []
}

const filtered = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  if (!k) return list.value
  return list.value.filter(i => (i.title || '').toLowerCase().includes(k))
})

const onRestore = async (row) => {
  await restoreNote(row.id)
  ElMessage.success('已恢复')
  window.dispatchEvent(new CustomEvent('note-restored', { detail: { id: row.id } }))
  await load()
}

const onHardDelete = async (row) => {
  await hardDeleteNote(row.id)
  ElMessage.success('已永久删除')
  window.dispatchEvent(new CustomEvent('note-hard-deleted', { detail: { id: row.id } }))
  await load()
}

onMounted(load)
</script>

<style scoped>
.trash-container { height: 100%; display: flex; flex-direction: column; }
.trash-header { height: 48px; display: flex; align-items: center; padding: 0 12px; gap: 12px; border-bottom: 1px solid #e6e6e6; }
.title { font-weight: 600; }
.search { width: 240px; }
:deep(.el-table) { flex: 1; }
</style>
