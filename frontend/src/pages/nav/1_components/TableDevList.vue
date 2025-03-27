<script lang="ts" setup>
import axios from 'axios'
import { ElMessage } from 'element-plus' // 新增Element消息组件导入
import { onMounted, onUnmounted, ref } from 'vue'

// 定义设备数据的类型
interface DeviceData {
  devName: string
  devDesc: string
  devId: string
  devType: string
  instId: string
  isRunning: boolean | null // 允许 isRunning 为 null,
}

const props = defineProps<{
  config: {
    apiUrl: string
  }
  jsonData: {
    devName: string
    devDesc: string
    devId: string
    devType: string
    instId: string
  }
}>()

// 定义事件
const emit = defineEmits(['data-click', 'point-click'])

// 使用 props.jsonData 中的数据
// console.log(props.jsonData.devName)

// 定义 tableData 的类型为 DeviceData[]
const tableData = ref<DeviceData[]>([])

// 定义定时器 ID
let intervalId: number | null = null

// 获取数据的函数
async function fetchData() {
  try {
    // 定义 POST 请求的请求体
    const requestBody = {
      instid: '',
      devType: '',
    }

    // 发送 POST 请求
    const response = await axios.post(props.config.apiUrl, requestBody, {
      headers: {
        'Content-Type': 'application/json', // 设置请求头为 JSON 格式
      },
    })
    console.log('响应数据:', response.data)
    // 定义 data 的类型
    const data: Record<string, DeviceData> = response.data.data

    // 将响应数据赋值给 tableData
    tableData.value = Object.values(data).map((item: DeviceData) => ({
      devName: item.devName,
      devDesc: item.devDesc,
      devId: item.devId,
      devType: item.devType,
      instId: item.instId,
      isRunning: item.isRunning,
    }))
  }
  catch (error) {
    console.error('加载表格数据失败:', error)
  }
}

// 新增对话框相关变量
const dialogVisible = ref(false)
const formData = ref({
  devName: '',
  devDesc: '',
  config: '',
  instId: '',
  devId: '',
  devType: '01', // 设置默认值
})
const appOptions = ref([])

// 新增获取实例列表的函数
async function fetchApps() {
  try {
    const response = await axios.get('/api/v1/listApps?appType=toSouth')
    const appsData = response.data.data
    appOptions.value = Object.values(appsData).map(app => ({
      value: app.instId,
      label: app.instName,
    }))
  }
  catch (error) {
    console.error('获取实例列表失败:', error)
  }
}

// 新增获取实例列表的函数
async function addNewDevice() {
  await fetchApps()
  dialogVisible.value = true
}

// 新增表单提交处理
async function handleFormSubmit() {
  if (!formData.value.devName) {
    ElMessage.error('设备名称不能为空')
    return
  }
  if (!formData.value.instId) {
    ElMessage.error('实例ID不能为空')
    return
  }
  // 新增检查名称重复
  checkDeviceName()
  if (nameError.value) {
    ElMessage.error('设备名称已存在')
    return
  }
  try {
    const response = await axios.post('/api/v1/newDev', formData.value)
    if (response.status === 200) {
      ElMessage.success('设备添加成功')
      dialogVisible.value = false
      fetchData() // 刷新表格数据
    }
    else {
      ElMessage.error(`添加失败: ${response.data.message}`)
    }
  }
  catch (error) {
    ElMessage.error('请求失败，请检查网络')
  }
}

// 新增确认删除对话框相关变量
const confirmDeleteVisible = ref(false)
const selectedDevId = ref('')
const selectedDevName = ref('') // 新增设备名称变量
const selectedInstId = ref('') // 新增实例ID变量

// 新增删除确认方法
async function handleDelete() {
  try {
    const response = await axios.post('/api/v1/delDev', {
      devList: [selectedDevId.value],
      instId: selectedInstId.value,
    })
    if (response.status === 200) {
      if (response.data.result === 'success') {
        ElMessage.success('设备删除成功')
        fetchData() // 刷新表格数据
      }
      else {
        ElMessage.error(`删除失败: 设备绑定的实例正在运行，不能删除设备`)
      }
      // ElMessage.success('设备删除成功')
      // fetchData() // 刷新表格数据
    }
  }
  catch (error) {
    ElMessage.error('删除失败，请检查网络')
    console.error('删除设备失败:', error)
  }
  finally {
    confirmDeleteVisible.value = false
  }
}

// 修改删除按钮事件处理，添加设备名称参数
function showDeleteConfirm(dev: DeviceData) { // 参数改为对象形式
  selectedDevId.value = dev.devId
  selectedDevName.value = dev.devName // 新增设备名称赋值
  selectedInstId.value = dev.instId // 新增实例ID赋值
  confirmDeleteVisible.value = true
}

// 新增错误提示变量
const nameError = ref('')

// 新增名称检查函数
function checkDeviceName() {
  const name = formData.value.devName
  if (!name) {
    nameError.value = ''
    return
  }
  nameError.value = tableData.value.some(item => item.devName === name)
    ? '设备名称已存在'
    : ''
}

// 新增启动方法
async function handleStart(dev: DeviceData) {
  try {
    const response = await axios.post('/api/v1/startApp', {
      instid: dev.instId,
    })
    if (response.data.data) {
      ElMessage.success(`启动成功: ${dev.instId}`)
      fetchData() // 刷新表格数据
    }
    else {
      ElMessage.error(`启动失败: ${response.data.details || '未知错误'}`)
    }
  }
  catch (error) {
    ElMessage.error(`启动失败: ${error.response?.data?.details || '网络错误'}`)
  }
}

