<template>
  <div class="card-container">
    <el-tabs type="border-card" @tab-click="handleTabClick">
      <el-tab-pane label="标签页 1" name="tab1">
        <el-table :data="tableData">
          <el-table-column prop="name" label="姓名" />
          <el-table-column prop="age" label="年龄" />
          <el-table-column prop="address" label="地址" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="标签页 2" name="tab2">
        <!-- 标签页 2 的内容 -->
      </el-tab-pane>
      <!-- 可以添加更多的标签页 -->
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const tableData = ref([
  { name: '张三', age: 30, address: '北京市' },
  { name: '李四', age: 25, address: '上海市' },
  { name: '王五', age: 35, address: '广州市' },
]);

// 修复：指定 tabName 的类型为 string
const handleTabClick = (tabName: string) => {
  console.log('Clicked tab:', tabName);
  // 可以在这里处理标签页点击事件
};
</script>

<style scoped>
.card-container {
  border: 1px solid #ccc;
  border-radius: 5px;
  padding: 10px;
}
</style>
