<script lang="ts" setup>
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { defineEmits, onMounted, ref } from 'vue'

const props = defineProps<{
  config: {
    apiUrl: string
  }
}>()

const emit = defineEmits(['closeTab']) // 修改事件名为 camelCase

const appCodes = ref<string[]>([])
const formData = ref({
  instName: '',
  instId: '',
  appCode: '',
  appType: '',
  autoStart: false,
  config: '',
})

// 获取应用编码列表
async function fetchAppCodes() {
  try {
    const response = await axios.get('/api/v1/listAppcode')
    appCodes.value = response.data.appCode
  }
  catch (error) {
    console.error('获取应用编码列表失败:', error)
  }
}

// 获取默认配置
async function fetchDefaultConfig(appCode: string) {
  try {
    const response = await axios.post(
      '/api/v1/getAppDefault',
      { appCode },
      { headers: { 'Content-Type': 'application/json' } },
    )
    const defaultConfig = response.data.data.appConfig
    formData.value = {
      ...formData.value,
      appType: defaultConfig.appType,
      autoStart: defaultConfig.autoStart,
      config: JSON.stringify(defaultConfig.config, null, 2),
    }
  }
  catch (error) {
    console.error('获取默认配置失败:', error)
  }
}

// 提交表单
async function submitForm() {
  let hasError = false

  if (!formData.value.appCode) {
    ElMessage.error('应用编码不能为空')
    hasError = true
  }

  if (!formData.value.appType) {
    ElMessage.error('应用类型不能为空')
    hasError = true
  }

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

  // 新增名称冲突检查
  if (formData.value.instName && appListNames.value.includes(formData.value.instName)) {
    ElMessage.error('应用名称已存在')
    return
  }

  try {
    const postData = {
      ...formData.value,
      config: JSON.parse(formData.value.config),
    }
    const response = await axios.post('/api/v1/newApp', postData, {
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

// 新增应用名称列表和错误提示
const appListNames = ref<string[]>([])
const nameError = ref('')

// 新增获取应用列表名称的方法
async function fetchAppList() {
  try {
    const response = await axios.get('/api/v1/listApps', {
      headers: {
        'Content-Type': 'application/json',
      },
    })
    console.log('获取应用列表成功:', response.data)
    if (
      response?.data
      && typeof response.data === 'object'
      && response.data.data !== undefined
    ) {
      appListNames.value = Object.values(response.data.data).map(item => item.instName)
    }
    else {
      throw new Error('响应数据格式不符合预期')
    }
  }
  catch (error) {
    // 使用可选链避免error.response未定义时报错
    console.error('获取应用列表失败:', error, '响应内容:', error.response?.data)
    appListNames.value = []
  }
}

// 新增名称冲突检查方法
function checkNameConflict() {
  const currentName = formData.value.instName
  if (!currentName) {
    nameError.value = ''
    return
  }
  nameError.value = appListNames.value.includes(currentName) ? '名称已存在' : ''
}

onMounted(() => {
  fetchAppCodes()
  fetchAppList() // 新增加载应用列表
})
</script>

<template>
  <el-form :model="formData" label-width="120px">
    <el-form-item label="应用编码">
      <el-select v-model="formData.appCode" placeholder="请选择应用" @change="fetchDefaultConfig">
        <el-option v-for="code in appCodes" :key="code" :label="code" :value="code" />
      </el-select>
    </el-form-item>

    <!-- 实例名称 -->
    <el-form-item
      label="实例名称"
      :error="nameError"
    >
      <el-input
        v-model="formData.instName"
        @input="checkNameConflict"
      />
    </el-form-item>

    <el-form-item label="应用类型">
      <el-input v-model="formData.appType" :disabled="true" />
    </el-form-item>
    <el-form-item label="自启动">
      <el-switch v-model="formData.autoStart" />
    </el-form-item>

    <!-- 配置 -->
    <el-form-item
      label="配置"
      :error="!formData.config && '配置不能为空'"
    >
      <el-input
        v-model="formData.config"
        type="textarea"
        :rows="10"
      />
    </el-form-item>

    <!-- 提交按钮 -->
    <el-form-item>
      <el-button type="primary" @click="submitForm">
        提交
      </el-button>
    </el-form-item>
  </el-form>
</template>
