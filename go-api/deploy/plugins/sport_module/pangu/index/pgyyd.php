<?php
$title = '盘古云运动';
require_once('head.php');
?>
<link rel="stylesheet" href="/pangu/static/element.css" />
<style>
  .card-header {
    margin-top: 20px;
    margin-bottom: 10px;
  }

  .card-body {
    margin-top: 10px;
    margin-bottom: 20px;
  }

  .el-form-item {
    margin-bottom: 15px;
  }

  .el-form-item__label {
    line-height: 30px;
  }

  .el-form-item__content {
    line-height: 30px;
  }

  .el-form-item__error {
    padding-top: 0;
  }

  .el-dialog__body {
    height: auto;
    overflow: auto;
  }

  #addDialog .el-dialog__body {
    height: 75vh;
    overflow: auto;
  }
</style>
<div id="content" role="main">
  <div class="app-content-body">
    <div class="wrapper-md control" id="pangu">
      <!-- 列表 - 开始 -->
      <el-card class="box-card">
        <el-alert title="下单前请先完成授权,授权地址：https://csauth.666hhh.sbs/" type="warning"></el-alert>
        <div style="height: 10px"></div>
        <el-button type="primary" @click="()=>{addDialog = true}" size="small" plain>提交订单</el-button>
        <el-button type="success" @click="()=>{window.open('https://csauth.666hhh.sbs/', '_blank')}" size="small" plain>授权链接</el-button>
        <el-button type="warning" size="small">价格：{{price}} / 次</el-button>
        <div style="height: 20px"></div>
        <el-input placeholder="请输入关键词" v-model="search.keywords" class="input-with-select" size="small">
          <el-select v-model="search.type" style="width: 100px" placeholder="条件" slot="prepend">
            <el-option label="订单ID" :value="1"></el-option>
            <el-option label="账号UID" :value="2"></el-option>
            <?php if ($userrow['uid'] == 1) { ?>
              <el-option label="UID" :value="3"></el-option>
            <?php } ?>
          </el-select>
          <el-button slot="append" icon="el-icon-search" @click="get"></el-button>
        </el-input>
        <div style="height: 20px"></div>

        <el-table :data="tableData" v-loading="tableLoading" element-loading-text="载入数据中" ref="multipleTable"
          size="small" empty-text="啊哦！一条订单都没有哦！" highlight-current-row border>
          <el-table-column prop="id" label="ID" width="80" align="center">
          </el-table-column>
          <?php if ($userrow['uid'] == 1) { ?>
            <el-table-column label="UID" prop="user_id" width="80" align="center">
            </el-table-column>
          <?php } ?>
          <el-table-column prop="uid" label="账号UID" width="120" align="center">
          </el-table-column>
          <el-table-column prop="residue_num" label="剩余次数" width="80" align="center">
            <template slot-scope="{ row }">
              <el-tag v-if="row.residue_num == 0" type="danger">0 次</el-tag>
              <el-tag v-if="row.residue_num > 0" type="success">{{ row.residue_num }} 次</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="跑区" width="150" align="center">
            <template slot-scope="{ row }">
              <el-tag type="success">{{ row.zone_name }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="跑步距离" width="80" align="center">
            <template slot-scope="{ row }">
              {{ row.run_meter }} KM
            </template>
          </el-table-column>
          <el-table-column label="跑步类型" width="100" align="center">
            <template slot-scope="{ row }">
              <el-tag type="success">随机跑</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="account_flag" label="授权状态" width="100" align="center">
            <template slot-scope="{ row }">
              <el-tag v-if="row.account_flag == 1" type="success">已授权</el-tag>
              <el-tag v-if="row.account_flag == 0" type="danger">未授权</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="订单状态" width="100" align="center">
            <template slot-scope="{ row }">
              <el-tag v-if="row.status == 0" type="primary">未完成</el-tag>
              <el-tag v-if="row.status == 1" type="success">已完成</el-tag>
              <el-tag v-if="row.status == 2" type="danger">已暂停</el-tag>
              <el-tag v-if="row.status == 3" type="warning">今日异常</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="跑步状态" width="80" align="center">
            <template slot-scope="{ row }">
              <el-switch
                v-model="row.run_status"
                @change="changeRunStatus(row)"
                active-value="1"
                inactive-value="0">
              </el-switch>
            </template>
          </el-table-column>
          <el-table-column prop="mark_text" label="订单备注" min-width="80" align="center">
          </el-table-column>
          <el-table-column prop="created_at" label="下单时间" width="100" align="center">
          </el-table-column>
          <el-table-column label="操作" width="80" align="center">
            <template slot-scope="scope">
              <el-dropdown @command="handleCommand">
                <span class="el-dropdown-link">
                  <el-button type="primary" size="small">操作</el-button>
                </span>
                <el-dropdown-menu slot="dropdown">
                  <el-dropdown-item :command="{item:scope.row,type:'detailLog'}">查看日志</el-dropdown-item>
                  <el-dropdown-item :command="{item:scope.row,type:'addNum'}">增加次数</el-dropdown-item>
                  <el-dropdown-item :command="{item:scope.row,type:'runLog'}">跑步记录</el-dropdown-item>
                  <el-dropdown-item :command="{item:scope.row,type:'delete'}">申请退款</el-dropdown-item>
                </el-dropdown-menu>
              </el-dropdown>
            </template>
          </el-table-column>
        </el-table>

        <div style="text-align: center;margin-top: 20px;">
          <el-pagination background @size-change="handleChangeSize" @current-change="handleChangePage" :current-page.sync="pagination.page"
            :page-sizes="[10, 20, 50, 100, 200, 500]" :page-size="pagination.limit"
            layout="total,sizes, prev, pager, next, jumper" :total="pagination.total">
          </el-pagination>
        </div>
      </el-card>
      <!-- 列表 - 结束 -->
      <!--添加弹窗 - 开始-->
      <el-dialog id="addDialog" title="添加订单" :visible.sync="addDialog" :before-close="handleAddClose"
        size="small" top="20px">
        <el-form label-width="90px">
          <el-form-item label="UID">
            <el-input v-model="userInfo.uid" :disabled="userInfo.disable" size="small" placeholder="请输入UID"></el-input>
          </el-form-item>
          <el-form-item label="操作">
            <el-button type="primary" @click="getRule" size="mini">获取规则</el-button>
          </el-form-item>
          <el-form-item label="开始日期">
            <el-date-picker size="small" v-model="addForm.startDate" type="date" placeholder="选择日期"
              value-format="timestamp">
            </el-date-picker>
          </el-form-item>
          <el-form-item label="跑步类型">
            <el-select v-model="addForm.runTaskId" placeholder="请选择跑步类型" size="small" @change="changeRunZone">
              <el-option v-for="item in runZoneOptions" :key="item.id" :label="item.raName" :value="item.id">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="温馨提示">
            <span style="color: red">云运动需要手动设置里程等数据，请按照学校规则设置</span>
            <span style="color: red">{{ currentRunningRuleText }}</span>
          </el-form-item>
          <el-form-item label="自定义时间">
            <el-switch v-model="customTime.flag" size="small"></el-switch>
          </el-form-item>
          <el-form-item label="跑步时间">
            <!-- 预选时间段 - 开始 -->
            <span v-show="!customTime.flag">{{ addForm.startTimeRange || '请先选择跑步类型' }}</span>
            <!-- 预选时间段 - 结束 -->
            <!-- 自定义时间段 - 开始 -->
            <el-row v-show="customTime.flag">
              <el-col :md="8">
                <el-time-select
                  size="small"
                  v-model="customTime.start"
                  :picker-options="{ start: '05:00', step: '00:15', end: '23:30' }"
                  placeholder="请选择开始时间">
                </el-time-select>
              </el-col>
              <el-col :md="{span: 8, offset: 1}">
                <el-time-select
                  size="small"
                  v-model="customTime.end"
                  :picker-options="{ start: customTime.start, step: '00:15', end: '23:30' }"
                  placeholder="请选择结束时间">
                </el-time-select>
              </el-col>
            </el-row>
            <!-- 自定义时间段 - 结束 -->
          </el-form-item>
          <el-form-item label="跑步距离">
            <el-input-number v-model.number="addForm.runMeter" :precision="1" :step="0.1" :min="distanceMin"
              :max="10" label="跑步距离" size="small"></el-input-number>
          </el-form-item>
          <el-form-item label="跑步周期">
            <el-checkbox v-for="item in runWeekOptions" v-model="addForm.runWeekDay" :label="item.value">{{ item.label }}</el-checkbox>
          </el-form-item>
          <el-form-item label="跑步配速">
            <div style="display: flex;align-items: center;justify-content: space-between;">
              <div>当前配速：{{ currentSpeed }}</div>
              <el-switch v-model="customSpeed.flag" active-text="自定义">
              </el-switch>
            </div>
            <el-slider class="mx-3" v-model="addForm.paceRange" :disabled="!customSpeed.flag" range show-stops
              :min="customSpeed.min" :max="customSpeed.max">
            </el-slider>
          </el-form-item>
          <el-form-item label="下单次数">
            <el-input-number v-model.number="addForm.orderNum" :step="1" :min="1" :max="100" label="下单次数"
              size="small"></el-input-number>
            <span style="color: red;margin-left: 12px;font-size: 18px;font-weight: bold;">￥ {{ total }}</span>
          </el-form-item>
          <el-form-item label="备注">
            <el-input type="textarea" :autosize="{ minRows: 2, maxRows: 4}" placeholder="请输入内容"
              v-model="addForm.markText">
            </el-input>
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button @click="addDialog = false">取 消</el-button>
          <el-button type="primary" @click="handleAdd" :loading="addLoading">确认下单</el-button>
        </div>
      </el-dialog>
      <!--添加弹窗 - 结束-->
      <!--增加次数弹窗 - 开始-->
      <el-dialog title="增加次数" :visible.sync="addNumDialog" size="small" top="20px">
        <el-form label-width="120px">
          <el-form-item label="UID">
            <el-input v-model="publicForm.uid" disabled="" size="small" style="width: 180px;margin-bottom: 10px;"></el-input>
          </el-form-item>
          <el-form-item label="增加次数">
            <el-input-number v-model.number="publicForm.orderNum" :step="1" :min="1" :max="100" label="增加次数" size="small"></el-input-number>
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button @click="addNumDialog = false">取 消</el-button>
          <el-button type="primary" @click="handleAddNum" :loading="addNumLoading">确认提交</el-button>
        </div>
      </el-dialog>
      <!--增加次数弹窗 - 结束-->
      <!--订单执行日志弹窗 - 开始-->
      <el-dialog title="订单执行日志" :visible.sync="logDialog" fullscreen :before-close="handleLogClose">
        <el-table :data="logForm.list" v-loading="logLoading" border size="small">
          <el-table-column align="center" property="id" label="ID" width="80"></el-table-column>
          <el-table-column align="center" property="msg" label="后端返回信息" width="100"></el-table-column>
          <el-table-column align="center" label="跑步类型" width="100">
            <template slot-scope="{ row }">
              <el-tag type="success">随机跑</el-tag>
            </template>
          </el-table-column>
          <el-table-column align="center" label="状态" width="120">
            <template slot-scope="{ row }">
              <el-tag v-if="row.status == 0" type="info">等待执行中</el-tag>
              <el-tag v-if="row.status == 1" type="success">跑步完成</el-tag>
              <el-tag v-if="row.status == 2" type="success">队列执行中</el-tag>
              <el-tag v-if="row.status == -1" type="danger">跑步失败</el-tag>
            </template>
          </el-table-column>
          <el-table-column align="center" property="peishu" label="配速" width="80"></el-table-column>
          <el-table-column align="center" property="run_meter" label="距离" width="80"></el-table-column>
          <el-table-column align="center" property="start_timestamp" label="实际跑步时间">
            <template slot-scope="{ row }">
              {{ row.start_timestamp | formatTime }}
            </template>
          </el-table-column>
          <el-table-column align="center" property="execution_time" label="计划跑步时间"></el-table-column>
          <el-table-column align="center" label="操作" width="150">
            <template slot-scope="{ row }">
              <el-button v-if="row.status == 0" type="primary" size="small" @click="beforeEditRunTime(row)">改时间</el-button>
              <el-button v-if="row.status == -1" type="warning" size="small" @click="handleReRun(row)">重跑</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div style="text-align: center;margin-top: 20px;">
          <el-pagination background @size-change="handleLogChangeSize" @current-change="handleLogChangePage" :current-page.sync="logForm.pagination.page"
            :page-sizes="[10, 20, 50, 100, 200, 500]" :page-size="logForm.pagination.limit"
            layout="total,sizes, prev, pager, next, jumper" :total="logForm.pagination.total">
          </el-pagination>
        </div>
      </el-dialog>
      <!--订单执行日志弹窗 - 结束-->
      <!--修改跑步时间弹窗 - 开始-->
      <el-dialog title="修改跑步时间" :visible.sync="editRunTimeDialog" size="small" top="20px">
        <el-form label-width="120px">
          <el-form-item label="跑步时间">
            <el-time-picker
              v-model="editRunTimeForm.startTime"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="选择跑步时间">
            </el-time-picker>
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button @click="editRunTimeDialog = false">取 消</el-button>
          <el-button type="primary" @click="handleEditRunTime" :loading="editRunTimeForm.loading">确认提交</el-button>
        </div>
      </el-dialog>
      <!--修改跑步时间弹窗 - 结束-->
    </div>
  </div>
</div>

<?php require_once("footer.php"); ?>
<script src="/pangu/static/axios.min.js"></script>
<script src="/pangu/static/vue.min.js"></script>
<script src="/pangu/static/vue-resource.min.js"></script>
<script src="/pangu/static/element.js"></script>
<script src="/pangu/static/dayjs.min.js"></script>
<script src="/pangu/static/pgyyd.js"></script>