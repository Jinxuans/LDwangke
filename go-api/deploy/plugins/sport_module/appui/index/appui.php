<?php
$title = 'APPUI打卡';
require_once('head.php');
?>
<link rel="stylesheet" href="/appui/static/element.css" />
<style>
  .el-dialog__body {
    height: 80vh;
  }

  @media screen and (min-width: 768px) {
    .el-dialog__body {
      height: 70vh;
    }

    .edit-dialog .el-dialog__body {
      height: 50vh;
    }

    .renew-dialog .el-dialog__body {
      height: 15vh;
    }

    .el-dialog {
      width: 40%;
    }
  }

  .add-form .el-select {
    width: 100%;
  }

  .el-dropdown-menu {
    background: #fff;
  }
</style>

<div id="content" class="lyear-layout-content" role="main">
  <div class="app-content-body">
    <div class="wrapper-md control" id="app">
      <!-- 添加弹窗 - 开始 -->
      <el-dialog title="添加订单" :visible.sync="addDialogVisible" top="30px" :before-close="addHandleClose">
        <el-form label-width="80px" class="add-form">
          <el-form-item label="选择平台">
            <el-select v-model="addForm.pid" @change="changeCourse" size="small" placeholder="请选择平台">
              <el-option
                v-for="item in courseList"
                :key="item.pid"
                :label="`${item.name}（${item.price}/天）`"
                :value="item.pid">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="课程说明">
            <span style="color: red">{{ courseTip }}</span>
          </el-form-item>
          <el-form-item label="周期选择">
            <el-select multiple v-model="addForm.week" size="small" placeholder="请选择周期">
              <el-option
                v-for="item in weekOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="报表选择">
            <el-select multiple v-model="addForm.report" size="small" placeholder="请选择报表">
              <el-option
                v-for="item in reportOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="上班时间">
            <el-time-picker
              v-model="addForm.shangban_time"
              size="small"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="上班时间">
            </el-time-picker>
          </el-form-item>
          <el-form-item label="下班时间">
            <el-time-picker
              v-model="addForm.xiaban_time"
              size="small"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="下班时间">
            </el-time-picker>
          </el-form-item>
          <el-form-item label="用户学校" v-show="showSchoolInput">
            <el-select filterable v-model="addForm.school" size="small" placeholder="请选择学校">
              <el-option
                v-for="(item, index) in schoolList"
                :key="index"
                :label="item"
                :value="item">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="用户账号">
            <el-input v-model="addForm.user" size="small" placeholder="请输入账号"></el-input>
          </el-form-item>
          <el-form-item label="用户密码">
            <el-input v-model="addForm.pass" size="small" placeholder="请输入密码">
              <el-button slot="append" type="primary" size="small" @click="handleQuery">点击查询</el-button>
            </el-input>
          </el-form-item>
          <div v-show="addForm.userName">
            <el-form-item label="用户姓名">
              <el-input v-model="addForm.userName" size="small" placeholder="请输入用户姓名" disabled=""></el-input>
            </el-form-item>
            <el-form-item label="签到地址">
              <el-input v-model="addForm.address" type="textarea" size="small" placeholder="请输入签到地址"></el-input>
            </el-form-item>
            <el-form-item label="下单天数">
              <div class="row">
                <el-input-number v-model="addForm.days1" size="small" :min="1" :max="365" label="下单天数"></el-input-number>
                <span style="color: red;margin-left: 12px;font-weight: 800;font-size: 18px;">￥ {{ total }}</span>
              </div>
            </el-form-item>
          </div>
        </el-form>
        <div slot="footer">
          <el-button @click="addHandleClose">取 消</el-button>
          <el-button type="primary" @click="handleAdd" :loading="addLoading">确认下单</el-button>
        </div>
      </el-dialog>
      <!-- 添加弹窗 - 结束 -->
      <!-- 编辑弹窗 - 开始 -->
      <el-dialog class="edit-dialog" title="编辑弹窗" :visible.sync="editDialogVisible" top="30px">
        <el-form label-width="80px" class="add-form">
          <el-form-item label="周期选择">
            <el-select multiple v-model="editForm.week" size="small" placeholder="请选择周期">
              <el-option
                v-for="item in weekOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="报表选择">
            <el-select multiple v-model="editForm.report" size="small" placeholder="请选择报表">
              <el-option
                v-for="item in reportOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="上班时间">
            <el-time-picker
              v-model="editForm.shangban_time"
              size="small"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="上班时间">
            </el-time-picker>
          </el-form-item>
          <el-form-item label="下班时间">
            <el-time-picker
              v-model="editForm.xiaban_time"
              size="small"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="下班时间">
            </el-time-picker>
          </el-form-item>
          <el-form-item label="用户密码">
            <el-input v-model="editForm.pass" size="small" placeholder="请输入用户密码"></el-input>
          </el-form-item>
          <el-form-item label="签到地址">
            <el-input v-model="editForm.address" type="textarea" size="small" placeholder="请输入签到地址"></el-input>
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button @click="editDialogVisible = false">取 消</el-button>
          <el-button type="primary" @click="handleEdit" :loading="editLoading">确认修改</el-button>
        </div>
      </el-dialog>
      <!-- 编辑弹窗 - 结束 -->
      <!-- 续费弹窗 - 开始 -->
      <el-dialog class="renew-dialog" title="续费弹窗" :visible.sync="renewDialogVisible" top="30px">
        <el-form label-width="80px" class="add-form">
          <el-form-item label="下单天数">
            <div class="row">
              <el-input-number v-model="publicForm.days1" size="small" :min="1" :max="365" label="下单天数"></el-input-number>
              <span style="color: red;margin-left: 12px;font-weight: 800;font-size: 18px;">￥ {{ renewTotal }}</span>
            </div>
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button @click="renewDialogVisible = false">取 消</el-button>
          <el-button type="primary" @click="handleRenew" :loading="renewLoading">确认续费</el-button>
        </div>
      </el-dialog>
      <!-- 续费弹窗 - 结束 -->
      <!-- 日志弹窗 - 开始 -->
      <el-dialog title="日志列表" :visible.sync="detailDialogVisible" fullscreen>
        <el-table :data="detailList">
          <el-table-column property="id" label="ID" width="100"></el-table-column>
          <el-table-column label="签到状态" width="100">
            <template slot-scope="scope">
              <el-tag v-if="scope.row.qd_status == '已签'" type="success">{{ scope.row.qd_status }}</el-tag>
              <el-tag v-else-if="scope.row.qd_status == '不签到'" type="info">{{ scope.row.qd_status }}</el-tag>
              <el-tag v-else-if="scope.row.qd_status == '待签'" type="primary">{{ scope.row.qd_status }}</el-tag>
              <el-tag v-else-if="scope.row.qd_status == '异常'" type="danger">{{ scope.row.qd_status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column property="qd_time" label="签到时间" width="150"></el-table-column>
          <el-table-column property="qd_msg" label="签到返回"></el-table-column>
          <el-table-column label="签退状态" width="100">
            <template slot-scope="scope">
              <el-tag v-if="scope.row.qt_status == '已签'" type="success">{{ scope.row.qt_status }}</el-tag>
              <el-tag v-else-if="scope.row.qt_status == '不签到'" type="info">{{ scope.row.qt_status }}</el-tag>
              <el-tag v-else-if="scope.row.qt_status == '待签'" type="primary">{{ scope.row.qt_status }}</el-tag>
              <el-tag v-else-if="scope.row.qt_status == '异常'" type="danger">{{ scope.row.qt_status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column property="qt_time" label="签退时间" width="150"></el-table-column>
          <el-table-column property="qt_msg" label="签退返回"></el-table-column>
          <el-table-column property="addtime" label="生成时间"></el-table-column>
        </el-table>
      </el-dialog>
      <!-- 日志弹窗 - 结束 -->
      <!-- 列表 - 开始 -->
      <el-card class="box-card">
        <div class="card-body">
          <div class="row">
            <div class="col-md-6 mb-2">
              <el-button type="primary" size="medium" plain @click="addDialogVisible = true">添加订单</el-button>
            </div>
            <el-input class="col-md-6 mb-2" size="medium" placeholder="关键词" v-model="search.keywords"
              class="input-with-select">
              <el-select v-model="search.type" size="medium" style="width: 100px" placeholder="条件"
                slot="prepend">
                <el-option label="订单ID" value="1"></el-option>
                <el-option label="下单账号" value="2"></el-option>
                <el-option label="下单密码" value="3"></el-option>
                <?php if ($userrow['uid'] == 1) { ?>
                  <el-option label="用户UID" value="4"></el-option>
                <?php } ?>
              </el-select>
              <el-button slot="append" size="medium" icon="el-icon-search" @click="get">搜索</el-button>
            </el-input>
          </div>
          <el-table :data="tableData" style="width: 100%" v-loading="tableLoading">
            <el-table-column align="center" prop="id" label="ID" width="80"></el-table-column>
            <?php if ($userrow['uid'] == 1) { ?>
              <el-table-column align="center" prop="uid" label="UID" width="60"></el-table-column>
            <?php } ?>
            <el-table-column align="center" label="平台名称" width="120">
              <template slot-scope="scope">
                <div>{{ scope.row.pid | formatPtName }}</div>
              </template>
            </el-table-column>
            <el-table-column align="center" label="账号信息" width="120">
              <template slot-scope="scope">
                <div>{{ scope.row.user }}</div>
                <div>{{ scope.row.pass }}</div>
              </template>
            </el-table-column>
            <el-table-column align="center" prop="name" label="用户名称" width="100"></el-table-column>
            <el-table-column align="center" label="剩余/总天数" width="100">
              <template slot-scope="scope">{{ scope.row.residue_day }} / {{ scope.row.total_day }}</template>
            </el-table-column>
            <el-table-column align="center" prop="status_display" label="订单状态" width="80">
              <template slot-scope="scope">
                <el-tag type="warning" v-if="scope.row.status=='进行中'">进行中</el-tag>
                <el-tag type="success" v-else-if="scope.row.status=='已完成'">已完成</el-tag>
                <el-tag type="primary" v-else-if="scope.row.status=='待处理'">{{ scope.row.status }}</el-tag>
                <el-tag type="danger" v-else>{{ scope.row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column align="center" prop="address" label="签到地址"></el-table-column>
            <el-table-column align="center" prop="addtime" label="下单时间" width="100"></el-table-column>
            <el-table-column align="center" label="用户操作" width="120">
              <template slot-scope="scope">
                <el-dropdown @command="handleMenu">
                  <el-button type="primary" size="mini">
                    操作<i style="color: #fff" class="el-icon-arrow-down el-icon--right"></i>
                  </el-button>
                  <el-dropdown-menu slot="dropdown">
                    <el-dropdown-item
                      :command="{ type: 'refresh', item: scope.row }">刷新订单</el-dropdown-item>
                    <el-dropdown-item
                      :command="{ type: 'renew', item: scope.row }">增加天数</el-dropdown-item>
                    <el-dropdown-item
                      :command="{ type: 'detail', item: scope.row }">查看日志</el-dropdown-item>
                    <el-dropdown-item
                      :command="{ type: 'edit', item: scope.row }">编辑订单</el-dropdown-item>
                    <el-dropdown-item
                      :command="{ type: 'refund', item: scope.row }">申请退款</el-dropdown-item>
                  </el-dropdown-menu>
                </el-dropdown>
              </template>
            </el-table-column>
          </el-table>

          <div style="display: flex;justify-content: center;margin-top: 20px;">
            <el-pagination background @size-change="handleSizeChange" @current-change="handleCurrentChange"
              :current-page="pagination.page" :page-sizes="[20, 50, 100, 300]" :page-size="pagination.limit"
              layout="total, sizes, prev, pager, next" :total="pagination.total">
            </el-pagination>
          </div>

        </div>
      </el-card>
      <!-- 列表 - 结束 -->
    </div>
  </div>

  <?php require_once("footer.php"); ?>
  <script src="/appui/static/axios.min.js"></script>
  <script src="/appui/static/vue.min.js"></script>
  <script src="/appui/static/vue-resource.min.js"></script>
  <script src="/appui/static/element.js"></script>
  <script src="/appui/static/appui.js"></script>