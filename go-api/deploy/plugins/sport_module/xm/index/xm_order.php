<?php
$mod='blank';
$title='XM运动';
require_once('head.php');
?>
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no" />
    <title>XM运动 · 跑步订单管理</title>
    <!-- 修复CDN加载顺序：先Vue再Element Plus -->
    <script src="https://unpkg.com/vue@3/dist/vue.global.prod.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/element-plus/dist/index.css" />
    <script src="https://unpkg.com/element-plus/dist/index.full.min.js"></script>
    <script src="https://unpkg.com/element-plus/dist/locale/zh-cn.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            background: #f5f5f5;
            padding: 10px;
            font-family: "Helvetica Neue", Arial, sans-serif;
        }
        .table-container {
            max-width: 100%;
            margin: 0 auto;
            background: #fff;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 12px rgba(0,0,0,0.08);
        }
        /* 操作按钮容器：紧凑排列+居中对齐 */
        .action-buttons {
            display: flex;
            flex-wrap: wrap;
            justify-content: center;
            align-items: center;
            gap: 4px; /* 缩小按钮间距 */
            width: 100%;
            padding: 3px 0;
        }
        /* 缩小按钮尺寸并固定样式 */
        .action-buttons .el-button--small {
            padding: 2px 6px !important; /* 极小内边距 */
            font-size: 11px !important; /* 缩小文字 */
            min-width: 50px !important; /* 固定最小宽度 */
            height: 24px !important; /* 固定高度 */
            text-align: center;
            white-space: nowrap;
        }
        /* 优化固定列样式（关键：避免滚动时偏移） */
        .el-table__fixed-right {
            box-shadow: -2px 0 10px rgba(0,0,0,0.05); /* 右侧阴影区分固定列 */
            border-left: 1px solid #eee;
        }
        /* 表格滚动时固定列表头对齐 */
        .el-table__fixed-header-wrapper .el-table__header {
            border-right: none !important;
        }
        /* 日志类型样式 */
        .log-type-success { 
            color: #67c23a; 
            background: #f0f9eb;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 12px;
        }
        .log-type-info { 
            color: #409eff; 
            background: #ecf5ff;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 12px;
        }
        .log-type-warning { 
            color: #e6a23c; 
            background: #faf6ed;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 12px;
        }
        .log-type-error { 
            color: #f56c6c; 
            background: #fef0f0;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 12px;
        }
        /* 表格单元格优化 */
        .el-table__cell {
            white-space: normal !important;
            word-break: break-all;
            line-height: 1.5;
            padding: 8px 5px !important;
        }
        /* 表单间距优化 */
        .demo-form-inline {
            margin-bottom: 15px;
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            align-items: center;
        }
        /* 分页样式优化 */
        .el-pagination {
            margin-top: 15px;
            text-align: right;
            font-size: 12px;
        }
        /* 手机端适配 */
        @media (max-width: 768px) {
            .action-buttons {
                gap: 2px;
            }
            .action-buttons .el-button--small {
                padding: 1px 4px !important;
                font-size: 10px !important;
                min-width: 45px !important;
                height: 22px !important;
            }
            /* 固定列宽度自适应 */
            .el-table__fixed-right {
                width: 180px !important;
            }
        }
    </style>
</head>
<body>
<div id="app"></div>

