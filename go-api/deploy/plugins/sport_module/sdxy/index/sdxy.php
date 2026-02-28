<?php
$title = '闪电闪动';
?>
<link rel="stylesheet" href="/sdxy/static/element-plus.css">
<div id="content" class="lyear-layout-content" role="main">
  <div class="app-content-body">
    <div class="wrapper-md control" id="app">

      <!-- 添加弹窗 - 开始 -->
      <el-dialog id="addDialog" title="添加订单" v-model="addDialogVisible" :before-close="addHandleClose" top="30px">
        <el-form label-width="80px">
          <el-form-item label="手机号">
            <el-input v-model="addForm.user" size="default" style="width: 250px;" placeholder="请输入账号"></el-input>
          </el-form-item>
          <el-form-item label="用户密码">
            <el-input v-model="addForm.pass" size="default" style="width: 250px;" placeholder="请输入密码"></el-input>
          </el-form-item>
          <el-form-item label="跑区简称">
            <el-input v-model="addForm.school" size="default" style="width: 250px;" placeholder="例如：东校区"></el-input>
          </el-form-item>
          <el-form-item label="日公里数">
            <el-input-number v-model="addForm.distance" size="default" :min="1" :max="100"
              :precision="1" :step="0.1" label="日公里数"></el-input-number>
          </el-form-item>
          <el-form-item label="跑步天数">
            <el-input-number v-model="addForm.day" size="default" :min="1" :max="365"
              :precision="0" :step="1" label="跑步天数"></el-input-number>
          </el-form-item>
          <el-form-item label="开始日期">
            <el-date-picker
              size="default"
              v-model="addForm.start_date"
              type="date"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              placeholder="选择日期">
            </el-date-picker>
          </el-form-item>
          <el-form-item label="开始小时">
            <el-input-number v-model="addForm.start_hour" size="default" :min="1" :max="24"
              :precision="0" :step="1" label="开始小时"></el-input-number>
          </el-form-item>
          <el-form-item label="开始分钟">
            <el-input-number v-model="addForm.start_minute" size="default" :min="0" :max="60"
              :precision="0" :step="1" label="开始分钟"></el-input-number>
          </el-form-item>
          <el-form-item label="结束小时">
            <el-input-number v-model="addForm.end_hour" size="default" :min="1" :max="24"
              :precision="0" :step="1" label="结束小时"></el-input-number>
          </el-form-item>
          <el-form-item label="结束分钟">
            <el-input-number v-model="addForm.end_minute" size="default" :min="0" :max="60"
              :precision="0" :step="1" label="结束分钟"></el-input-number>
          </el-form-item>
          <el-form-item label="跑步周期">
            <el-select v-model="addForm.run_week" multiple placeholder="请选择跑步周期" size="default">
              <el-option
                v-for="item in weekOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="价格总计">
            <el-input class="col-md-3" v-model="total" size="default" disabled="" style="width: 200px;"></el-input>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="addHandleClose">取 消</el-button>
          <el-button type="primary" size="default" @click="handleAdd">确认下单</el-button>
        </template>
      </el-dialog>
      <!-- 添加弹窗 - 结束 -->

      <!-- 列表 - 开始 -->
      <el-card class="box-card mb-4">
        <div class="card-body">
          <el-alert
            title="下单前确认好；账号；密码正确，有问题的订单会进行退款！"
            type="success">
          </el-alert>
          <el-alert
            title="建议一周起下，相当稳定！"
            type="success">
          </el-alert>
          <el-alert
            title="跑单中间，切勿上号，以免造成登录设备过多，本平台保证一号一设备，不会有新增设备登录！"
            type="error">
          </el-alert>
          <el-alert
            title="下单后请先联系授权客服进行短信验证码授权 QQ：2523241886 ！请提醒客户登录一次就要找客服接一次验证码！"
            type="success">
          </el-alert>

          <div class="row mt-3">
            <div class="col-md-2 mb-3">
              <el-select size="medium" v-model="search.status" placeholder="请选择订单状态">
                <el-option label="所有订单"></el-option>
                <el-option label="等待处理" value="1"></el-option>
                <el-option label="处理成功" value="2"></el-option>
                <el-option label="退款成功" value="3"></el-option>
              </el-select>
            </div>

            <div class="col-md-10 mb-3">
              <el-input size="medium" placeholder="关键词" v-model="search.keywords"
                class="input-with-select">
                <template #prepend>
                  <el-select v-model="search.type" size="medium" style="width: 100px" placeholder="条件"
                    slot="">
                    <el-option label="订单ID" value="1"></el-option>
                    <el-option label="下单账号" value="2"></el-option>
                    <el-option label="下单密码" value="3"></el-option>
                    <?php if ($userrow['uid'] == 1) { ?>
                      <el-option label="用户UID" value="4"></el-option>
                    <?php } ?>
                  </el-select>
                </template>

                <template #append>
                  <el-button size="default" @click="get">
                    <el-icon style="vertical-align: middle">
                      <Search />
                    </el-icon>
                    <span style="vertical-align: middle"> 搜索 </span>
                  </el-button>
                </template>
              </el-input>
            </div>
          </div>

          <div class="mb-3">
            <el-button type="primary" size="medium" plain @click="addDialogVisible = true">添加订单</el-button>
            <el-button type="warning" size="medium">价格：{{ price }} / 公里</el-button>
          </div>

          <el-table :data="tableData" v-loading="tableLoading" highlight-current-row border>
            <el-table-column align="center" prop="id" label="ID" width="100"></el-table-column>
            <?php if ($userrow['uid'] == 1) { ?>
              <el-table-column align="center" prop="uid" label="UID" width="80"></el-table-column>
            <?php } ?>
            <el-table-column align="center" label="账号" width="120">
              <template v-slot:default="scope">
                <div>{{ scope.row.user }}</div>
                <div>{{ scope.row.pass }}</div>
              </template>
            </el-table-column>
            <el-table-column align="center" prop="school" label="学校" width="120"></el-table-column>
            <el-table-column align="center" label="日公里数" width="100">
              <template v-slot:default="scope">{{ scope.row.distance }} KM</template>
            </el-table-column>
            <el-table-column align="center" label="天数" width="60">
              <template v-slot:default="scope">{{ scope.row.day }}</template>
            </el-table-column>
            <el-table-column align="center" label="开始日期" width="120">
              <template v-slot:default="scope">{{ scope.row.start_date }}</template>
            </el-table-column>
            <el-table-column align="center" label="开始时间" width="100">
              <template v-slot:default="scope">{{ scope.row.start_hour }}:{{ scope.row.start_minute }}</template>
            </el-table-column>
            <el-table-column align="center" label="结束时间" width="100">
              <template v-slot:default="scope">{{ scope.row.end_hour }}:{{ scope.row.end_minute }}</template>
            </el-table-column>
            <el-table-column align="center" label="跑步周期" width="120">
              <template v-slot:default="scope">
                <span v-if="scope.row.run_week.indexOf(1) > -1">周一</span>
                <span v-if="scope.row.run_week.indexOf(2) > -1">周二</span>
                <span v-if="scope.row.run_week.indexOf(3) > -1">周三</span>
                <span v-if="scope.row.run_week.indexOf(4) > -1">周四</span>
                <span v-if="scope.row.run_week.indexOf(5) > -1">周五</span>
                <span v-if="scope.row.run_week.indexOf(6) > -1">周六</span>
                <span v-if="scope.row.run_week.indexOf(7) > -1">周日</span>
              </template>
            </el-table-column>
            <el-table-column align="center" label="订单状态" width="100">
              <template v-slot:default="scope">
                <el-tag type="primary" v-if="scope.row.status==1">等待处理</el-tag>
                <el-tag type="success" v-else-if="scope.row.status==2">处理成功</el-tag>
                <el-tag type="danger" v-else-if="scope.row.status==3">退款成功</el-tag>
              </template>
            </el-table-column>
            <el-table-column align="center" prop="remarks" label="备注" min-width="120"></el-table-column>
            <el-table-column align="center" prop="addtime" label="下单时间" width="120"></el-table-column>
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

    </div>
  </div>
</div>

<script src="/sdxy/static/vue.global.min.js"></script>
<script src="/sdxy/static/axios.min.js"></script>
<script src="/sdxy/static/element-plus.js"></script>
<script src="/sdxy/static/zh-cn.min.js"></script>
<script src="/sdxy/static/element-plus-icons.min.js"></script>
<script src="/sdxy/static/sdxy.js?v=1.0"></script>