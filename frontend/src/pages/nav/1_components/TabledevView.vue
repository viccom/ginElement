<script lang="ts" setup>
import { Download, Finished, Plus, Refresh, Upload } from '@element-plus/icons-vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
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

interface TagDataItem {
  pointName: string
  description: string
  type: string
  prop1: string
  prop2: string
  prop3: string
  prop4: string
}
// 数据表搜索相关数据
const datasearch = ref('') // 搜索关键字
// 点表搜索相关数据
const tagsearch = ref('') // 搜索关键字

// 定义 tableData 的类型为 TagData[]
const tableData = ref<TagData[]>([])

// 定义 tableData 的类型为 TagData[]
const tableTag = ref<TagDataItem[]>([])

// 分页相关数据
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 定时器 ID
let intervalId: number | null = null

// 新增：处理每页记录数变化
function handleSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1 // 重置到第一页
}

// 处理分页变化
function handlePageChange(page: number) {
  currentPage.value = page
}

// 获取数据的函数
async function fetchData() {
  try {
    // 定义 POST 请求的请求体
    const requestBody = {
      devid: props.jsonData.devId, // 使用 props.jsonDev.devId 'DEV_4vyYRDmIkIrQbOWD'
    }

    // 发送 POST 请求
    const response = await axios.post(`/api/v1/getDevvalues`, requestBody, {
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

// 点表获取函数
async function fetchTagData() {
  try {
    const requestBody = {
      devList: [props.jsonData.devId],
    }
    const response = await axios.post('/api/v1/getDevtags', requestBody)
    const devData = response.data.data[props.jsonData.devId]
    tableTag.value = Object.entries(devData).map(([_, values]) => ({
      pointName: values[0],
      description: values[1],
      type: values[2],
      prop1: values[3],
      prop2: values[4],
      prop3: values[5],
      prop4: values[6],
    }))
    total.value = tableData.value.length
  }
  catch (error) {
    console.error('加载点表数据失败:', error)
    ElMessage({
      message: `加载点表数据失败: ${error} `,
      type: 'error',
      duration: 3000,
      center: true,
    })
  }
}

// 组件挂载时启动定时器
onMounted(() => {
  fetchData()
  fetchTagData()
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
    duration: 3000,
    center: true,
  })
}

// 处理“下置”按钮点击事件
function handleSetValueClick(tagName: string) {
  ElMessage({
    message: `${tagName} 下置暂未实现`,
    type: 'info',
    duration: 3000,
    center: true,
  })
}

// 新增记录
const isAddDialogVisible = ref(false)
const newRowData = ref<TagDataItem>({
  pointName: '',
  description: '',
  type: '',
  prop1: '',
  prop2: '',
  prop3: '',
  prop4: '',
})

function handleAddClick() {
  newRowData.value = {
    pointName: '',
    description: '',
    type: '',
    prop1: '',
    prop2: '',
    prop3: '',
    prop4: '',
  }
  isAddDialogVisible.value = true
}

function confirmAdd() {
  tableTag.value.push({ ...newRowData.value })
  isAddDialogVisible.value = false
  total.value = tableTag.value.length // 更新总记录数
  ElMessage({
    message: `${newRowData.value.pointName} 已成功添加`,
    type: 'success',
    duration: 3000,
    center: true,
  })
}

// 批量删除
const selectedRows = ref<(TagData | TagDataItem)[]>([])

function handleSelectionChange(rows: (TagData | TagDataItem)[]) {
  selectedRows.value = rows
}

function handleMultipleDelete() {
  if (selectedRows.value.length === 0) {
    ElMessage({
      message: '请选择要删除的记录',
      type: 'warning',
      duration: 3000,
      center: true,
    })
    return
  }
  tableTag.value = tableTag.value.filter(
    item => !selectedRows.value.some(selected => 'pointName' in selected && selected.pointName === (item as TagDataItem).pointName),
  )
  total.value = tableTag.value.length // 更新总记录数
  ElMessage({
    message: `已成功删除 ${selectedRows.value.length} 条记录`,
    type: 'success',
    duration: 3000,
    center: true,
  })
}

// 保存数据
async function saveData() {
  try {
    // 1. 构建tagsMap结构
    const tagsMap = tableTag.value.reduce((acc, item) => {
      // 新增条件判断：跳过pointName为空的无效项
      if (!item.pointName.trim()) {
        return acc // 直接返回当前累加器，不处理该行
      }
      acc[item.pointName] = [
        item.pointName,
        item.description,
        item.type,
        item.prop1,
        item.prop2,
        item.prop3,
        item.prop4,
      ]
      return acc
    }, {} as Record<string, any[]>)

    // 新增：检查tagsMap是否为空
    if (Object.keys(tagsMap).length === 0) {
      ElMessage.error('表格中没有数据，无法保存')
      return
    }

    // 2. 构建完整请求数据
    const requestData = {
      devId: devId.value,
      instid: instId.value, // 注意字段名是instid
      tagsMap,
    }

    // 3. 发送POST请求
    const response = await axios.post('/api/v1/newDevtags', requestData)

    // 4. 处理响应
    if (response.data.result === 'success') {
      ElMessage.success('设备新增采集点成功')
      await fetchTagData() // 刷新数据
    }
    else {
      ElMessage.error(`保存失败: ${response.data.message}`)
    }
  }
  catch (error) {
    console.error('保存失败:', error)
    ElMessage.error('网络请求失败，请检查连接')
  }
}

// 导出CSV
function exportCSV() {
  const csvContent = tableTag.value.map(item => Object.values(item).join(',')).join('\n')
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `${devName.value}-${devId.value}-tags.csv`
  link.click()
}

// 导入CSV

async function handleImport(file: File) {
  const reader = new FileReader()
  try {
    reader.onload = (e) => {
      const text = e.target?.result as string
      const rows = text.split('\n')
      const parsedData: TagDataItem[] = []
      let errorCount = 0 // 新增错误计数器

      rows.forEach((row, index) => {
        console.log('Index:', index, 'Row:', row)
        const cols = row.split(',')
        if (cols.length < 3) {
          errorCount++
          ElMessage.error(`第${index + 1}行数据不足，需至少包含点名、描述、类型`)
          return
        }
        parsedData.push({
          pointName: (cols[0] || '').trim(),
          description: (cols[1] || '').trim(),
          type: (cols[2] || '').trim(),
          prop1: (cols[3] || '').trim() || '',
          prop2: (cols[4] || '').trim() || '',
          prop3: (cols[5] || '').trim() || '',
          prop4: (cols[6] || '').trim() || '',
        })
      })
      // console.log('导入的点表数据：', parsedData)
      // 直接替换数据，无论是否有错误
      tableTag.value = parsedData
      total.value = tableTag.value.length // 更新总记录数
      // 根据错误情况显示不同提示
      if (errorCount > 0) {
        ElMessage.warning(`共${errorCount}行数据不符合要求，已跳过`)
      }
      else {
        ElMessage.success('导入成功，数据已替换')
      }
    }

    reader.onerror = (error) => {
      ElMessage.error('文件读取失败，请检查文件格式')
      console.error('文件读取错误:', error)
    }

    reader.readAsText(file)
  }
  catch (error) {
    ElMessage.error('导入过程中发生错误')
    console.error('导入错误:', error)
  }
}

// 修改数据
const isDialogVisible = ref(false)
const currentRowData = ref<TagDataItem>({
  pointName: '',
  description: '',
  type: '',
  prop1: '',
  prop2: '',
  prop3: '',
  prop4: '',
})
const currentRowIndex = ref(-1)

function confirmModify() {
  tableTag.value[currentRowIndex.value] = { ...currentRowData.value }
  isDialogVisible.value = false
  ElMessage({
    message: `${currentRowData.value.pointName} 已成功修改`,
    type: 'success',
    duration: 3000,
    center: true,
  })
}

// 操作列按钮处理
function handleModifyClick(row: TagDataItem) {
  currentRowIndex.value = tableTag.value.findIndex(item => item.pointName === row.pointName)
  currentRowData.value = { ...row }
  isDialogVisible.value = true
}

function handleDeleteClick(pointName: string) {
  // 新增：根据pointName过滤删除当前行
  tableTag.value = tableTag.value.filter(item => item.pointName !== pointName)
  total.value = tableTag.value.length // 更新总记录数
  ElMessage({
    message: `${pointName} 已成功删除`,
    type: 'success',
    duration: 3000,
    center: true,
  })
}

// 新增浮动面板相关变量
const panelVisible = ref(false)
const selectedIsRunning = ref(false)

// 新增显示浮动面板的方法
function showPanel(isRunning: boolean) {
  selectedIsRunning.value = isRunning
  panelVisible.value = true
}

// 新增启动方法
async function handleStart() {
  try {
    const response = await axios.post('/api/v1/startApp', {
      instid: instId.value,
    })
    if (response.data.data) {
      ElMessage.success(`启动成功: ${instId.value}`)
      fetchData() // 刷新表格数据
    }
    else {
      ElMessage.error(`启动失败: ${response.data.details || '未知错误'}`)
    }
  }
  catch (error) {
    ElMessage.error(`启动失败: ${error.response?.data?.details || '网络错误'}`)
  }
  finally {
    panelVisible.value = false
  }
}

// 新增停止方法
async function handleStop() {
  try {
    const response = await axios.post('/api/v1/stopApp', {
      instid: instId.value,
    })
    if (response.data.data) {
      ElMessage.success(`停止成功: ${instId.value}`)
      fetchData() // 刷新表格数据
    }
    else {
      ElMessage.error(`停止失败: ${response.data.details || '未知错误'}`)
    }
  }
  catch (error) {
    ElMessage.error(`停止失败: ${error.response?.data?.details || '网络错误'}`)
  }
  finally {
    panelVisible.value = false
  }
}

// 新增重启方法
async function handleRestart() {
  try {
    const response = await axios.post('/api/v1/restartApp', {
      instid: instId.value,
    })
    if (response.data.data) {
      ElMessage.success(`重启成功: ${instId.value}`)
      fetchData() // 刷新表格数据
    }
    else {
      ElMessage.error(`重启失败: ${response.data.details || '未知错误'}`)
    }
  }
  catch (error) {
    ElMessage.error(`重启失败: ${error.response?.data?.details || '网络错误'}`)
  }
  finally {
    panelVisible.value = false
  }
}

</script>

<template>
  <div class="container">
    <el-tabs class="demo-tabs">
      <el-tab-pane label="数据">
        <el-row :gutter="20">
          <el-col :span="2">
            <div>
              <el-input
                v-model="datasearch"
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
              <el-popover placement="bottom" trigger="hover">
                <template #reference>
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
                </template>
                <el-menu :default-active="selectedIsRunning ? 'stop' : 'start'" class="el-menu-demo" mode="vertical">
                  <el-menu-item v-if="!selectedIsRunning" index="start" @click="handleStart" class="start-button">启动</el-menu-item>
                  <el-menu-item v-if="selectedIsRunning" index="stop" @click="handleStop" class="stop-button">停止</el-menu-item>
                  <el-menu-item index="restart" @click="handleRestart" class="restart-button">重启</el-menu-item>
                </el-menu>
              </el-popover>
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
        <el-table :data="tableData.filter(data => !datasearch || data.tagName.includes(datasearch)).slice((currentPage - 1) * pageSize, currentPage * pageSize)" style="width: 96%">
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

        <!-- 分页组件 -->
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          background
          :total="tableData.filter(data => !tagsearch || data.tagName.includes(tagsearch)).length"
          layout="total, sizes, prev, pager, next"
          :page-sizes="[10, 20, 50, 100]"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </el-tab-pane>

      <el-tab-pane label="点表">
        <el-row :gutter="20">
          <el-col :span="3">
            <div>
              <el-input v-model="tagsearch" placeholder="请输入点名过滤" clearable />
            </div>
          </el-col>
          <el-col :span="5">
            <div>
              <el-input v-model="devName" style="max-width: 100%" disabled placeholder="Please input">
                <template #prepend>
                  名称：
                </template>
              </el-input>
            </div>
          </el-col>
          <el-col :span="5">
            <div>
              <el-input v-model="devDesc" style="max-width: 100%" disabled placeholder="Please input">
                <template #prepend>
                  描述：
                </template>
              </el-input>
            </div>
          </el-col>
          <el-col :span="5">
            <div>
              <el-input v-model="devId" style="max-width: 100%" disabled placeholder="Please input">
                <template #prepend>
                  设备ID：
                </template>
              </el-input>
            </div>
          </el-col>
          <el-col :span="5">
            <div>
              <el-popover placement="bottom" trigger="hover">
                <template #reference>
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
                </template>
                <el-menu :default-active="selectedIsRunning ? 'stop' : 'start'" class="el-menu-demo" mode="vertical">
                  <el-menu-item v-if="!selectedIsRunning" index="start" @click="handleStart" class="start-button">启动</el-menu-item>
                  <el-menu-item v-if="selectedIsRunning" index="stop" @click="handleStop" class="stop-button">停止</el-menu-item>
                  <el-menu-item index="restart" @click="handleRestart" class="restart-button">重启</el-menu-item>
                </el-menu>
              </el-popover>
            </div>
          </el-col>
        </el-row>

        <el-table
          :data="tableTag.filter(data => !tagsearch || data.pointName.includes(tagsearch)).slice((currentPage - 1) * pageSize, currentPage * pageSize)"
          style="width: 96%"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="pointName" label="点名" />
          <el-table-column prop="description" label="描述" />
          <el-table-column prop="type" label="类型" />
          <el-table-column prop="prop1" label="属性1" />
          <el-table-column prop="prop2" label="属性2" />
          <el-table-column prop="prop3" label="属性3" />
          <el-table-column prop="prop4" label="属性4" />
          <el-table-column label="操作" width="150">
            <template #default="scope">
              <el-button size="small" type="primary" @click="handleModifyClick(scope.row)">
                修改
              </el-button>
              <el-button size="small" type="danger" @click="handleDeleteClick(scope.row.pointName)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          layout="total, sizes, prev, pager, next"
          :page-sizes="[10, 20, 50, 100]"
        />

        <el-row style="margin-top: 20px">
          <el-col :span="4">
            <el-button type="primary" :icon="Refresh" @click="fetchTagData">
              重置
            </el-button>
          </el-col>
          <el-col :span="4">
            <el-button type="primary" :icon="Plus" @click="handleAddClick">
              新增
            </el-button>
          </el-col>
          <el-col :span="4">
            <el-upload
              action="#"
              :show-file-list="false"
              :on-change="(uploadFile) => handleImport(uploadFile.raw)"
              accept=".csv"
            >
              <el-button type="primary" :icon="Upload">
                导入
              </el-button>
            </el-upload>
          </el-col>
          <el-col :span="4">
            <el-button type="warning" :icon="Finished" @click="saveData">
              保存
            </el-button>
          </el-col>
          <el-col :span="4">
            <el-button type="success" :icon="Download" @click="exportCSV">
              导出
            </el-button>
          </el-col>
          <el-col :span="4">
            <el-button type="danger" :icon="Finished" @click="handleMultipleDelete">
              批量删除
            </el-button>
          </el-col>
        </el-row>

        <el-dialog v-model="isDialogVisible" title="修改数据" width="30%">
          <el-form :model="currentRowData">
            <el-form-item label="点名" required>
              <el-input v-model="currentRowData.pointName" />
            </el-form-item>
            <el-form-item label="描述" required>
              <el-input v-model="currentRowData.description" />
            </el-form-item>
            <el-form-item label="类型" required>
              <el-select v-model="currentRowData.type" placeholder="请选择类型" style="width: 100%">
                <el-option v-for="option in ['int', 'float', 'string', 'bool']" :key="option" :label="option" :value="option" />
              </el-select>
            </el-form-item>
            <el-form-item label="属性1">
              <el-input v-model="currentRowData.prop1" />
            </el-form-item>
            <el-form-item label="属性2">
              <el-input v-model="currentRowData.prop2" />
            </el-form-item>
            <el-form-item label="属性3">
              <el-input v-model="currentRowData.prop3" />
            </el-form-item>
            <el-form-item label="属性4">
              <el-input v-model="currentRowData.prop4" />
            </el-form-item>
          </el-form>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="isDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="confirmModify">确定</el-button>
            </span>
          </template>
        </el-dialog>

        <el-dialog v-model="isAddDialogVisible" title="新增记录" width="30%">
          <el-form :model="newRowData">
            <el-form-item label="点名" required>
              <el-input v-model="newRowData.pointName" />
            </el-form-item>
            <el-form-item label="描述" required>
              <el-input v-model="newRowData.description" />
            </el-form-item>
            <el-form-item label="类型" required>
              <el-select v-model="newRowData.type" placeholder="请选择类型" style="width: 100%">
                <el-option v-for="option in ['int', 'float', 'string', 'bool']" :key="option" :label="option" :value="option" />
              </el-select>
            </el-form-item>
            <el-form-item label="属性1">
              <el-input v-model="newRowData.prop1" />
            </el-form-item>
            <el-form-item label="属性2">
              <el-input v-model="newRowData.prop2" />
            </el-form-item>
            <el-form-item label="属性3">
              <el-input v-model="newRowData.prop3" />
            </el-form-item>
            <el-form-item label="属性4">
              <el-input v-model="newRowData.prop4" />
            </el-form-item>
          </el-form>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="isAddDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="confirmAdd">确定</el-button>
            </span>
          </template>
        </el-dialog>
      </el-tab-pane>
    </el-tabs>
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
