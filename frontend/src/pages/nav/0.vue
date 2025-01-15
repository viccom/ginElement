<script>
import axios from 'axios'
import { onMounted, ref } from 'vue'

export default {
  setup() {
    const activeTab = ref('tab1')
    const tableData = ref([])

    const fetchData = async () => {
      try {
        const response = await axios.get(`/api/v1/listApps?t=${Date.now()}`, {
          headers: { accept: 'application/json' },
        })
        const data = response.data.data
        tableData.value = Object.values(data).map(item => ({
          instName: item.instName,
          instId: item.instId,
          appCode: item.appCode,
          appType: item.appType,
          autoStart: item.autoStart,
        }))
        // eslint-disable-next-line no-console
        // console.log('tableData:', tableData.value) // 打印数据以验证
      }
      catch (error) {
        console.error('获取数据失败:', error)
      }
    }

    const handleDetail = (row) => {
      // eslint-disable-next-line no-console
      console.log('详情:', row)
    }

    const handleStart = (row) => {
      // eslint-disable-next-line no-console
      console.log('启动:', row)
    }

    const handleStop = (row) => {
      // eslint-disable-next-line no-console
      console.log('停止:', row)
    }

    const handleDelete = (row) => {
      // eslint-disable-next-line no-console
      console.log('删除:', row)
    }

    onMounted(() => {
      fetchData()
    })

    return {
      activeTab,
      tableData,
      handleDetail,
      handleStart,
      handleStop,
      handleDelete,
    }
  },
}
</script>

<template>
  <div>
    <el-tabs v-model="activeTab" type="card">
      <el-tab-pane label="标签页 1" name="tab1">
        <!-- 新增按钮和表格的容器 -->
        <div class="table-header">
          <el-button type="primary" @click="handleAddInstance">
            新增实例
          </el-button>
        </div>
        <el-table :data="tableData" style="width: 100%">
          <el-table-column prop="instName" label="实例名称" width="180" />
          <el-table-column prop="instId" label="实例ID" width="280" />
          <el-table-column prop="appCode" label="应用代码" width="160" />
          <el-table-column prop="appType" label="应用类型" width="160" />
          <el-table-column prop="autoStart" label="自启动" width="120">
            <template #default="scope">
              <el-tag :type="scope.row.autoStart ? 'success' : 'danger'">
                {{ scope.row.autoStart ? '是' : '否' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="280">
            <template #default="scope">
              <el-button size="small" @click="handleDetail(scope.row)">
                详情
              </el-button>
              <el-button size="small" type="success" @click="handleStart(scope.row)">
                启动
              </el-button>
              <el-button size="small" type="warning" @click="handleStop(scope.row)">
                停止
              </el-button>
              <el-button size="small" type="danger" @click="handleDelete(scope.row)">
                删除
              </el-button>
            </template>
          </el-table-column>
          <el-table-column width="0" />
        </el-table>
      </el-tab-pane>

    </el-tabs>
  </div>
</template>

<style scoped>
/* 你可以在这里添加一些自定义样式 */
/* 按钮和表格的布局 */
.table-header {
  display: flex;
  justify-content: flex-start; /* 按钮靠右 */
  margin-bottom: 16px; /* 按钮和表格之间的间距 */
}
</style>
