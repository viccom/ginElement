<script lang="ts" setup>
import { Download, Finished, Plus, Refresh, Upload } from '@element-plus/icons-vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { onMounted, onUnmounted, ref } from 'vue'

interface TagDataItem {
  pointName: string
  description: string
  type: string
  prop1: string
  prop2: string
  prop3: string
  prop4: string
  // 根据实际数据扩展其他属性
}

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

const devName = ref(props.jsonData.devName)
const instId = ref(props.jsonData.instId)
const devDesc = ref(props.jsonData.devDesc)
const devId = ref(props.jsonData.devId)

const tableData = ref<TagDataItem[]>([])
const search = ref('') // 新增搜索功能

// 分页相关数据
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 定时器ID
const intervalId: number | null = null

// 数据获取函数
async function fetchData() {
  try {
    const requestBody = {
      devList: [props.jsonData.devId],
    }
    const response = await axios.post(props.config.apiUrl, requestBody)
    const devData = response.data.data[props.jsonData.devId]
    tableData.value = Object.entries(devData).map(([_, values]) => ({
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
  // intervalId = setInterval(fetchData, 3000) // 每3秒刷新数据
})

// 组件卸载时清除定时器
onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
  }
})

// 操作列按钮处理
function handleModifyClick(row: TagDataItem) {
  currentRowIndex.value = tableData.value.findIndex(item => item.pointName === row.pointName)
  currentRowData.value = { ...row }
  isDialogVisible.value = true
}

function handleDeleteClick(pointName: string) {
  // 新增：根据pointName过滤删除当前行
  tableData.value = tableData.value.filter(item => item.pointName !== pointName)
  ElMessage({
    message: `${pointName} 已成功删除`,
    type: 'success',
    duration: 3000,
    center: true,
  })
}

