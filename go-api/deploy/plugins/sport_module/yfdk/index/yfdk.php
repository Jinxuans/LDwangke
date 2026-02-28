<?php
$title = 'YF打卡';
include('head.php');
?>

<div class="lyear-layout-content" style="margin:0; padding:20px; width:100%; box-sizing:border-box;">
    <div id="qg">
        <el-dialog :title="dialog.title + '订单'" :visible.sync="dialog.order" :width="width" :top="top">
            <div style="height:5px"></div>
            <el-form ref="ruleFormRef" :model="submitForm" :rules="getCurrentRules()"
                     label-width="120px" size="small" class="clean-form" :show-message="false" status-icon>

                <div class="basic-section">
                    <el-form-item label="学校" v-if="classInfo.school == 1">
                        <el-select
                                v-model="submitForm.school"
                                filterable
                                clearable
                                placeholder="请选择学校"
                                :filter-method="filterSchools"
                                @focus="handleSchoolFocus">
                            <el-option
                                    v-for="school in filteredSchoolOptions"
                                    :key="school.id + '_' + school.name"
                                    :label="school.name"
                                    :value="getSchoolValue(school)">
                            </el-option>
                        </el-select>
                    </el-form-item>

                    <el-form-item label="账号" prop="user">
                        <el-input v-model="submitForm.user" class="input-with-select" clearable>
                            <template #append>
                                <el-button icon="el-icon-search" @click="userQuery" :loading="queryLoad">
                                    获取账号信息
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>

                    <el-form-item label="密码" prop="pass">
                        <el-input v-model="submitForm.pass" clearable></el-input>
                    </el-form-item>

                    <el-form-item label="邀请码" v-if="classInfo.cid == '39'" prop="yzm_code">
                        <el-input v-model="submitForm.yzm_code" placeholder="请输入云实习助理邀请码" clearable></el-input>
                    </el-form-item>

                    <el-form-item label="通知邮箱">
                        <el-input v-model="submitForm.email" placeholder="(非必填)" clearable></el-input>
                    </el-form-item>

                    <el-form-item label="公众号推送">
                        <el-input v-model="submitForm.push_url" placeholder="关注微信公众号“showdoc推送服务”复制你的专属推送地址到此处" clearable></el-input>
                    </el-form-item>

                    <el-form-item label="实习岗位">
                        <el-input v-model="submitForm.offer" placeholder="部分官方不支持获取的请手动填写" clearable></el-input>
                    </el-form-item>
                </div>

                <div class="advanced-settings-toggle" @click="showAdvanced = !showAdvanced" :class="{ expanded: showAdvanced }">
                    <span>{{ showAdvanced ? '收起详情' : '展开详情' }}</span>
                    <el-icon class="toggle-icon">
                        <ArrowDown />
                    </el-icon>
                </div>

                <div class="advanced-content" :class="{ collapsed: !showAdvanced, expanded: showAdvanced }">
                    <div v-show="showAdvanced || dialog.title == '编辑'">
                        <el-form-item label="姓名">
                            <el-input v-model="submitForm.name" placeholder="用户姓名（可自定义或自动获取）" clearable></el-input>
                        </el-form-item>

                        <el-form-item label="公司名称" v-if="classInfo.cid == '1' || classInfo.cid == '15' || classInfo.cid == '16' || classInfo.cid == '17' || classInfo.cid == '30' || classInfo.cid == '36' || classInfo.cid == '37' || classInfo.cid == '38' || classInfo.cid == '39' || classInfo.cid == '40' || classInfo.cid == '41' || classInfo.cid == '42' || classInfo.cid == '43' || classInfo.cid == '46' || classInfo.cid == '48' || classInfo.cid == '49' || classInfo.cid == '50' || classInfo.cid == '51' || classInfo.cid == '52' || classInfo.cid == '53'">
                            <el-input v-model="submitForm.company" placeholder="公司名称（自动获取）" readonly></el-input>
                        </el-form-item>

                        <el-form-item label="公司地址" v-if="classInfo.cid == '14' || classInfo.cid == '15' || classInfo.cid == '16' || classInfo.cid == '30' || classInfo.cid == '38' || classInfo.cid == '39' || classInfo.cid == '40' || classInfo.cid == '41' || classInfo.cid == '43' || classInfo.cid == '44' || classInfo.cid == '45' || classInfo.cid == '48' || classInfo.cid == '49' || classInfo.cid == '50' || classInfo.cid == '51'">
                            <el-input v-model="submitForm.company_address" placeholder="公司详细地址（自动获取）" clearable readonly></el-input>
                        </el-form-item>

                        <el-form-item label="打卡地址" :prop="needsLocation() ? 'address' : ''" v-if="classInfo.cid != '1'">
                            <el-input
                                    v-model="submitForm.address"
                                    placeholder="打卡地址（可自定义或自动获取）"
                                    clearable>
                            </el-input>
                        </el-form-item>

                        <el-form-item label="经度" :prop="needsLocation() ? 'longitude' : ''" v-if="classInfo.cid != '1' && classInfo.cid != '30'">
                            <el-input v-model="submitForm.longitude" placeholder="经度" readonly></el-input>
                        </el-form-item>

                        <el-form-item label="纬度" :prop="needsLocation() ? 'latitude' : ''" v-if="classInfo.cid != '1' && classInfo.cid != '30'">
                            <el-input v-model="submitForm.latitude" placeholder="纬度" readonly></el-input>
                        </el-form-item>

                        <el-form-item label="实习计划" v-if="classInfo.cid == '15' || classInfo.cid == '17' || classInfo.cid == '36' || classInfo.cid == '38' || classInfo.cid == '40' || classInfo.cid == '41' || classInfo.cid == '42' || classInfo.cid == '43' || classInfo.cid == '44' || classInfo.cid == '45' || classInfo.cid == '49' || classInfo.cid == '51' || classInfo.cid == '52'">
                            <el-input v-model="submitForm.plan_name" placeholder="实习计划名称（自动获取）" clearable readonly></el-input>
                        </el-form-item>

                        <el-form-item label="批次日期" v-if="classInfo.cid == '38'">
                            <el-input v-model="submitForm.sxksrq" placeholder="实习批次日期（自动获取）" readonly></el-input>
                        </el-form-item>

                        <el-form-item label="设备ID" v-if="classInfo.cid == '35'">
                            <el-input v-model="submitForm.device_id" placeholder="今日校园设备ID" clearable></el-input>
                        </el-form-item>

                        <el-form-item label="设备信息" v-if="classInfo.cid == '35'">
                            <el-input v-model="submitForm.cpdaily_info" type="textarea" :rows="3" placeholder="今日校园CpdailyInfo" clearable></el-input>
                        </el-form-item>

                        <el-form-item label="打卡周期">
                            <el-checkbox-group v-model="submitForm.week">
                                <el-checkbox :label="1">周一</el-checkbox>
                                <el-checkbox :label="2">周二</el-checkbox>
                                <el-checkbox :label="3">周三</el-checkbox>
                                <el-checkbox :label="4">周四</el-checkbox>
                                <el-checkbox :label="5">周五</el-checkbox>
                                <el-checkbox :label="6">周六</el-checkbox>
                                <el-checkbox :label="7">周日</el-checkbox>
                            </el-checkbox-group>
                        </el-form-item>

                        <el-form-item label="打卡时间" prop="worktime">
                            <el-time-picker v-model="submitForm.worktime" placeholder="选择打卡时间" format="HH:mm" value-format="HH:mm" clearable></el-time-picker>
                        </el-form-item>

                        <el-form-item label="下班打卡" v-if="classInfo.cid == '14' || classInfo.cid == '15' || classInfo.cid == '16' || classInfo.cid == '17' || classInfo.cid == '20' || classInfo.cid == '36' || classInfo.cid == '37' || classInfo.cid == '43' || classInfo.cid == '44' || classInfo.cid == '46' || classInfo.cid == '49'">
                            <el-switch v-model="submitForm.offwork" :active-value="1" :inactive-value="0"></el-switch>
                        </el-form-item>

                        <el-form-item label="下班时间" v-if="submitForm.offwork">
                            <el-time-picker v-model="submitForm.offtime" placeholder="选择打卡时间" format="HH:mm" value-format="HH:mm" clearable></el-time-picker>
                        </el-form-item>

                        <el-form-item label="日周月报" v-if="classInfo.cid != '37' && classInfo.cid != '18'">
                            <div class="inline-report-config">
                                <div class="report-inline-item" v-if="classInfo.cid != '24' && classInfo.cid != '39' && classInfo.cid != '43'">
                                    <span class="report-label">日报</span>
                                    <el-switch v-model="submitForm.day_report" :active-value="1" :inactive-value="0"></el-switch>
                                </div>

                                <div class="report-inline-item">
                                    <span class="report-label">周报</span>
                                    <el-switch v-model="submitForm.week_report" :active-value="1" :inactive-value="0"></el-switch>
                                </div>

                                <div class="report-inline-item" v-if="classInfo.cid != '40' && classInfo.cid != '41' && classInfo.cid != '39' && classInfo.cid != '43'">
                                    <span class="report-label">月报</span>
                                    <el-switch v-model="submitForm.month_report" :active-value="1" :inactive-value="0"></el-switch>
                                </div>
                            </div>
                        </el-form-item>

                        <el-form-item label="周报日期" v-if="submitForm.week_report">
                            <el-select v-model="submitForm.week_date" placeholder="请选择">
                                <el-option label="周一" value="1"></el-option>
                                <el-option label="周二" value="2"></el-option>
                                <el-option label="周三" value="3"></el-option>
                                <el-option label="周四" value="4"></el-option>
                                <el-option label="周五" value="5"></el-option>
                                <el-option label="周六" value="6"></el-option>
                                <el-option label="周日" value="7"></el-option>
                            </el-select>
                        </el-form-item>

                        <el-form-item label="月报日期" v-if="submitForm.month_report && classInfo.cid != '40' && classInfo.cid != '41'">
                            <el-select v-model="submitForm.month_date" placeholder="请选择">
                                <el-option v-for="number in Array.from({ length: 31 }, (v, k) => k + 1)" :label="number + '号'" :value="number"></el-option>
                            </el-select>
                        </el-form-item>

                        <el-form-item label="上传图片" v-if="classInfo.cid == '15' || classInfo.cid == '16' || classInfo.cid == '35' || classInfo.cid == '36' || classInfo.cid == '46' || classInfo.cid == '49'">
                            <el-switch v-model="submitForm.image" :active-value="1" :inactive-value="0"></el-switch>
                        </el-form-item>

                        <el-form-item label="法定节假日">
                            <el-switch v-model="submitForm.skip_holidays" :active-value="1" :inactive-value="0"></el-switch>
                        </el-form-item>

                        <el-form-item label="打卡天数" prop="day">
                            <el-input-number v-model="submitForm.day" :min="1"></el-input-number>
                        </el-form-item>

                        <el-form-item label="预估扣费" v-if="dialog.title == '添加'">
                            <el-input v-model="submitForm.yuji" readonly></el-input>
                        </el-form-item>
                    </div>
                </div>

                <el-form-item style="margin-top: 30px;">
                    <el-button type="success" size="large" :loading="submitLoad" v-if="dialog.title == '添加'" @click="add" plain>
                        创建订单
                    </el-button>
                    <el-button type="success" size="large" :loading="submitLoad" v-if="dialog.title == '编辑'" @click="save" plain>
                        保存
                    </el-button>
                    <el-button size="large" @click="dialog.order=false">取消</el-button>
                </el-form-item>
            </el-form>
        </el-dialog>

        <el-dialog title="订单日志" :visible.sync="dialog.log" :width="width" :top="top">
            <el-table :data="drawer.list" v-loading="drawer.loading" max-height="500">
                <el-table-column type="index" width="50" align="center"></el-table-column>
                <el-table-column property="time" label="时间" align="center" width="180"></el-table-column>
                <el-table-column property="content" label="内容" show-overflow-tooltip></el-table-column>
            </el-table>
            <template #footer>
                <el-button type="warning" size="large" @click="dialog.log=false" round>关闭</el-button>
            </template>
        </el-dialog>

        <el-dialog title="补交报告" :visible.sync="dialog.reset" width="500px" :top="top">
            <div class="rice_tag" style="background: #f0f9ff; padding: 15px; border-radius: 8px; border-left: 4px solid #409eff;">
                <strong>补签规则说明：</strong><br>
                日报补签：按选择的日期范围内每天补签一次<br>
                周报补签：按选择的日期范围内每7天补签一次<br>
                月报补签：按选择的日期范围内每月补签一次<br>
            </div>
            <div style="height:15px"></div>
            <el-form label-width="auto" v-loading="report.loading">
                <el-form-item label="报告类型">
                    <el-radio-group v-model="report.type">
                        <el-radio label="day">日报</el-radio>
                        <el-radio label="week">周报</el-radio>
                        <el-radio label="month">月报</el-radio>
                    </el-radio-group>
                </el-form-item>
                <el-form-item label="开始日期" v-if="report.type">
                    <el-date-picker
                            v-model="report.start"
                            type="date"
                            format="yyyy-MM-dd"
                            value-format="yyyy-MM-dd"
                            placeholder="选择起始日期">
                    </el-date-picker>
                </el-form-item>
                <el-form-item label="结束日期" v-if="report.type">
                    <el-date-picker
                            v-model="report.end"
                            type="date"
                            format="yyyy-MM-dd"
                            value-format="yyyy-MM-dd"
                            placeholder="选择结束日期">
                    </el-date-picker>
                </el-form-item>
            </el-form>
            <template #footer>
                <div style="text-align: left; padding-left: 20px;">
                    <el-button type="success" @click="patchReport" round>提交</el-button>
                    <el-button type="warning" @click="dialog.reset=false" round>关闭</el-button>
                </div>
            </template>
        </el-dialog>

        <el-dialog title="公告" :visible.sync="popupNotice.visible" :width="noticeWidth" :top="noticeTop" custom-class="notice-dialog">
            <div v-html="popupNotice.content" style="line-height: 1.8; font-size: 14px;"></div>
            <template #footer>
                <el-button type="primary" size="large" @click="popupNotice.visible=false" round>确定</el-button>
            </template>
        </el-dialog>

        <div style="height:10px"></div>
        <el-card class="box-card">
            <div class="text item">
                <div style="padding: 10px 0;">
                <span style="color: #666; font-size: 14px;">
                    YF打卡 当前版本: <span style="color: #409eff; font-weight: 500;">v{{ localVersion }}</span>
                    <?php if ($userrow['uid'] == 1): ?>
                    <span class="version-status" :class="versionStatusClass" @click="hasNewVersion && (showUpdateDialog = true)" style="margin-left: 8px; padding: 2px 8px; border-radius: 4px; font-size: 12px; cursor: pointer;">
                        {{ versionStatusText }}
                    </span>
                    <?php endif; ?>
                    <el-divider direction="vertical"></el-divider>
                    <a href="https://dk.blwl.fun/Student" target="_blank" style="color: #67C23A; text-decoration: none; font-weight: 500;">
                        <i class="el-icon-upload"></i> 学生端（上传图片、自行操作）
                    </a>
                    <span v-if="sourceBalance !== null" style="margin-left: 15px; color: #E6A23C; font-weight: 500;">
                        <el-divider direction="vertical"></el-divider>
                        源台余额：{{ sourceBalance }}元
                    </span>
                </span>
                </div>
                <el-form :model="cx" size="small" inline>
                    <el-form-item>
                        <el-select v-model="add_cid" size="small" placeholder="请选择">
                            <el-option v-for="item in platform" :label="item.name" :value="item.cid"></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="warning" @click="checkAdd" size="small" plain>添加订单</el-button>
                    </el-form-item>
                    <el-form-item label="模糊查询">
                        <el-input
                                v-model="cx.keyword"
                                placeholder="账号/密码/姓名"
                                clearable
                                style="width: 200px"
                                @clear="query"
                                @keyup.enter.native="query">
                            <el-button slot="append" icon="el-icon-search" @click="query"></el-button>
                        </el-input>
                    </el-form-item>
                    <el-form-item label="状态筛选">
                        <el-select v-model="cx.status" placeholder="请选择" style="width:130px" clearable>
                            <el-option label="全部" value=""></el-option>
                            <el-option label="正常" value="1"></el-option>
                            <el-option label="暂停" value="0"></el-option>
                            <el-option label="已过期" value="2"></el-option>
                            <el-option label="即将到期" value="3"></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item label="平台">
                        <el-select v-model="cx.cid" placeholder="请选择">
                            <el-option v-for="item in platform" :label="item.name" :value="item.cid"></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" @click="query">搜索</el-button>
                    </el-form-item>
                </el-form>
                <div style="height:10px"></div>
                <el-table v-loading="loading" element-loading-text="正在加载中" max-height="700" :data="list"
                          size="small" highlight-current-row border>
                    <el-table-column type="index" label="序号" width="50" align="center"></el-table-column>
                    <el-table-column label="姓名" width="85" align="center" show-overflow-tooltip>
                        <template #default="scope">
                            <el-tag type="primary" effect="plain">
                                {{ scope.row.name || '未设置' }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column label="操作" width="260" align="center">
                        <template #default="scope">
                            <el-tooltip content="编辑" placement="top">
                                <el-button type="primary" @click="handleOrder(scope.row)" size="mini" icon="el-icon-edit" circle></el-button>
                            </el-tooltip>
                            <el-tooltip content="续费" placement="top">
                                <el-button type="success" @click="handleDays(scope.row.id)" size="mini" icon="el-icon-coin" circle></el-button>
                            </el-tooltip>
                            <el-tooltip content="日志" placement="top">
                                <el-button type="warning" @click="handleLog(scope.row.id)" size="mini" icon="el-icon-document" circle></el-button>
                            </el-tooltip>
                            <el-tooltip content="打卡" placement="top">
                                <el-button type="warning" @click="handleClock(scope.row.id)" size="mini" icon="el-icon-circle-check" circle></el-button>
                            </el-tooltip>
                            <el-tooltip content="补签" placement="top">
                                <el-button type="info" @click="handlePatchReport(scope.row)" size="mini" icon="el-icon-edit-outline" circle></el-button>
                            </el-tooltip>
                            <el-tooltip content="删除" placement="top">
                                <el-button type="danger" @click="handleDelete(scope.row)" size="mini" icon="el-icon-delete" circle></el-button>
                            </el-tooltip>
                        </template>
                    </el-table-column>
                    <el-table-column label="平台" width="160" align="center">
                        <template #default="scope">
                            <el-tag type="primary" effect="plain" size="small">
                                {{ getPlatformName(scope.row.cid) }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column label="账号" width="120" align="center" show-overflow-tooltip>
                        <template #default="scope">{{ scope.row.username }}</template>
                    </el-table-column>
                    <el-table-column label="密码" width="120" align="center" show-overflow-tooltip>
                        <template #default="scope">{{ scope.row.password }}</template>
                    </el-table-column>
                    <el-table-column label="最新日志" width="380" align="center" show-overflow-tooltip>
                        <template #default="scope">
                            <el-tag :type="getValidTagType(scope.row.mark)" effect="plain">
                                {{ scope.row.mark || '暂无日志' }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column label="开/关" width="70" align="center">
                        <template #default="scope">
                            <el-switch
                                    v-model="scope.row.status"
                                    :active-value="1"
                                    :inactive-value="0"
                                    :disabled="isOrderExpired(scope.row.endtime)"
                                    @change="handleStatus(scope.row.id, scope.row.status)">
                            </el-switch>
                        </template>
                    </el-table-column>
                    <el-table-column label="剩余天数" width="78" align="center">
                        <template #default="scope">
                            <el-tag :type="getRemainingDaysType(scope.row.endtime)" effect="plain">
                                {{ calculateRemainingDays(scope.row.endtime) }}天
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column prop="create_time" label="创建时间" width="155" align="center" sortable></el-table-column>
                </el-table>
                <el-divider></el-divider>
                <el-pagination @size-change="sizechange" @current-change="pagechange" :current-page.sync="currentpage" :page-sizes="[10, 20, 50, 100, 200, 500]" :page-size="pagesize" layout="total,sizes, prev, pager, next, jumper" :total="pagecount">
                </el-pagination>
                <el-divider></el-divider>
            </div>
        </el-card>


    </div>
</div>
<script type="text/javascript" src="assets/LightYear/js/jquery.min.js"></script>
<script type="text/javascript" src="assets/LightYear/js/bootstrap.min.js"></script>
<script type="text/javascript" src="assets/LightYear/js/perfect-scrollbar.min.js"></script>
<script type="text/javascript" src="assets/LightYear/js/main.min.js"></script>
<script src="assets/js/vue.min.js"></script>
<script src="assets/js/vue-resource.min.js"></script>
<script src="assets/js/element.js"></script>
<script type="text/javascript">
    var vm = new Vue({
        el: "#qg",
        data: {
            sourceBalance: null,
            width: $(window).width() > 400 ? ($(window).width() < 900 ? ($(window).width() > 650 ? '80%' : '95%') : '60%') : '95%',
            top: $(window).height() > 700 ? '50px' : '10px',
            noticeWidth: $(window).width() > 600 ? '500px' : '90%',
            noticeTop: '25vh',
            schoolLoading: false,
            schoolOptions: [],
            filteredSchoolOptions: [],
            dialog: {
                order: false,
                loading: false,
                details: false,
                reset: false,
                log: false,
                patchReport: false,
                title: '添加'
            },
            popupNotice: {
                visible: false,
                content: ''
            },
            drawer: {
                oid: '',
                show: false,
                loading: false,
                list: []
            },
            add_cid: null,
            loading: false,
            diaload: false,
            submitLoad: false,
            queryLoad: false,
            showAdvanced: false,
            yuji: null,
            currentPatchOrderId: null,
            list: [],
            platform: [],
            form: {
                day: "30"
            },
            cx: {
                status: '',
                cid: '',
                keyword: ''
            },
            classInfo:{
                name: '',
                content: '',
                price: '',
                cid: this.add_cid,
                school: 0
            },
            submitForm: {
                id: '',
                oid: '',
                cid: this.add_cid,
                user: '',
                pass: '',
                school: '',
                email: '',
                push_url: '',
                yzm_code: '',
                device_id: '',
                cpdaily_info: '',
                enrollment_year: '',
                device: '',
                name: '',
                offer: '',
                plan_name: '',
                company: '',
                company_address: '',
                address: '',
                longitude: '',
                latitude: '',
                week: [1, 2, 3, 4, 5, 6, 7],
                worktime: "0" + Math.round(Math.random() + 7) + ":" + ("0" + Math.floor(Math.random() * 60)).slice(-2),
                offwork: 0,
                offtime: '',
                day: 1,
                day_report: 1,
                week_report: 0,
                week_date: "7",
                month_report: 0,
                month_date: Math.floor(Math.random() * 8) + 20,
                image: 0,
                skip_holidays: 0,
                originalAddress: '',
                remark: '',
                yuji: '0.00元'
            },
            currentpage: 1,
            pagesize: 20,
            pagecount: 100,
            patchReportForm: {
                id: '',
                cid: '',
                type: 'day',
                startDate: '',
                endDate: '',
                submitting: false
            },
            report: {
                type: '',
                start: '',
                end: '',
                loading: false
            },
            localVersion: '1.4',
            remoteVersion: '',
            updateInfo: null,
            showUpdateDialog: false,
            updateLoading: false,
            hasNewVersion: false,
            versionStatusText: '检测中...',
            versionStatusClass: 'status-checking'
        },
        methods: {
            getSourceBalance: function() {
                var that = this;
                this.$http.post("/daka.php?act=getSourceBalance", {}, {
                    emulateJSON: true
                }).then(function(response) {
                    if (response.body.code == 0) {
                        that.sourceBalance = parseFloat(response.body.money).toFixed(2);
                    }
                }).catch(function(error) {
                });
            },

            filterSchools: function(query) {
                if (!query || query.trim() === '') {
                    this.filteredSchoolOptions = this.schoolOptions;
                } else {
                    const keyword = query.toLowerCase().trim();
                    this.filteredSchoolOptions = this.schoolOptions.filter(school => {
                        const schoolName = (school.name || '').toLowerCase();
                        return schoolName.includes(keyword);
                    });
                }
            },

            handleSchoolFocus: function() {
                const supportedCids = ['16', '24', '35', '39', '40', '42', '43', '44', '52'];
                if (!supportedCids.includes(String(this.classInfo.cid))) return;

                if (this.schoolOptions.length === 0) {
                    this.schoolLoading = true;

                    this.$http.post("/daka.php?act=getSchools", {
                        cid: this.classInfo.cid
                    }, {
                        emulateJSON: true
                    }).then(function(response) {
                        this.schoolLoading = false;
                        if (response.body.code === 0 && response.body.data) {
                            this.schoolOptions = response.body.data;
                            this.filteredSchoolOptions = response.body.data;
                        }
                    }).catch(function(error) {
                        this.schoolLoading = false;
                        this.$message.error('加载学校列表失败');
                    });
                } else {
                    this.filteredSchoolOptions = this.schoolOptions;
                }
            },

            loadPopupNotice: function() {
                this.$http.post("/daka.php?act=getPopupNotice", {}, {
                    emulateJSON: true
                }).then(function(response) {
                    if (response.body.code == 0 && response.body.data.has_notice) {
                        this.popupNotice.content = response.body.data.content;
                        this.popupNotice.visible = true;
                    }
                }).catch(function(error) {
                });
            },
            handleOrder: function(data) {
                const loading = this.$loading({
                    lock: true,
                    text: '正在从源台加载订单信息...'
                });

                this.$http.post("/daka.php?act=getOrderDetail", { id: data.id }, { emulateJSON: true })
                    .then(function(response) {
                        loading.close();

                        if (response.body.code !== 0) {
                            this.$message.error('获取订单详情失败');
                            return;
                        }

                        const api = response.body.data;

                        this.submitForm = {
                            id: data.id,
                            oid: api.order_id,
                            cid: data.cid,
                            user: api.username || '',
                            pass: api.password || '',
                            school: api.school || '',
                            name: api.name || '',
                            email: api.email || '',
                            push_url: api.push_url || '',
                            offer: api.offer || '',
                            plan_name: api.plan_name || '',
                            company: api.company || '',
                            company_address: api.company_address || '',
                            address: api.address || '',
                            longitude: api.long || '',
                            latitude: api.lat || '',
                            week: Array.isArray(api.week) ? api.week.map(Number) : [1,2,3,4,5,6,7],
                            worktime: api.worktime || '',
                            offwork: api.offwork || 0,
                            offtime: api.offtime || '',
                            day: data.day || 1,
                            day_report: api.day_report !== undefined ? api.day_report : 1,
                            week_report: api.week_report !== undefined ? api.week_report : 0,
                            week_date: api.week_date || "7",
                            month_report: api.month_report !== undefined ? api.month_report : 0,
                            month_date: api.month_date || 25,
                            image: api.image || 0,
                            skip_holidays: api.skip_holidays || 0,
                            enrollment_year: api.enrollment_year || '',
                            device_id: api.device_id || '',
                            cpdaily_info: api.cpdaily_info || '',
                            remark: api.remark || ''
                        };

                        this.classInfo.cid = data.cid;
                        this.dialog.title = '编辑';
                        this.dialog.order = true;
                        this.showAdvanced = true;
                    }).catch(function(error) {
                    loading.close();
                    this.$message.error('网络错误');
                });
            },
            userQuery: function() {
                const currentCid = this.dialog.title === '编辑' ? this.classInfo.cid : this.add_cid;

                if (!currentCid) {
                    this.$message.error("请先选择平台！");
                    return;
                }

                if (!this.submitForm.user || !this.submitForm.pass) {
                    this.$message.error("请先输入账号和密码！");
                    return;
                }

                if (currentCid == '39' && !this.submitForm.yzm_code) {
                    this.$message.error("请先输入邀请码！");
                    return;
                }

                this.queryLoad = true;
                const data = {
                    cid: currentCid,
                    user: this.submitForm.user,
                    pass: this.submitForm.pass,
                    school: this.submitForm.school,
                    yzm_code: this.submitForm.yzm_code
                };

                this.$http.post("/daka.php?act=getAccountInfo", data, { emulateJSON: true }).then(function(data) {
                    this.queryLoad = false;
                    if (data.body.code != 0) {
                        this.$message.error(data.body.msg || '获取账号信息失败');
                        return;
                    }

                    if(data.body.data.school) this.submitForm.school = data.body.data.school;
                    if(data.body.data.address) this.submitForm.address = data.body.data.address;
                    if(data.body.data.student) this.submitForm.name = data.body.data.student;
                    if(data.body.data.offer) this.submitForm.offer = data.body.data.offer;
                    if(data.body.data.company) this.submitForm.company = data.body.data.company;
                    if(data.body.data.plan_name) this.submitForm.plan_name = data.body.data.plan_name;
                    if(data.body.data.company_address) this.submitForm.company_address = data.body.data.company_address;
                    if(data.body.data.day !== undefined) this.submitForm.day_report = data.body.data.day;
                    if(data.body.data.week !== undefined) this.submitForm.week_report = data.body.data.week;
                    if(data.body.data.month !== undefined) this.submitForm.month_report = data.body.data.month;
                    if(data.body.data.long !== undefined && data.body.data.long !== null && data.body.data.long !== '') {
                        this.submitForm.longitude = data.body.data.long;
                    }
                    if(data.body.data.lat !== undefined && data.body.data.lat !== null && data.body.data.lat !== '') {
                        this.submitForm.latitude = data.body.data.lat;
                    }

                    this.showAdvanced = true;
                    this.$message.success('账号信息获取成功！');
                }).catch(function(error) {
                    this.queryLoad = false;
                    this.$message.error('网络请求失败，请检查网络连接');
                });
            },
            needsSchoolDropdown: function() {
                const dropdownProjects = ['16', '24', '39', '40', '42', '43', '44'];
                return dropdownProjects.includes(String(this.classInfo.cid));
            },
            getSchoolValue: function(school) {
                if (this.classInfo.cid == '39' || this.classInfo.cid == '42') {
                    return school.id;
                }
                return school.name;
            },
            checkAdd: function() {
                if(this.add_cid) {
                    const selectedPlatform = this.platform.find(p => p.cid == this.add_cid);

                    this.classInfo.cid = this.add_cid;
                    this.classInfo.school = selectedPlatform ? selectedPlatform.school : 0;
                    this.dialog.title = '添加';
                    this.schoolOptions = [];
                    this.filteredSchoolOptions = [];

                    this.submitForm = {
                        id: '',
                        oid: '',
                        cid: this.add_cid,
                        user: '',
                        pass: '',
                        school: '',
                        email: '',
                        push_url: '',
                        yzm_code: '',
                        device_id: '',
                        cpdaily_info: '',
                        enrollment_year: '',
                        device: '',
                        name: '',
                        offer: '',
                        plan_name: '',
                        company: '',
                        company_address: '',
                        address: '',
                        longitude: '',
                        latitude: '',
                        week: [1, 2, 3, 4, 5, 6, 7],
                        worktime: "0" + Math.round(Math.random() + 7) + ":" + ("0" + Math.floor(Math.random() * 60)).slice(-2),
                        offwork: 0,
                        offtime: '',
                        day: 1,
                        day_report: 1,
                        week_report: 0,
                        week_date: "7",
                        month_report: 0,
                        month_date: Math.floor(Math.random() * 8) + 20,
                        image: 0,
                        skip_holidays: 0,
                        originalAddress: '',
                        remark: '',
                        yuji: '0.00元'
                    };

                    this.$http.post("/daka.php?act=getmoney", {
                        cid: this.add_cid,
                        day: 1
                    }, { emulateJSON: true }).then(function(data) {
                        if (data.body.code == 1) {
                            this.submitForm.yuji = data.body.msg;
                        }
                    });

                    this.showAdvanced = false;
                    this.dialog.order = true;
                } else {
                    this.$message.error("请先选择平台再进行下单!");
                }
            },
            needsLocation: function() {
                const skipLocationProjects = [30];
                const currentCid = parseInt(this.classInfo.cid);
                return !skipLocationProjects.includes(currentCid);
            },
            getCurrentRules: function() {
                const baseRules = {
                    user: [{ required: true, message: '请先输入账号！', trigger: 'blur' }],
                    pass: [{ required: true, message: '请先输入密码！', trigger: 'blur' }],
                    worktime: [{ required: true, message: '请先输入打卡时间！', trigger: 'blur' }],
                    day: [{ required: true, message: '请先设置打卡天数！', trigger: 'blur' }],
                };

                if (this.needsLocation()) {
                    baseRules.address = [{ required: true, message: '请先输入地址！', trigger: 'blur' }];
                    baseRules.longitude = [{ required: true, message: '经度信息不能为空！', trigger: 'blur' }];
                    baseRules.latitude = [{ required: true, message: '纬度信息不能为空！', trigger: 'blur' }];
                }

                return baseRules;
            },
            calculateRemainingDays: function(endtime) {
                if (!endtime) return 0;
                const now = new Date();
                const end = new Date(endtime);
                const diffTime = end - now;
                const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
                return diffDays > 0 ? diffDays : 0;
            },
            getRemainingDaysType: function(endtime) {
                const days = this.calculateRemainingDays(endtime);
                if (days > 7) return 'success';
                if (days > 3) return 'warning';
                if (days > 0) return 'danger';
                return 'info';
            },
            getValidTagType: function(mark) {
                if (!mark) return 'info';

                const markLower = String(mark).toLowerCase();

                if (markLower.includes('成功') || markLower.includes('完成') ||
                    markLower.includes('提交成功') || markLower.includes('已完成') ||
                    markLower.includes('打卡成功') || markLower.includes('无需打卡') ||
                    markLower.includes('已经汇报') || markLower.includes('已汇报') ||
                    markLower.includes('汇报成功')) {
                    return 'success';
                }

                if (markLower.includes('失败') || markLower.includes('错误') ||
                    markLower.includes('异常') || markLower.includes('该岗位暂无模版')) {
                    return 'danger';
                }

                if (markLower.includes('等待') || markLower.includes('即将') ||
                    markLower.includes('创建任务') || markLower.includes('开始')) {
                    return 'primary';
                }

                if (markLower.includes('已过期')) {
                    return 'info';
                }

                if (markLower.includes('暂无日志')) {
                    return 'info';
                }

                return 'primary';
            },
            getPlatformName: function(cid) {
                const platform = this.platform.find(p => p.cid == cid);
                return platform ? platform.name : '未知平台';
            },
            sizechange: function(val) {
                this.pagesize = val;
                this.query();
            },
            pagechange: function(val) {
                this.currentpage = val;
                this.query();
            },
            getProjects: function() {
                this.loading = true;
                this.$http.post("/daka.php?act=getProjects").then(function(data) {
                    this.loading = false;
                    if (data.body.code == 0) {
                        this.platform = data.body.data;
                    }
                });
            },
            query: function() {
                this.loading = true;

                const data = {
                    cx: {
                        status: this.cx.status || '',
                        cid: this.cx.cid || '',
                        keyword: this.cx.keyword || ''
                    },
                    page: this.currentpage,
                    size: this.pagesize
                };

                this.$http.post("/daka.php?act=order", data, {
                    emulateJSON: true
                }).then(function(response) {
                    this.loading = false;
                    if (response.data.code == "0") {
                        this.pagecount = Number(response.body.count);
                        this.list = response.body.data;
                    } else {
                        this.$message.error(response.data.msg);
                    }
                }).catch(function(error) {
                    this.loading = false;
                    this.$message.error('网络请求失败');
                });
            },
            add: function() {
                if (!this.submitForm.user || !this.submitForm.pass) {
                    this.$message.error("账号和密码不能为空！");
                    return false;
                }

                if (this.needsLocation()) {
                    if (!this.submitForm.address || !this.submitForm.longitude || !this.submitForm.latitude) {
                        this.$message.error("该平台需要完整的地址和经纬度信息！");
                        return false;
                    }
                }

                if (!this.submitForm.worktime) {
                    this.$message.error("请设置打卡时间！");
                    return false;
                }

                if (this.submitForm.day < 1) {
                    this.$message.error("打卡天数必须大于0！");
                    return false;
                }

                this.submitLoad = true;
                this.submitForm.cid = this.add_cid;

                this.$http.post("/daka.php?act=add", { form: this.submitForm}, {
                    emulateJSON: true
                }).then(function(data) {
                    this.submitLoad = false;
                    if (data.body.code == 1) {
                        this.$message.success(data.body.msg);
                        this.dialog.order = false;
                        setTimeout(function(){window.location.reload()}, 800);
                    } else {
                        this.$message.error(data.body.msg);
                    }
                }).catch(function(error) {
                    this.submitLoad = false;
                    this.$message.error('网络请求失败，请检查网络连接');
                });
            },
            handleLog: function(id) {
                this.drawer.loading = true;
                this.dialog.log = true;
                this.drawer.list = [];

                this.$http.post("/daka.php?act=getOrderLogs", {
                    id: id
                }, {
                    emulateJSON: true
                }).then(function(data) {
                    this.drawer.loading = false;
                    if (Array.isArray(data.data)) {
                        this.drawer.list = data.data;
                    } else {
                        this.$message.error('日志格式错误');
                    }
                }).catch(function(error) {
                    this.drawer.loading = false;
                    this.$message.error('获取日志失败');
                });
            },
            handleDays: function(id) {
                this.$prompt('请输入要续费的天数', '续费订单', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    inputPattern: /^[1-9]\d*$/,
                    inputErrorMessage: '请输入有效的天数（正整数）'
                }).then(({ value }) => {
                    const days = parseInt(value);
                    if (days > 0) {
                        this.renewOrder(id, days);
                    } else {
                        this.$message.error('请输入有效的天数');
                    }
                }).catch(() => {
                    this.$message.info('已取消续费');
                });
            },
            renewOrder: function(id, days) {
                this.loading = true;
                this.$http.post("/daka.php?act=renewOrder", {
                    id: id,
                    days: days
                }, {
                    emulateJSON: true
                }).then(function(data) {
                    this.loading = false;
                    if (data.data.code == "1") {
                        this.$message.success(data.data.msg || '续费成功！');
                        this.query();
                    } else {
                        this.$message.error(data.data.msg || '续费失败');
                    }
                }).catch(function(error) {
                    this.loading = false;
                    this.$message.error('网络错误，请重试');
                });
            },
            save: function() {
                if (!this.submitForm.user || !this.submitForm.pass) {
                    this.$message.error("账号和密码不能为空！");
                    return false;
                }

                if (this.needsLocation()) {
                    if (!this.submitForm.address || !this.submitForm.longitude || !this.submitForm.latitude) {
                        this.$message.error("保证必填项不能为空！");
                        return false;
                    }
                }

                this.submitLoad = true;

                this.$http.post("/daka.php?act=save", { form: this.submitForm}, {
                    emulateJSON: true
                }).then(function(data) {
                    this.submitLoad = false;
                    if (data.data.code == 1) {
                        this.$message.success(data.data.msg);
                        this.dialog.order = false;
                        this.query();
                    } else {
                        this.$message.error(data.data.msg);
                    }
                }).catch(function(error) {
                    this.submitLoad = false;
                    this.$message.error('网络错误，请重试');
                });
            },
            isOrderExpired: function(endtime) {
                if (!endtime) return false;
                const now = new Date();
                const end = new Date(endtime);
                return end < now;
            },
            handleStatus: function(id, newStatus) {
                const orderIndex = this.list.findIndex(o => o.id === id);
                const order = this.list[orderIndex];

                if (order && this.isOrderExpired(order.endtime) && newStatus === 1) {
                    this.$message.error("订单已过期,无法开启。请先续费后再操作。");
                    this.$set(this.list[orderIndex], 'status', 0);
                    return;
                }

                const loadingInstance = this.$loading({
                    lock: true,
                    text: '正在更新状态...',
                    spinner: 'el-icon-loading'
                });

                this.$http.post("/daka.php?act=save", {
                    form: {
                        id: id,
                        status: newStatus
                    }
                }, {
                    emulateJSON: true
                }).then(function(data) {
                    loadingInstance.close();
                    if (data.body.code == 1) {
                        this.$message.success("状态更新成功");
                        this.query();
                    } else {
                        this.$message.error(data.body.msg || '状态更新失败');
                        this.$set(this.list[orderIndex], 'status', newStatus === 1 ? 0 : 1);
                    }
                }).catch(function(error) {
                    loadingInstance.close();
                    this.$message.error('网络错误,请重试');
                    this.$set(this.list[orderIndex], 'status', newStatus === 1 ? 0 : 1);
                });
            },
            handleClock: function(id) {
                this.$confirm('确定要立即执行打卡吗？', '立即打卡', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    const loading = this.$loading({
                        lock: true,
                        text: '正在提交打卡任务...',
                        spinner: 'el-icon-loading'
                    });

                    this.$http.post("/daka.php?act=manualClock", {
                        id: id
                    }, {
                        emulateJSON: true
                    }).then(function(data) {
                        loading.close();
                        if (data.body.code == 1) {
                            this.$message.success(data.body.msg);
                        } else {
                            this.$message.error(data.body.msg);
                        }
                    }).catch(function(error) {
                        loading.close();
                        this.$message.error('网络错误，请重试');
                    });
                }).catch(() => {
                    this.$message.info('已取消打卡');
                });
            },
            handleDelete: function(row) {
                this.$confirm('确定要删除该订单吗?删除后将无法恢复!', '删除确认', {
                    confirmButtonText: '确定删除',
                    cancelButtonText: '取消',
                    type: 'warning',
                    center: true
                }).then(() => {
                    const loading = this.$loading({
                        lock: true,
                        text: '正在删除订单...',
                        spinner: 'el-icon-loading'
                    });

                    this.$http.post("/daka.php?act=deleteOrder", {
                        id: row.id
                    }, {
                        emulateJSON: true
                    }).then(function(data) {
                        loading.close();
                        if (data.body.code == 1) {
                            this.$message.success(data.body.msg);
                            this.query();
                        } else {
                            this.$message.error(data.body.msg);
                        }
                    }).catch(function(error) {
                        loading.close();
                        this.$message.error('删除失败,请检查网络连接');
                    });
                }).catch(() => {
                    this.$message.info('已取消删除');
                });
            },
            handlePatchReport: function(row) {
                this.report = {
                    type: 'day',
                    start: '',
                    end: '',
                    loading: false
                };

                this.currentPatchOrderId = row.id;

                this.dialog.reset = true;
            },
            patchReport: function() {
                if (!this.report.type) {
                    this.$message.error('请选择报告类型');
                    return;
                }

                if (!this.report.start || !this.report.end) {
                    this.$message.error('请选择开始和结束日期');
                    return;
                }

                if (this.report.start > this.report.end) {
                    this.$message.error('开始日期不能大于结束日期');
                    return;
                }

                this.$confirm(
                    '确定要执行补报告操作吗？系统将自动计算费用并扣除',
                    '确认补签',
                    {
                        confirmButtonText: '确定',
                        cancelButtonText: '取消',
                        type: 'warning'
                    }
                ).then(() => {
                    this.report.loading = true;

                    this.$http.post("/daka.php?act=patchReport", {
                        id: this.currentPatchOrderId,
                        startDate: this.report.start,
                        endDate: this.report.end,
                        type: this.report.type
                    }, { emulateJSON: true }).then(function(data) {
                        this.report.loading = false;

                        if (data.body.code == 1) {
                            this.$message.success(data.body.msg);
                            this.dialog.reset = false;

                            this.report.type = '';
                            this.report.start = '';
                            this.report.end = '';

                            this.query();
                        } else {
                            this.$message.error(data.body.msg);
                        }
                    }).catch(function(error) {
                        this.report.loading = false;
                        this.$message.error('补报告提交失败，请检查网络连接');
                    });
                }).catch(() => {
                    this.$message.info('已取消补签操作');
                });
            },
            showmsg: function(data) {
                if (this.loading) this.loading = false;
                if (data.code == 1) {
                    this.query();
                    this.$message.success(data.msg);
                } else this.$message.error(data.msg);
            },
            checkVersion: function() {
                var that = this;
                this.$http.get("/daka.php?act=checkVersion").then(function(res) {
                    if (res.body.code == 1) {
                        that.remoteVersion = res.body.data.version;
                        that.updateInfo = res.body.data;

                        if (that.compareVersion(that.remoteVersion, that.localVersion) > 0) {
                            that.hasNewVersion = true;
                            that.versionStatusClass = 'status-new';
                            that.versionStatusText = '发现新版本 v' + that.remoteVersion + ' 点击更新';
                        } else {
                            that.versionStatusClass = 'status-latest';
                            that.versionStatusText = '已是最新版本';
                        }
                    } else {
                        that.versionStatusClass = 'status-checking';
                        that.versionStatusText = '检测失败';
                    }
                }).catch(function() {
                    that.versionStatusClass = 'status-checking';
                    that.versionStatusText = '检测失败';
                });
            },
            compareVersion: function(v1, v2) {
                var arr1 = v1.toString().split('.');
                var arr2 = v2.toString().split('.');
                var len = Math.max(arr1.length, arr2.length);
                for (var i = 0; i < len; i++) {
                    var num1 = parseInt(arr1[i] || 0);
                    var num2 = parseInt(arr2[i] || 0);
                    if (num1 > num2) return 1;
                    if (num1 < num2) return -1;
                }
                return 0;
            },
            doUpdate: function() {
                var that = this;
                this.updateLoading = true;

                this.$http.get("/daka.php?act=doUpdate&file=daka").then(function(res) {
                    if (res.body.code == 1) {
                        return that.$http.get("/daka.php?act=doUpdate&file=yfdk");
                    } else {
                        throw new Error(res.body.msg);
                    }
                }).then(function(res) {
                    that.updateLoading = false;
                    if (res.body.code == 1) {
                        that.$message.success('更新成功！2秒后刷新页面...');
                        that.showUpdateDialog = false;
                        setTimeout(function() { location.reload(true); }, 2000);
                    } else {
                        throw new Error(res.body.msg);
                    }
                }).catch(function(err) {
                    that.updateLoading = false;
                    that.$message.error('更新失败：' + (err.message || '未知错误'));
                });
            }
        },
        watch: {
            'submitForm.day': function(newVal) {
                if (this.dialog.title === '添加' && this.add_cid && newVal > 0) {
                    this.$http.post("/daka.php?act=getmoney", {
                        cid: this.add_cid,
                        day: newVal
                    }, { emulateJSON: true }).then(function(data) {
                        if (data.body.code == 1) {
                            this.submitForm.yuji = data.body.msg;
                        }
                    });
                }
            }
        },
        mounted() {
            this.getProjects();
            this.query();
            this.loadPopupNotice();
            this.getSourceBalance();
            <?php if ($userrow['uid'] == 1): ?>
            this.checkVersion();
            <?php endif; ?>
        }
    });
</script>
<?php include('foot.php'); ?>