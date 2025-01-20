<script lang="ts" setup>
import { Refresh } from '@element-plus/icons-vue'
import axios from 'axios'
import { ElMessage } from 'element-plus' // 引入 ElMessage 用于显示提示信息
import { onMounted, onUnmounted, ref } from 'vue'

const props = defineProps<{
  config: {
    apiUrl: string
  }
  jsonData: {
    devName: string
    devDesc: string
    devId: string
    instId: string
  }
}>()

// 使用 props.jsonData 中的数据
console.log(props.jsonData.devName)

const devName = ref(props.jsonData.devName)
const instId = ref(props.jsonData.instId)
const devDesc = ref(props.jsonData.devDesc)
const devId = ref(props.jsonData.devId)

// 定义表格数据的类型
interface TagData {
  tagName: string
  timeStr: string
  value: string | number | boolean
  utc: number
}

const search = ref('') // 搜索关键字

// 定义 tableData 的类型为 TagData[]
const tableData = ref<TagData[]>([])

// 定义定时器 ID
let intervalId: number | null = null

// 获取数据的函数
async function fetchData() {
  try {
    // 定义 POST 请求的请求体
    const requestBody = {
      devid: props.jsonData.devId, // 使用 props.jsonDev.devId 'DEV_4vyYRDmIkIrQbOWD'
    }

    // 发送 POST 请求
    const response = await axios.post(props.config.apiUrl, requestBody, {
      headers: {
        'Content-Type': 'application/json', // 设置请求头为 JSON 格式
      },
    })

    // 获取后端返回的数据
    const data = response.data.data

    // 将数据转换为表格所需的格式
    tableData.value = Object.entries(data).map(([tagName, valueArray]) => {
      // 显式定义 valueArray 的类型
      const [timeStr, value, utc] = valueArray as [string, string | number | boolean, number]

      return {
        tagName,
        timeStr,
        value,
        utc,
      }
    })
  }
  catch (error) {
    console.error('加载表格数据失败:', error)
  }
}

// 组件挂载时启动定时器
onMounted(() => {
  // 立即获取一次数据
  fetchData()

  // 每隔 3 秒获取一次数据
  intervalId = setInterval(fetchData, 3000)
})

// 组件卸载时清除定时器
onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
  }
})

// 排序函数
function sortTimeStr(a: TagData, b: TagData) {
  return new Date(a.timeStr).getTime() - new Date(b.timeStr).getTime()
}

// 处理“历史”按钮点击事件
function handleHistoryClick(tagName: string) {
  ElMessage({
    message: `${tagName} 历史暂未实现`,
    type: 'info',
    duration: 3000, // 3 秒后消失
    center: true, // 提示信息居中
  })
}

// 处理“下置”按钮点击事件
function handleSetValueClick(tagName: string) {
  ElMessage({
    message: `${tagName} 下置暂未实现`,
    type: 'info',
    duration: 3000, // 3 秒后消失
    center: true, // 提示信息居中
  })
}
</script>

<template>
  <div class="container">
    <el-row :gutter="20">
      <el-col :span="2">
        <div>
          <el-input
            v-model="search"
            placeholder="请输入点名过滤"
            clearable
          />
        </div>
      </el-col>

      <el-col :span="4">
        <div>
          <el-input
            v-model="devName"
            style="max-width: 100%"
            disabled
            placeholder="Please input"
          >
            <template #prepend>
              名称：
            </template>
          </el-input>
        </div>
      </el-col>
      <el-col :span="4">
        <div>
          <el-input
            v-model="devDesc"
            style="max-width: 100%"
            disabled
            placeholder="Please input"
          >
            <template #prepend>
              描述：
            </template>
          </el-input>
        </div>
      </el-col>
      <el-col :span="5">
        <div>
          <el-input
            v-model="devId"
            style="max-width: 100%"
            disabled
            placeholder="Please input"
          >
            <template #prepend>
              设备ID：
            </template>
          </el-input>
        </div>
      </el-col>
      <el-col :span="5">
        <div>
          <el-input
            v-model="instId"
            style="max-width: 100%"
            disabled
            placeholder="Please input"
          >
            <template #prepend>
              实例：
            </template>
          </el-input>
        </div>
      </el-col>
      <el-col :span="1">
        <div>
          <el-button type="primary" :icon="Refresh" @click="fetchData">
            刷新
          </el-button>
        </div>
      </el-col>
    </el-row>

    <!-- 表格 -->
    <el-table :data="tableData.filter(data => !search || data.tagName.includes(search))" style="width: 96%">
      <!-- 第1列：名称 -->
      <el-table-column prop="tagName" label="名称" />
      <!-- 第2列：时间（支持排序） -->
      <el-table-column
        prop="timeStr"
        label="时间"
        sortable
        :sort-method="sortTimeStr"
      />
      <!-- 第3列：数值 -->
      <el-table-column prop="value" label="数值" />
      <!-- 第4列：操作 -->
      <el-table-column label="操作" width="200">
        <template #default="scope">
          <el-button size="small" type="primary" @click="handleHistoryClick(scope.row.tagName)">
            历史
          </el-button>
          <el-button size="small" type="success" @click="handleSetValueClick(scope.row.tagName)">
            下置
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style>
.container {
  padding-left: 10px; /* 上下左右各10px的间距 */
  padding-right: 10px;
  padding-top: 0px;
  padding-bottom: 0px;
}
.el-row {
  margin-bottom: 20px;
}
.el-row:last-child {
  margin-bottom: 0;
}
.el-col {
  border-radius: 4px;
}

.grid-content {
  border-radius: 4px;
  min-height: 36px;
}
</style>