<script>
    const { createApp, reactive, ref, onMounted } = Vue;
    const { ElMessage, ElMessageBox } = ElementPlus;
    const zhCn = ElementPlusLocaleZhCn;

    const App = {
        template: `
          <div class="table-container">
            <el-card>
              <template #header>
                <span style="font-size: 18px; font-weight: 600; color: #409EFF;">跑步订单列表</span>
              </template>
              
              <!-- 搜索表单 -->
              <el-form :inline="true" :model="query" class="demo-form-inline" @submit.prevent>
                <el-form-item label="账号" label-width="60px">
                  <el-input v-model="query.account" placeholder="请输入账号" clearable @keyup.enter="fetchData" />
                </el-form-item>
                <el-form-item label="学校" label-width="60px">
                  <el-input v-model="query.school" placeholder="请输入学校名称" clearable @keyup.enter="fetchData" />
                </el-form-item>
                <el-form-item label="状态" label-width="60px">
                  <el-input v-model="query.status" placeholder="请输入订单状态" clearable @keyup.enter="fetchData" />
                </el-form-item>
                <el-form-item label="项目" label-width="60px">
                  <el-select v-model="query.project" placeholder="请选择项目" clearable>
                    <el-option v-for="item in projectOptions" :key="item.value" :label="item.label" :value="item.value" />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="fetchData" :loading="tableLoading">搜索</el-button>
                  <el-button @click="resetQuery">重置</el-button>
                </el-form-item>
              </el-form>

              <!-- 订单表格 -->
              <el-table 
                  :data="tableData" 
                  border 
                  style="width: 100%;" 
                  :header-cell-style="{ background: '#f5f7fa', fontSize: '14px' }"
                  v-loading="tableLoading"
                  element-loading-text="加载中..."
                  empty-text="暂无订单数据"
              >
                <el-table-column prop="id" label="订单ID" width="80" align="center" />
                <el-table-column prop="updated_at" label="更新时间" align="center" width="160">
                  <template #default="{ row }">{{ formatDate(row.updated_at) }}</template>
                </el-table-column>
                <el-table-column label="项目名称" align="center" width="150">
                  <template #default="{ row }">{{ getProjectLabel(row.project_id) }}</template>
                </el-table-column>
                <el-table-column prop="type" label="类型" align="center" width="120" />
                <el-table-column prop="school" label="学校" align="center" width="150" />
                <el-table-column prop="account" label="账号" align="center" width="150" />
                <el-table-column prop="password" label="密码" align="center" width="150" />
                <el-table-column label="跑步周期" align="center" width="200">
                  <template #default="{ row }">{{ formatWeekdays(row.run_date) }}</template>
                </el-table-column>
                <el-table-column prop="total_km" label="下单次数" align="center" width="100" />
                <el-table-column prop="run_km" label="已跑次数" align="center" width="100" />
                <el-table-column prop="start_day" label="开始日期" align="center" width="120" />
                <el-table-column prop="start_time" label="开始时间" align="center" width="120" />
                <el-table-column prop="end_time" label="结束时间" align="center" width="120" />
                <el-table-column prop="deduction" label="费用" align="center" width="80" />
                <el-table-column prop="status_name" label="状态" align="center" width="100">
                  <template #default="{ row }">
                    <span :class="getStatusClass(row.status_name)">
                      {{ row.status_name }}
                    </span>
                  </template>
                </el-table-column>
                <!-- 操作列：固定在右侧，宽度适配按钮 -->
                <el-table-column label="操作" align="center" width="80" fixed="right">
                  <template #default="{ row }">
                    <div class="action-buttons">
                      <el-button size="small" type="success" icon="el-icon-document" @click="handleViewLogs(row.id)">日志</el-button>
                    
                      <el-button size="small" type="danger" icon="el-icon-circle-close" @click="handleRefund(row.id)">退款</el-button>
                      
                    </div>
                  </template>
                </el-table-column>
              </el-table>

              <!-- 分页 -->
              <el-pagination
                  v-if="total > 0"
                  background
                  layout="total, prev, pager, next, jumper"
                  :total="total"
                  :current-page="query.page"
                  :page-size="query.page_size"
                  @current-change="handlePageChange"
              />
            </el-card>

            <!-- 日志查看弹窗 -->
            <el-dialog
                v-model="logDialog.visible"
                :title="'订单日志 (ID: ' + logDialog.orderId + ')'"
                width="80%"
                :before-close="handleCloseLogDialog"
                top="20px"
            >
              <el-table 
                  :data="logDialog.data" 
                  border 
                  style="width: 100%"
                  v-loading="logDialog.loading"
                  element-loading-text="加载日志中..."
                  empty-text="暂无日志记录"
                  :header-cell-style="{ background: '#f5f7fa', fontSize: '14px' }"
              >
                <el-table-column prop="id" label="日志ID" width="80" align="center" />
                <el-table-column label="类型" width="120" align="center">
                  <template #default="{ row }">
                    <span :class="getLogTypeClass(row.log_type)">
                      {{ formatLogType(row.log_type) }}
                    </span>
                  </template>
                </el-table-column>
                <el-table-column prop="message" label="消息内容" />
                <el-table-column prop="updated_at" label="时间" width="180" align="center">
                  <template #default="{ row }">{{ formatDate(row.updated_at) }}</template>
                </el-table-column>
              </el-table>

              <el-pagination
                  v-if="logDialog.total > 0"
                  background
                  layout="total, prev, pager, next, jumper"
                  :total="logDialog.total"
                  :current-page="logDialog.page"
                  :page-size="logDialog.page_size"
                  @current-change="handleLogPageChange"
                  style="margin-top: 20px; text-align: right;"
              />

              <template #footer>
                <span class="dialog-footer">
                  <el-button @click="logDialog.visible = false">关闭</el-button>
                  <el-button type="primary" @click="fetchOrderLogs(logDialog.orderId, logDialog.page)">刷新</el-button>
                </span>
              </template>
            </el-dialog>
          </div>
        `,
        setup() {
            // 订单表格数据
            const tableData = ref([]);
            const total = ref(0);
            const tableLoading = ref(false);
            const projectOptions = ref([]);

            // 查询参数
            const query = reactive({
                account: "",
                school: "",
                status: "",
                project: "",
                page: 1,
                page_size: 10,
            });

            // 日志弹窗数据
            const logDialog = reactive({
                visible: false,
                loading: false,
                orderId: null,
                data: [],
                total: 0,
                page: 1,
                page_size: 10,
            });

            // 星期映射
            const weekMap = {
                1: '星期一',
                2: '星期二',
                3: '星期三',
                4: '星期四',
                5: '星期五',
                6: '星期六',
                7: '星期日'
            };

            // 日志类型映射
            const logTypeMap = {
                'success': '成功',
                'info': '信息',
                'warning': '警告',
                'error': '错误',
                1: '成功',
                2: '信息',
                3: '警告',
                4: '错误'
            };

            // 订单状态样式映射
            const statusClassMap = {
                '已完成': 'log-type-success',
                '运行中': 'log-type-info',
                '待处理': 'log-type-warning',
                '已取消': 'log-type-error',
                '退款中': 'log-type-warning',
                '已退款': 'log-type-error'
            };

            // 获取项目名称
            function getProjectLabel(projectId) {
                if (!projectId) return "未知项目";
                const project = projectOptions.value.find(item => item.value == projectId);
                return project ? project.label : "未知项目";
            }

            // 格式化星期
            function formatWeekdays(days) {
                if (!days) return "无";
                const dayArr = Array.isArray(days) ? days : JSON.parse(days || '[]');
                return dayArr.map(day => weekMap[day] || day).join("、") || "无";
            }

            // 格式化日期
            function formatDate(dateStr) {
                if (!dateStr) return "-";
                if (typeof dateStr === 'number') {
                    dateStr = dateStr.toString().length === 10 ? dateStr * 1000 : dateStr;
                }
                const date = new Date(dateStr);
                return date.toLocaleString('zh-CN', {
                    year: 'numeric',
                    month: '2-digit',
                    day: '2-digit',
                    hour: '2-digit',
                    minute: '2-digit',
                    second: '2-digit'
                });
            }

            // 格式化日志类型
            function formatLogType(logType) {
                return logTypeMap[logType] || logType || '未知';
            }

            // 获取日志类型样式
            function getLogTypeClass(logType) {
                const type = formatLogType(logType);
                const typeMap = {
                    '成功': 'log-type-success',
                    '信息': 'log-type-info',
                    '警告': 'log-type-warning',
                    '错误': 'log-type-error'
                };
                return typeMap[type] || 'log-type-info';
            }

            // 获取订单状态样式
            function getStatusClass(status) {
                return statusClassMap[status] || 'log-type-info';
            }

            // 获取项目列表
            async function fetchProjectList() {
                try {
                    const res = await fetch("/xm_apis.php?act=getProjects", { credentials: "include" });
                    const json = await res.json();
                    if (json.code === 1 || json.code === 200) {
                        projectOptions.value = json.data?.map(p => ({
                            label: p.name,
                            value: p.id,
                        })) || [];
                    } else {
                        ElMessage.error("获取项目列表失败：" + (json.msg || '未知错误'));
                    }
                } catch (e) {
                    console.error("获取项目失败", e);
                    ElMessage.error("获取项目列表失败，请刷新页面重试");
                }
            }

            // 加载订单数据（防抖）
            let fetchTimeout = null;
            async function fetchData() {
                if (fetchTimeout) clearTimeout(fetchTimeout);
                fetchTimeout = setTimeout(async () => {
                    if (projectOptions.value.length === 0) {
                        ElMessage.warning("项目列表加载中，请稍后再试");
                        return;
                    }

                    tableLoading.value = true;
                    const params = new URLSearchParams();
                    for (const key in query) {
                        if (query[key] !== "" && query[key] !== undefined && query[key] !== null) {
                            params.append(key, query[key]);
                        }
                    }

                    try {
                        const res = await fetch("/xm_apis.php?act=get_orders&" + params.toString(), {
                            credentials: "include",
                            headers: { 'Content-Type': 'application/json' }
                        });
                        const json = await res.json();
                        if (json.code === 200) {
                            tableData.value = json.data || [];
                            total.value = json.total || 0;
                        } else {
                            ElMessage.error(json.msg || "加载订单失败");
                            tableData.value = [];
                            total.value = 0;
                        }
                    } catch (e) {
                        console.error("加载订单失败", e);
                        ElMessage.error("加载订单失败，请检查网络连接");
                        tableData.value = [];
                        total.value = 0;
                    } finally {
                        tableLoading.value = false;
                    }
                }, 300);
            }

            // 获取订单日志
            async function fetchOrderLogs(orderId, page = 1) {
                logDialog.loading = true;
                try {
                    const params = new URLSearchParams({
                        act: 'get_order_logs',
                        order_id: orderId,
                        page: page,
                        page_size: logDialog.page_size
                    });

                    const res = await fetch("/xm_apis.php?" + params.toString(), {
                        credentials: "include"
                    });
                    const json = await res.json();
                    
                    if (json.code === 200) {
                        logDialog.data = json.data || [];
                        logDialog.total = json.total || 0;
                        logDialog.page = json.page || page;
                    } else {
                        ElMessage.error(json.msg || "获取日志失败");
                        logDialog.data = [];
                        logDialog.total = 0;
                    }
                } catch (e) {
                    console.error("获取日志失败", e);
                    ElMessage.error("获取日志失败，请检查网络连接");
                } finally {
                    logDialog.loading = false;
                }
            }

            // 查看日志
            function handleViewLogs(orderId) {
                logDialog.orderId = orderId;
                logDialog.visible = true;
                logDialog.page = 1;
                fetchOrderLogs(orderId, 1);
            }

            // 关闭日志弹窗
            function handleCloseLogDialog() {
                logDialog.visible = false;
                logDialog.data = [];
                logDialog.total = 0;
                logDialog.page = 1;
            }

            // 日志分页变更
            function handleLogPageChange(val) {
                logDialog.page = val;
                fetchOrderLogs(logDialog.orderId, val);
            }

            // 重置查询
            function resetQuery() {
                query.account = "";
                query.school = "";
                query.status = "";
                query.project = "";
                query.page = 1;
                fetchData();
            }

            // 订单分页变更
            function handlePageChange(val) {
                query.page = val;
                fetchData();
            }

            // 退款操作
            async function handleRefund(orderId) {
                try {
                    await ElMessageBox.confirm(
                        '确定要申请退款吗？',
                        '温馨提示',
                        {
                            confirmButtonText: '确定',
                            cancelButtonText: '取消',
                            type: 'warning',
                            center: true
                        }
                    );

                    const res = await fetch(`/xm_apis.php?act=refund_order&order_id=${encodeURIComponent(orderId)}`, {
                        credentials: 'include'
                    });
                    const json = await res.json();
                    if (json.code === 200) {
                        ElMessage.success(json.msg || '退款申请成功');
                        fetchData();
                    } else {
                        ElMessage.error(json.msg || '退款申请失败');
                    }
                } catch (e) {
                    if (e.name !== 'Error') console.error(e);
                }
            }

            // 删除订单
            async function handleDelete(orderId) {
                try {
                    const res = await fetch(`/xm_apis.php?act=delete_order&order_id=${encodeURIComponent(orderId)}`, {
                        method: 'GET',
                        credentials: 'include'
                    });
                    const json = await res.json();
                    if (json.code === 200) {
                        ElMessage.success(json.msg || '删除成功');
                        fetchData();
                    } else {
                        ElMessage.error(json.msg || '删除失败');
                    }
                } catch (e) {
                    console.error(e);
                    ElMessage.error('删除请求失败');
                }
            }

            // 同步订单
            async function handleSync(orderId) {
                try {
                    await ElMessageBox.confirm(
                        '确定要同步该订单吗？同步可能会覆盖远程数据',
                        '温馨提示',
                        {
                            confirmButtonText: '确定',
                            cancelButtonText: '取消',
                            type: 'info',
                            center: true
                        }
                    );

                    const res = await fetch(`/xm_apis.php?act=sync_order&order_id=${encodeURIComponent(orderId)}`, {
                        credentials: 'include'
                    });
                    const json = await res.json();
                    if (json.code === 200) {
                        ElMessage.success(json.msg || '同步成功');
                        fetchData();
                    } else {
                        ElMessage.error(json.msg || '同步失败');
                    }
                } catch (e) {
                    if (e.name !== 'Error') console.error(e);
                }
            }

            // 页面初始化
            onMounted(async () => {
                await fetchProjectList();
                fetchData();
            });

            return {
                tableData,
                total,
                tableLoading,
                query,
                projectOptions,
                logDialog,
                getProjectLabel,
                formatWeekdays,
                formatDate,
                formatLogType,
                getLogTypeClass,
                getStatusClass,
                fetchData,
                fetchOrderLogs,
                resetQuery,
                handlePageChange,
                handleRefund,
                handleDelete,
                handleSync,
                handleViewLogs,
                handleCloseLogDialog,
                handleLogPageChange,
            };
        },
    };

    const app = createApp(App);
    app.use(ElementPlus, { locale: zhCn });
    app.mount('#app');
</script>
</body>
</html>