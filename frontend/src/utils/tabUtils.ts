// src/utils/tabUtils.ts

// 定义标签页类型
export type TabType = 'appList' | 'appEditor' | 'devs' | 'form' | 'data' | 'chart' | string // 支持多种类型

// 定义 Tab 接口
export interface Tab {
  title: string
  name: string
  type: TabType
  config: {
    apiUrl: string
  }
  jsonData: {
    [key: string]: any // 支持动态数据结构
  }
}

// 初始化默认的TAB标签页
export function initDefaultTab(defaultTab: Tab) {
  return [defaultTab]
}

// 保存TAB标签页状态到 localStorage
export function saveTabsToLocalStorage(tabs: Tab[], activeTab: string, storageKey: string = 'tabs') {
  localStorage.setItem(storageKey, JSON.stringify(tabs))
  localStorage.setItem(`${storageKey}_activeTab`, activeTab)
}

// 从 localStorage 恢复TAB标签页状态
export function restoreTabsFromLocalStorage(storageKey: string = 'tabs') {
  const savedTabs = localStorage.getItem(storageKey)
  const savedActiveTab = localStorage.getItem(`${storageKey}_activeTab`)

  return {
    tabs: savedTabs ? JSON.parse(savedTabs) : [],
    activeTab: savedActiveTab || '1',
  }
}

// 添加新标签页
export function addTab(
  tabs: Tab[],
  activeTab: string,
  title: string,
  type: TabType,
  config: { apiUrl: string },
  jsonData: { [key: string]: any },
) {
  // 检查是否已存在相同 title 的标签页
  const existingTab = tabs.find(tab => tab.title === title)

  if (existingTab) {
    // 如果存在，则激活该标签页
    activeTab = existingTab.name
  }
  else {
    // 如果不存在，则创建新标签页
    const newTabName = `${tabs.length + 1}`
    const newTab = {
      title,
      name: newTabName,
      type,
      config,
      jsonData,
    }
    tabs.push(newTab)
    activeTab = newTabName
  }

  return { tabs, activeTab }
}

// 处理标签页关闭
export function handleTabsEdit(tabs: Tab[], activeTab: string, targetName: string, storageKey: string = 'tabs') {
  if (targetName === tabs[0].name) {
    console.log('最左边的标签页不允许关闭')
    return { tabs, activeTab }
  }

  const targetIndex = tabs.findIndex((tab: Tab) => tab.name === targetName)
  if (targetIndex !== -1) {
    tabs.splice(targetIndex, 1)
    if (activeTab === targetName) {
      activeTab = tabs[0]?.name || ''
    }
  }

  // 保存TAB标签页状态
  saveTabsToLocalStorage(tabs, activeTab, storageKey)

  return { tabs, activeTab }
}
