<?php
$title = '闪电闪动校园';
?>
<link rel="stylesheet" href="/flash/static/element-plus.css" />
<link rel="stylesheet" href="/flash/static/oneui.min.css" />
<style>
  @media (max-width: 768px) {
    /* 手机端弹窗全屏样式 */
    #addDialog {
      padding: 0 !important;
      margin: 0 !important;
      width: 100% !important;
      height: 100% !important;
      top: 0 !important;
      left: 0 !important;
      transform: none !important;
    }
    
    #addDialog .el-dialog {
      width: 100vw !important;
      height: 100vh !important;
      max-width: 100% !important;
      max-height: 100% !important;
      border-radius: 0 !important;
      padding: 0 !important;
      margin: 0 !important;
    }
    
    #addDialog .el-dialog__header {
      padding: 15px;
      border-bottom: 1px solid #e4e7ed;
    }
    
    #addDialog .el-dialog__body {
      height: calc(100vh - 110px) !important; /* 减去头部和底部高度 */
      overflow: auto !important;
      padding: 15px;
    }
    
    #addDialog .el-dialog__footer {
      padding: 10px 15px;
      border-top: 1px solid #e4e7ed;
      position: fixed;
      bottom: 0;
      left: 0;
      right: 0;
      background: #fff;
      z-index: 10;
    }
    
    /* 适配手机端表单样式 */
    #addDialog .add-form {
      width: 100%;
    }
    
    #addDialog .el-form-item {
      margin-bottom: 15px;
    }
    
    #addDialog .plan-time-container {
      margin-bottom: 15px;
    }
  }

  .plan-time-container {
    border: 1px solid #e4e7ed;
    border-radius: 4px;
    margin-bottom: 10px;
  }

  .plan-time-container .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px;
    border-bottom: 1px solid #e4e7ed;
  }

  .plan-time-container .header .title {
    font-weight: bold;
  }

  .plan-time-container .footer {
    margin-bottom: 10px;
    display: flex;
    flex-wrap: wrap;
  }

  .plan-time-container .footer .task-item {
    background-color: rgb(236, 245, 255);
    border-color: rgb(217, 236, 255);
    color: #409eff;
    display: inline-flex;
    justify-content: center;
    align-items: center;
    vertical-align: middle;
    height: 24px;
    padding: 0 9px;
    font-size: 12px;
    line-height: 1;
    border-width: 1px;
    border-style: solid;
    border-radius: 4px;
    box-sizing: border-box;
    white-space: nowrap;
    margin: 5px;
  }
</style>