// 新增停止方法
async function handleStop(dev: DeviceData) {
  try {
    const response = await axios.post('/api/v1/stopApp', {
      instid: dev.instId,
    })
    if (response.data.data) {
      ElMessage.success(`停止成功: ${dev.instId}`)
      fetchData() // 刷新表格数据
    }
    else {
      ElMessage.error(`停止失败: ${response.data.details || '未知错误'}`)
    }
  }
  catch (error) {
    ElMessage.error(`停止失败: ${error.response?.data?.details || '网络错误'}`)
  }
}

// 新增重启方法
async function handleRestart(dev: DeviceData) {
  try {
    const response = await axios.post('/api/v1/restartApp', {
      instid: dev.instId,
    })
    if (response.data.data) {
      ElMessage.success(`重启成功: ${dev.instId}`)
      fetchData() // 刷新表格数据
    }
    else {
      ElMessage.error(`重启失败: ${response.data.details || '未知错误'}`)
    }
  }
  catch (error) {
    ElMessage.error(`重启失败: ${error.response?.data?.details || '网络错误'}`)
  }
}

// 修改onMounted初始化
onMounted(() => {
  fetchData()
  intervalId = setInterval(fetchData, 3000)
  fetchApps() // 新增加载实例列表
})

// 组件卸载时清除定时器
onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
  }
})
</script>

<template>
  <el-table :data="tableData" style="width: 100%">
    <el-table-column prop="devName" label="设备名称" />
    <el-table-column prop="devDesc" label="设备描述" />
    <el-table-column prop="devId" label="设备ID" />
    <el-table-column prop="devType" label="设备类型" />
    <el-table-column prop="instId" label="宿主ID" />
    <!--    <el-table-column prop="instId" label="宿主ID"> -->
    <!--      <template #default="scope"> -->
    <!--        <el-popover placement="bottom" trigger="hover"> -->
    <!--          <template #reference> -->
    <!--            <el-button size="small" type="text"> -->
    <!--              {{ scope.row.instId }} -->
    <!--            </el-button> -->
    <!--          </template> -->
    <!--          <el-menu :default-active="scope.row.isRunning ? 'stop' : 'start'" class="el-menu-demo" mode="vertical"> -->
    <!--            <el-menu-item v-if="!scope.row.isRunning" index="start" @click="handleStart(scope.row)"> -->
    <!--              启动 -->
    <!--            </el-menu-item> -->
    <!--            <el-menu-item v-if="scope.row.isRunning" index="stop" style="color: darkred;" @click="handleStop(scope.row)"> -->
    <!--              停止 -->
    <!--            </el-menu-item> -->
    <!--            <el-menu-item index="restart" @click="handleRestart(scope.row)"> -->
    <!--              重启 -->
    <!--            </el-menu-item> -->
    <!--          </el-menu> -->
    <!--        </el-popover> -->
    <!--      </template> -->
    <!--    </el-table-column> -->
    <el-table-column label="宿主状态">
      <template #default="scope">
        <el-tag
          :type="
            scope.row.isRunning === true
              ? 'success'
              : scope.row.isRunning === false
                ? 'danger'
                : 'info'
          "
        >
          {{ scope.row.isRunning === true ? '运行' : scope.row.isRunning === false ? '停止' : '未知' }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="操作" width="200">
      <template #default="scope">
        <el-button size="small" type="info" @click="emit('data-click', scope.row.devName, scope.row.devDesc, scope.row.devId, scope.row.instId)">
          查看
        </el-button>
        <el-button
          size="small"
          type="danger"
          @click="showDeleteConfirm(scope.row)"
        >
          删除
        </el-button>
      </template>
    </el-table-column>
  </el-table>

  <div style="display: flex; align-items: center; margin-bottom: 10px;">
    <el-button type="primary" @click="addNewDevice">
      新增设备
    </el-button>
  </div>

  <!-- 新增对话框 -->
  <el-dialog v-model="dialogVisible" title="新增设备" width="30%" draggable>
    <el-form :model="formData">
      <el-form-item label="实例名称" required>
        <el-select v-model="formData.instId" placeholder="请选择实例">
          <el-option
            v-for="option in appOptions"
            :key="option.value"
            :label="option.label"
            :value="option.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="实例 ID" required>
        <el-input v-model="formData.instId" disabled />
      </el-form-item>
      <el-form-item label="设备类型" required>
        <el-input v-model="formData.devType" disabled />
      </el-form-item>
      <el-form-item label="设备名称" required :error="nameError">
        <el-input v-model="formData.devName" placeholder="请输入设备名称" @input="checkDeviceName" />
      </el-form-item>
      <el-form-item label="设备描述">
        <el-input v-model="formData.devDesc" placeholder="请输入设备描述" />
      </el-form-item>
      <el-form-item label="设备配置">
        <el-input v-model="formData.config" type="textarea" placeholder="请输入配置信息" />
      </el-form-item>
      <el-form-item v-show="false" label="设备ID">
        <el-input v-model="formData.devId" placeholder="请输入设备ID" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleFormSubmit">
          提交
        </el-button>
        <el-button @click="dialogVisible = false">
          取消
        </el-button>
      </el-form-item>
    </el-form>
  </el-dialog>

  <!-- 新增删除确认对话框 -->
  <el-dialog
    v-model="confirmDeleteVisible"
    title="删除确认"
    width="30%"
    draggable
  >
    <span>确定要删除设备 <strong>{{ selectedDevName }}</strong> (ID: {{ selectedDevId }}) 吗？</span>
    <template #footer>
      <el-button @click="confirmDeleteVisible = false">
        取消
      </el-button>
      <el-button type="danger" @click="handleDelete">
        确认
      </el-button>
    </template>
  </el-dialog>
</template>
