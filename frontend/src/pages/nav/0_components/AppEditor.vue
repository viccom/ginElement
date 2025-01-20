<script lang="ts" setup>
import axios from 'axios'
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

function submitForm() {
  console.log('提交表单:', {
    ...formData.value,
    config: JSON.parse(formData.value.config), // 将字符串解析为 JSON
  })
}

// 启用编辑模式
function enableEditing() {
  isEditing.value = true
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
    isEditing.value = false // 刷新后禁用表单
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
      <el-input v-model="formData.instName" :disabled="!isEditing" />
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
      <el-switch v-model="formData.autoStart" :disabled="!isEditing" />
    </el-form-item>
    <el-form-item label="配置">
      <el-input
        v-model="formData.config"
        type="textarea"
        :rows="10"
        :disabled="!isEditing"
      />
    </el-form-item>

    <el-form-item>
      <el-row :gutter="10">
        <el-col :span="6">
          <el-button type="primary" :disabled="!isEditing" @click="submitForm">
            提交
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button type="warning" :disabled="isEditing" @click="enableEditing">
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