<div id="content" role="main">
  <div class="app-content-body">
    <div class="wrapper-md control" id="app">

      <!-- 列表 - 开始 -->
      <el-card class="box-card mb-4">
        <div class="card-body">
          <el-alert
            title="闪动客户短信登录和人脸管理页面地址：https://dsabc.sbs/#/auth/sdxy 和 https://sdnb.one/#/auth/sdxy"
            type="success">
          </el-alert>
          <div class="row mt-2">
            <div class="col-md-6 mb-3">
              <el-button type="primary" size="default" plain @click="addDialogVisible = true">添加订单</el-button>
              <el-button type="success" onclick="javascript:void(0);window.open('https://dsabc.sbs/#/auth/sdxy');" size="default" plain>自助接短信验证链接</el-button>
              <el-button type="success" onclick="javascript:void(0);window.open('https://sdnb.one/#/auth/sdxy');" size="default" plain>自助上传人脸链接</el-button>
              <el-button type="warning" size="default">价格：{{ price }} / 次</el-button>
            </div>
            <div class="col-md-6 mb-3">
              <el-input size="default" placeholder="关键词" v-model="search.keywords"
                class="input-with-select">
                <template #prepend>
                  <el-select v-model="search.type" size="default" style="width: 100px" placeholder="条件"
                    slot="prepend">
                    <el-option label="订单ID" value="1"></el-option>
                    <el-option label="下单账号" value="2"></el-option>
                    <el-option label="下单密码" value="3"></el-option>
                    <?php if ($userrow['uid'] == 1) { ?>
                      <el-option label="用户UID" value="4"></el-option>
                    <?php } ?>
                  </el-select>
                </template>
                <template #append>
                  <el-button size="default" @click="loadData">
                    <el-icon style="vertical-align: middle">
                      <Search />
                    </el-icon>
                    <span style="vertical-align: middle"> 搜索 </span>
                  </el-button>
                </template>
              </el-input>
            </div>
          </div>
          <el-table :data="tableData" style="width: 100%" v-loading="tableLoading" highlight-current-row border>
            <el-table-column align="center" prop="id" label="ID" width="80"></el-table-column>
            <?php if ($userrow['uid'] == 1) { ?>
              <el-table-column align="center" prop="uid" label="UID" width="80"></el-table-column>
            <?php } ?>
            <el-table-column align="center" label="学校" width="200">
              <template v-slot:default="scope">
                <div>{{ scope.row.school }}</div>
              </template>
            </el-table-column>
            <el-table-column align="center" label="账号信息" width="150">
              <template v-slot:default="scope">
                <div>{{ scope.row.user }}</div>
                <div>{{ scope.row.pass }}</div>
              </template>
            </el-table-column>
            <el-table-column align="center" label="跑步类型" width="100">
              <template v-slot:default="scope">
                <el-tag type="primary" v-if="scope.row.run_type == 'SUN'">阳光跑</el-tag>
                <el-tag type="success" v-else>自由跑</el-tag>
              </template>
            </el-table-column>
            <el-table-column align="center" label="公里数" width="100">
              <template v-slot:default="scope">{{ scope.row.distance }}</template>
            </el-table-column>
            <el-table-column align="center" label="次数" width="100">
              <template v-slot:default="scope">{{ scope.row.num }}</template>
            </el-table-column>
            <el-table-column align="center" prop="run_rule" label="跑步计划"></el-table-column>
            <el-table-column label="跑步状态" width="90" align="center">
              <template v-slot:default="{ row }">
                <el-switch
                  v-model="row.pause"
                  @change="changePause(row)"
                  active-value="1"
                  inactive-value="0">
                </el-switch>
              </template>
            </el-table-column>
            <el-table-column align="center" label="订单状态" width="100">
              <template v-slot:default="scope">
                <el-tag type="primary" v-if="scope.row.status == 1">进行中</el-tag>
                <el-tag type="success" v-else-if="scope.row.status == 2">已完成</el-tag>
                <el-tag type="danger" v-else-if="scope.row.status == 3">异常</el-tag>
                <el-tag type="warning" v-else-if="scope.row.status == 4">需短信</el-tag>
                <el-tag type="info" v-else-if="scope.row.status == 5">已退款</el-tag>
              </template>
            </el-table-column>
            <el-table-column align="center" prop="created_at" label="下单时间" width="120"></el-table-column>
            <el-table-column align="center" label="用户操作" width="120">
              <template v-slot:default="scope">
                <el-dropdown @command="handleMenu" trigger="click">
                  <el-button type="primary">
                    操作<el-icon class="el-icon--right"><arrow-down /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <!-- 订单状态-1:进行中-2:完成-3:异常-4:需短信-5:已退款 -->
                      <el-dropdown-item :command="{ type: 'log', item: scope.row }">查看日志</el-dropdown-item>
                      <el-dropdown-item v-if="scope.row.status == 3 || scope.row.status == 4" :command="{ type: 'delay', item: scope.row }">延期跑步</el-dropdown-item>
                      <el-dropdown-item v-if="scope.row.status != 2 && scope.row.status != 5" :command="{ type: 'refund', item: scope.row }">退款订单</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <div class="d-none d-md-flex" style="display: flex;justify-content: center;margin-top: 20px;">
            <el-pagination background @size-change="handleSizeChange" @current-change="handleCurrentChange" :current-page.sync="pagination.page"
              :page-sizes="[20, 50, 100, 200, 300]" :page-size="pagination.limit"
              layout="total, sizes, prev, pager, next" :total="pagination.total">
            </el-pagination>
          </div>

          <div class="d-md-flex d-md-none" style="display: flex;justify-content: center;margin-top: 20px;">
            <el-pagination background pager-count="3" @size-change="handleSizeChange" @current-change="handleCurrentChange" :current-page.sync="pagination.page" :page-sizes="[20, 50, 100, 200, 300]" :page-size="pagination.limit" layout="prev, pager, next" :total="pagination.total">
            </el-pagination>
          </div>
        </div>
      </el-card>
      <!-- 列表 - 结束 -->

      <!-- 添加弹窗 - 开始 -->
      <el-dialog title="添加订单" id="addDialog" v-model="addDialogVisible" top="30px">
        <el-form label-width="80px" class="add-form">
          <div style="display: flex;justify-content: center;margin-bottom: 20px; width: 100%;">
            <el-radio-group v-model="userInfoForm.loginType">
              <el-radio-button value="password">账号密码登录</el-radio-button>
              <el-radio-button value="code">验证码登录</el-radio-button>
            </el-radio-group>
          </div>
          <el-form-item label="手机号">
            <el-input v-model="userInfoForm.phone" size="default" placeholder="请输入手机号">
              <template #append v-if="userInfoForm.loginType == 'code'">
                <el-button type="primary" size="default" @click="sendCode">发送验证码</el-button>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item label="验证码" v-if="userInfoForm.loginType == 'code'">
            <el-input v-model="userInfoForm.code" size="default" placeholder="请输入验证码">
              <template #append>
                <el-button type="primary" size="default" @click="getUserInfoByCode">点击查询</el-button>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item label="密码" v-if="userInfoForm.loginType == 'password'">
            <el-input v-model="userInfoForm.password" size="default" placeholder="老单可不用密码">
              <template #append>
                <el-button type="primary" size="default" @click="getUserInfoByPassword">点击查询</el-button>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item label="跑步类型">
            <el-radio-group v-model="addForm.run_type">
              <el-radio value="SUN">阳光跑</el-radio>
              <el-radio value="FREE">自由跑</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="跑步计划">
            <div class="d-flex align-items-center" style="width: 100%;">
              <el-select v-model="addForm.run_rule_id" size="default" placeholder="请选择跑步计划">
                <el-option
                  v-for="item in runRuleList"
                  :key="item.run_rule_id"
                  :label="item.label"
                  :value="item.run_rule_id">
                </el-option>
              </el-select>
              <el-popconfirm
                title="确认当前跑步计划有误？请勿频繁重复提交！"
                placement="top-end"
                @confirm="updateRunRule">
                <template #reference>
                  <el-button style="margin-left: 10px;" type="success" v-if="userInfo?.student?.student_id">更新计划</el-button>
                </template>
              </el-popconfirm>
            </div>
          </el-form-item>
          <el-form-item label="跑步区域">
            <el-select v-model="addForm.zone_id" size="default" placeholder="请选择跑步区域">
              <el-option
                v-for="item in zoneList"
                :key="item.zone_id"
                :label="`${item.sub_school.school.name}${item.name}`"
                :value="item.zone_id">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="公里数">
            <el-input-number v-model="addForm.dis" size="default" :min="0.1" :max="15.0" :step="0.1" :precision="1" label="公里数"></el-input-number>
          </el-form-item>

          <el-divider>设置时间</el-divider>

          <div v-for="(item, index) in planTimeConfig" :key="index" class="plan-time-container">
            <div class="header">
              <div class="title">时段 {{ index + 1 }} {{ planTimeShow(index) }}</div>
              <div>
                <el-button type="danger" size="small" icon="Delete" @click="removePlanTime(index)">删除时段</el-button>
              </div>
            </div>
            <div class="content">
              <el-form-item label="跑步时段">
                <el-time-picker
                  v-model="planTimeConfig[index].time_range"
                  is-range
                  format="HH:mm"
                  value-format="HH:mm"
                  range-separator="至"
                  start-placeholder="开始时间"
                  end-placeholder="结束时间" />
              </el-form-item>
              <el-form-item label="开始日期">
                <el-date-picker
                  v-model="planTimeConfig[index].start_date"
                  type="date"
                  format="YYYY-MM-DD"
                  value-format="YYYY-MM-DD"
                  placeholder="选择开始日期" />
              </el-form-item>
              <el-form-item label="跑步天数">
                <el-input-number v-model="planTimeConfig[index].num" size="default" :min="1" :max="365" :step="1" :precision="0" label="跑步天数"></el-input-number>
              </el-form-item>
              <el-form-item label="跑步周期">
                <el-select
                  v-model="planTimeConfig[index].week"
                  multiple
                  placeholder="选择跑步周期">
                  <el-option
                    v-for="item in weekOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value" />
                </el-select>
              </el-form-item>
              <div class="footer">
                <div class="task-item" v-for="item in planTimeList[index].task_list" :key="item.start_time">{{ item.start_time }}</div>
              </div>
            </div>
          </div>

          <div class="d-flex justify-content-center align-items-center">
            <div style="margin-right: 10px;">{{ taskTotal }} 次 x {{ price }} 元 = {{ (price * taskTotal).toFixed(4) }}元</div>
            <el-button type="primary" size="small" @click="addPlanTime">新增时段</el-button>
          </div>

        </el-form>
        <template #footer>
          <el-button @click="addDialogVisible = false">取 消</el-button>
          <el-button type="primary" @click="handleAdd" :loading="addLoading">确认下单</el-button>
        </template>
      </el-dialog>
      <!-- 添加弹窗 - 结束 -->

      <!-- 查看日志弹窗 - 开始 -->
      <el-dialog title="查看日志" v-model="logDialogVisible" top="30px">
        <el-table :data="logData.list" border style="width: 100%">
          <el-table-column align="center" label="状态" width="120">
            <template #default="scope">
              <el-tag type="success" v-if="scope.row.status_display == '成功'">已完成</el-tag>
              <el-tag type="primary" v-else-if="scope.row.status_display == '未开始'">等待跑步</el-tag>
              <el-tag type="danger" v-else-if="scope.row.status_display == '失败'">失败</el-tag>
              <el-tag type="info" v-else-if="scope.row.status_display == '退款'">已退款</el-tag>
              <el-tag type="warning" v-else>{{ scope.row.status_display }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column align="center" label="跑步时间" width="350">
            <template #default="scope">
              <div v-if="changeTaskTimeForm.run_task_id == scope.row.run_task_id" class="d-flex">
                <div>
                  <el-date-picker
                    v-model="changeTaskTimeForm.start_time"
                    type="datetime"
                    size="small"
                    format="YYYY-MM-DD HH:mm:ss"
                    value-format="YYYY-MM-DD HH:mm:ss"
                    placeholder="请选择跑步时间" />
                </div>
                <div class="d-flex justify-content-center">
                  <div class="mx-2"><el-button size="small" type="success" icon="Check" @click="handleSaveTaskTime" /></div>
                  <div><el-button size="small" type="danger" icon="Close" @click="handleCancelChangeTaskTime" /></div>
                </div>
              </div>
              <div v-else>{{ scope.row.start_time }}</div>
            </template>
          </el-table-column>
          <el-table-column align="center" label="操作" width="120">
            <template #default="scope">
              <el-button v-if="scope.row.status != 'SUCCESS' && scope.row.status != 'REFUND'" type="primary" size="small" @click="handleChangeTaskTime(scope.row)">修改时间</el-button>
            </template>
          </el-table-column>
          <el-table-column align="center" label="信息" prop="info"></el-table-column>
        </el-table>

        <template #footer>
          <!-- 分页 -->
          <div class="d-none d-md-flex" style="display: flex;justify-content: center;margin-top: 20px;">
            <el-pagination background @size-change="handleLogSizeChange" @current-change="handleLogCurrentChange" :current-page.sync="logData.pagination.page"
              :page-sizes="[10, 20, 50, 100, 200]" :page-size="logData.pagination.limit"
              layout="total, sizes, prev, pager, next" :total="logData.pagination.total">
            </el-pagination>
          </div>

          <div class="d-md-flex d-md-none" style="display: flex;justify-content: center;margin-top: 20px;">
            <el-pagination background pager-count="3" @size-change="handleLogSizeChange" @current-change="handleLogCurrentChange" :current-page.sync="logData.pagination.page" :page-sizes="[10, 20, 50, 100, 200]" :page-size="logData.pagination.limit" layout="prev, pager, next" :total="logData.pagination.total">
            </el-pagination>
          </div>
        </template>
      </el-dialog>
      <!-- 查看日志弹窗 - 结束 -->

    </div>
  </div>
</div>

<?php require_once("footer.php"); ?>
<script src="/flash/static/axios.min.js"></script>
<script src="/flash/static/vue.global.min.js"></script>
<script src="/flash/static/zh-cn.min.js"></script>
<script src="/flash/static/element-plus.js"></script>
<script src="/flash/static/element-plus-icons.min.js"></script>
<script src="/flash/static/sdxy.js"></script>