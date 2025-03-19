<script lang="ts" setup>
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { onMounted, ref } from 'vue'

const props = defineProps<{
  config: {
    apiUrl: string
  }
  jsonData: {
    instName: string
    instId: string
    isRunning: boolean
  }
}>()

// 定义事件
const emit = defineEmits(['close-tab'])

const formData = ref({
  instName: props.jsonData.instName,
  instId: props.jsonData.instId,
  appCode: '',
  appType: '',
  autoStart: '',
  config: '', // 存储格式化后的 JSON 字符串
})

const isEditing = ref(false) // 控制表单是否可编辑
const isEditinga = ref(false)

// 格式化 JSON 数据
function formatJson(json: any) {
  try {
    return JSON.stringify(json, null, 2) // 缩进 2 个空格
  }
  catch (error) {
    console.error('格式化 JSON 失败:', error)
    return ''
  }
}

onMounted(async () => {
  const requestBody = {
    instid: props.jsonData.instId,
  }
  try {
    const response = await axios.post(props.config.apiUrl, requestBody)
    formData.value = {
      ...response.data.data,
      config: formatJson(response.data.data.config), // 格式化 JSON
    }
  }
  catch (error) {
    console.error('加载表单数据失败:', error)
  }
})

// 提交表单
async function submitForm() {
  let hasError = false

  if (!formData.value.instName) {
    ElMessage.error('实例名称不能为空')
    hasError = true
  }

  if (!formData.value.config) {
    ElMessage.error('配置不能为空')
    hasError = true
  }

  if (hasError) {
    return
  }

  try {
    const postData = {
      ...formData.value,
      config: JSON.parse(formData.value.config)
    }
    const response = await axios.post('/api/v1/modApp', postData, {
      headers: { 'Content-Type': 'application/json' },
    })

    if (response.status === 200) {
      ElMessage.success('提交成功')
      // 提交成功后通知父组件关闭当前标签页
      const apptabs_activeTab = localStorage.getItem('apptabs_activeTab')
      console.log('触发关闭事件，参数：', apptabs_activeTab)
      emit('closeTab', apptabs_activeTab)
    }
    else {
      ElMessage.error(`提交失败: ${response.data.details || '未知错误'}`)
    }
  }
  catch (error) {
    ElMessage.error(`提交失败: ${error.response?.data?.details || '网络错误'}`)
  }
}

// 启用编辑模式
function enableEditing() {
  isEditinga.value = true
}

// 刷新数据
async function refreshData() {
  const requestBody = {
    instid: props.jsonData.instId,
  }
  try {
    const response = await axios.post(props.config.apiUrl, requestBody)
    formData.value = {
      ...response.data.data,
      config: formatJson(response.data.data.config), // 格式化 JSON
    }
    // 刷新后禁用表单
    // isEditing.value = false
    // isEditinga.value = false
  }
  catch (error) {
    console.error('刷新数据失败:', error)
  }
}

// 关闭当前标签页
function closeTab() {
  const savedActiveTab = localStorage.getItem(`apptabs_activeTab`)
  emit('close-tab', savedActiveTab) // 通知父组件关闭当前标签页
}
</script>

<template>
  <el-form :model="formData" label-width="120px">
    <el-form-item label="实例名称">
      <el-input v-model="formData.instName" :disabled="!isEditinga" />
    </el-form-item>
    <el-form-item label="实例ID">
      <el-input v-model="formData.instId" :disabled="!isEditing" />
    </el-form-item>
    <el-form-item label="应用编码">
      <el-input v-model="formData.appCode" :disabled="!isEditing" />
    </el-form-item>
    <el-form-item label="应用类型">
      <el-input v-model="formData.appType" :disabled="!isEditing" />
    </el-form-item>
    <el-form-item label="自启动">
      <el-switch v-model="formData.autoStart" :disabled="!isEditinga" />
    </el-form-item>
    <el-form-item label="配置">
      <el-input
        v-model="formData.config"
        type="textarea"
        :rows="10"
        :disabled="!isEditinga"
      />
    </el-form-item>

    <el-form-item>
      <el-row :gutter="10">
        <el-col :span="6">
          <el-button type="primary" :disabled="!isEditinga" @click="submitForm">
            提交
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button type="warning" :disabled="isEditinga" @click="enableEditing">
            编辑
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button type="info" @click="refreshData">
            刷新
          </el-button>
        </el-col>
        <el-col :span="6" style="text-align: right;">
          <el-button type="danger" @click="closeTab">
            返回
          </el-button>
        </el-col>
      </el-row>
    </el-form-item>
  </el-form>
</template>