async function handleImport(file: File) {
  const reader = new FileReader()
  try {
    reader.onload = (e) => {
      const text = e.target?.result as string
      const rows = text.split('\n')
      const parsedData: TagDataItem[] = []
      let errorCount = 0 // 新增错误计数器

      rows.forEach((row, index) => {
        const cols = row.split(',')
        if (cols.length < 3) {
          errorCount++
          ElMessage.error(`第${index + 1}行数据不足，需至少包含点名、描述、类型`)
          return
        }
        parsedData.push({
          pointName: (cols[0] || '').replace(/\s/g, ''),
          description: (cols[1] || '').replace(/\s/g, ''),
          type: (cols[2] || '').replace(/\s/g, ''),
          prop1: (cols[3] || '').replace(/\s/g, '') || '',
          prop2: (cols[4] || '').replace(/\s/g, '') || '',
          prop3: (cols[5] || '').replace(/\s/g, '') || '',
          prop4: (cols[6] || '').replace(/\s/g, '') || '',
        })
      })
      console.log('导入的点表数据：', parsedData)
      // 直接替换数据，无论是否有错误
      tableData.value = parsedData

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

function exportCSV() {
  const csvContent = tableData.value.map(({ pointName, description, type }) => `${pointName},${description},${type}`).join('\n')
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.download = `${devName.value}-${devId.value}-tags.csv` // 新增动态文件名
  link.href = URL.createObjectURL(blob)
  link.click()
}

async function saveData() {
  try {
    // 1. 构建tagsMap结构
    const tagsMap = tableData.value.reduce((acc, item) => {
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
      await fetchData() // 刷新数据
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

// 新增响应式变量
const isDialogVisible = ref(false)
const currentRowData = ref<TagDataItem>({ pointName: '', description: '', type: '', prop1: '', prop2: '', prop3: '', prop4: '' })
const currentRowIndex = ref(-1)

// 新增确认修改函数
function confirmModify() {
  // 新增的验证逻辑开始
  if (!currentRowData.value.pointName.trim()) {
    ElMessage.error('点名不能为空')
    return
  }
  if (!currentRowData.value.description.trim()) {
    ElMessage.error('描述不能为空')
    return
  }
  if (!currentRowData.value.type.trim()) {
    ElMessage.error('类型不能为空')
    return
  }

  const pointNameRegex = /^[a-z]\w*$/i
  if (!pointNameRegex.test(currentRowData.value.pointName)) {
    ElMessage.error('点名必须以字母开头，仅包含字母/数字/下划线')
    return
  }
  // 新增的验证逻辑结束

  if (currentRowIndex.value !== -1) {
    tableData.value[currentRowIndex.value] = { ...currentRowData.value }
    ElMessage.success('修改成功')
  }
  isDialogVisible.value = false
}

// 新增响应式变量
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

// 新增方法
function handleAddClick() {
  isAddDialogVisible.value = true
}

function confirmAdd() {
  // 新增必填项验证逻辑
  if (!newRowData.value.pointName.trim()) {
    ElMessage.error('点名不能为空')
    return
  }
  if (!newRowData.value.description.trim()) {
    ElMessage.error('描述不能为空')
    return
  }
  if (!newRowData.value.type.trim()) {
    ElMessage.error('类型不能为空')
    return
  }

  // 新增点名格式验证
  const pointNameRegex = /^[a-z]\w*$/i
  if (!pointNameRegex.test(newRowData.value.pointName)) {
    ElMessage.error('点名必须以字母开头，仅包含字母/数字/下划线')
    return
  }

  tableData.value.push({ ...newRowData.value })
  isAddDialogVisible.value = false
  ElMessage.success('新增成功')
  // 重置输入数据
  newRowData.value = {
    pointName: '',
    description: '',
    type: '',
    prop1: '',
    prop2: '',
    prop3: '',
    prop4: '',
  }
}

// 新增多选功能
const multipleSelection = ref<TagDataItem[]>([])

function handleSelectionChange(val: TagDataItem[]) {
  multipleSelection.value = val
}

function handleMultipleDelete() {
  if (multipleSelection.value.length === 0) {
    ElMessage.warning('请选择要删除的行')
    return
  }
  const pointNamesToDelete = multipleSelection.value.map(item => item.pointName)
  tableData.value = tableData.value.filter(item => !pointNamesToDelete.includes(item.pointName))
  ElMessage({
    message: `已成功删除 ${pointNamesToDelete.length} 行`,
    type: 'success',
    duration: 3000,
    center: true,
  })
}
</script>

<template>
  <el-row :gutter="20">
    <el-col :span="3">
      <div>
        <el-input
          v-model="search"
          placeholder="请输入点名过滤"
          clearable
        />
      </div>
    </el-col>

    <el-col :span="5">
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
    <el-col :span="5">
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
  </el-row>

  <el-table
    :data="tableData.filter(data => !search || data.pointName.includes(search)).slice((currentPage - 1) * pageSize, currentPage * pageSize)"
    style="width: 100%"
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
        <el-button
          size="small"
          type="primary"
          @click="handleModifyClick(scope.row)"
        >
          修改
        </el-button>
        <el-button
          size="small"
          type="danger"
          @click="handleDeleteClick(scope.row.pointName)"
        >
          删除
        </el-button>
      </template>
    </el-table-column>
  </el-table>

  <!-- 新增分页组件 -->
  <el-pagination
    v-model:current-page="currentPage"
    v-model:page-size="pageSize"
    :total="total"
    layout="total, sizes, prev, pager, next"
    :page-sizes="[10, 20, 50, 100]"
  />

  <!-- 新增按钮组 -->
  <el-row style="margin-top: 20px">
    <el-col :span="4">
      <el-button type="primary" :icon="Refresh" @click="fetchData">
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

  <!-- 新增对话框组件 -->
  <el-dialog
    v-model="isDialogVisible"
    title="修改数据"
    width="30%"
  >
    <el-form :model="currentRowData">
      <el-form-item label="点名" required>
        <el-input v-model="currentRowData.pointName" />
      </el-form-item>
      <el-form-item label="描述" required>
        <el-input v-model="currentRowData.description" />
      </el-form-item>
      <el-form-item label="类型" required>
        <el-select v-model="currentRowData.type" placeholder="请选择类型" style="width: 100%">
          <el-option
            v-for="option in ['int', 'float', 'string', 'bool']"
            :key="option"
            :label="option"
            :value="option"
          />
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

  <!-- 新增对话框 -->
  <el-dialog
    v-model="isAddDialogVisible"
    title="新增记录"
    width="30%"
  >
    <el-form :model="newRowData">
      <el-form-item label="点名" required>
        <el-input v-model="newRowData.pointName" />
      </el-form-item>
      <el-form-item label="描述" required>
        <el-input v-model="newRowData.description" />
      </el-form-item>
      <el-form-item label="类型" required>
        <el-select v-model="newRowData.type" placeholder="请选择类型" style="width: 100%">
          <el-option
            v-for="option in ['int', 'float', 'string', 'bool']"
            :key="option"
            :label="option"
            :value="option"
          />
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
</template>
