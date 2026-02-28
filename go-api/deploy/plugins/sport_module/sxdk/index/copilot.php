<?php
$title = 'copilot';
require_once('head.php');
$addsalt = md5(mt_rand(0, 999) . time());
$_SESSION['addsalt'] = $addsalt;

?>
<!-- 引入样式 -->

<link rel="stylesheet" href="../sxdk/element/index.css">
<!--<script src="https://cdn.bootcdn.net/ajax/libs/jquery/1.9.1/jquery.js"></script>-->
<!--<script src="https://cdn.bootcdn.net/ajax/libs/vue/2.6.1/vue.js"></script>-->
<!--<script src="https://cdn.staticfile.org/vue-resource/1.5.1/vue-resource.min.js"></script>-->
<script src="../sxdk/jquery/jquery.js"></script>
<script src="../sxdk/vue/vue.js"></script>
<script src="../sxdk/vue/vue-resource.min.js"></script>
<script src="../sxdk/element/index.js"></script>


<link rel="stylesheet" href="//at.alicdn.com/t/c/font_3807157_95r8c8ifkoo.css" />
<div id="content" class="lyear-layout-content" role="main" style="padding-left: 5px; padding-top: 20px">
	<div class="app-content-body">
		<div class="wrapper-md control" id="app">
			<!-- 下单弹窗-->
			<el-dialog class="addeditDialog" :title="edit?'修改订单':'添加订单'" :visible.sync="dialog"
				:before-close="handleClose" :width="addeditDialogWidth" :height="addeditDialogHeight" align-center>
				<el-form :model="form" :rules="addFormRules" ref="Form" label-width="auto" size="mini"
					label-position="left">
					<el-form-item label="平台" prop="platform">
						<el-select v-model="form.platform" placeholder="请选择" size="mini" :disabled="edit">
							<el-option v-for="item in platFormList" :key="item.value" :label="item.label"
								@click.native="get_price(item.value)" :value="item.value">
							</el-option>
						</el-select>
						<span style="color:#409eff;margin-left:10px;"> (单价：{{price}}/天)</span>
					</el-form-item>
					<el-form-item label="登录方式" v-if="form.platform=='xxt'" prop="loginType">
						<el-switch v-model="userSchool" active-text="手机号" inactive-text="学号" active-color="#13ce66"
							inactive-color="#13ce66" @change="xxtLoginTypeChange">
						</el-switch>
					</el-form-item>
					<el-form-item label="运行方式" v-if="form.platform=='xyb'" prop="runType">
						<el-radio-group v-model="form.runType">
                            <el-radio :label="1">抓包</el-radio>
                            <el-radio :label="2">账密1号</el-radio>
                            <el-radio :label="3">绑微1号</el-radio>
                        </el-radio-group>
					</el-form-item>
					<!-- 学习通 -->
					<el-form-item v-if="form.platform=='xxt'&&!userSchool" label="学校" prop="schoolId">
						<el-select v-model="form.schoolId" placeholder="请选择学校" size="mini"
							@change="(value)=>xxtschoolChange(value)" filterable remote
							:remote-method='xxtGetSchoolList' :loading="xxtgetSchoolLoading">
							<el-option v-for="item in xxtSchoolList" :key="item.value" :label="item.text"
								:value="item.value">
							</el-option>
						</el-select>
					</el-form-item>
					<!-- 习讯云 -->
					<el-form-item v-if="form.platform=='xxy'" label="学校" prop="schoolId">
						<el-select v-model="form.schoolId" placeholder="请选择学校" size="mini"
							@change="(value)=>xxyschoolChange(value)" filterable>
							<el-option v-for="item in xxySchoolList" :key="item.value" :label="item.text"
								:value="item.value">
							</el-option>
						</el-select>
					</el-form-item>
					<!-- 慧职教 -->
					<el-form-item v-if="form.platform=='hzj'" label="学校" prop="schoolId">
						<el-select v-model="form.schoolId" placeholder="请选择学校" size="mini"
							@change="(value)=>hzjschoolChange(value)" filterable>
							<el-option v-for="item in hzjSchoolList" :key="item.value" :label="item.text"
								:value="item.value">
							</el-option>
						</el-select>
					</el-form-item>
					<el-form-item label="账号" prop="phone">
						<el-input v-model="form.phone" size="mini" :disabled="edit"></el-input>
					</el-form-item>
					<el-form-item label="密码" prop="password">
						<el-input v-model="form.password" size="mini"></el-input>
					</el-form-item>
					<el-form-item label="项目关键词" prop="projectName" v-if="form.platform=='xyb'||form.platform=='gxy'||form.platform=='xxt'">
						<el-input v-model="form.projectName" size="mini"></el-input>
					</el-form-item>
					<el-form-item prop="autoInfo">
						<el-button @click="getPhoneInfo" size="mini" type="primary"
							:loading="getPhoneInfoLoading">获取打卡信息</el-button>
						<span style="color:#409eff;margin-left:10px;"> 别看表单多，输入账号密码点此按钮试试。</span>
					</el-form-item>
					<el-form-item label="姓名" prop="name">
						<el-input v-model="form.name" size="mini"></el-input>
					</el-form-item>
					<el-form-item label="岗位名称" prop="gwName" v-if="form.platform!='zxjy'">
						<el-input v-model="form.gwName" size="mini"></el-input>
					</el-form-item>
					<el-form-item label="岗位名称" prop="customizedGwName" v-if="form.platform=='zxjy'">
						<el-input v-model="form.customizedGwName" size="mini"></el-input>
					</el-form-item>
					<!-- 工学云/校友帮 -->
					<el-form-item label="国家" prop="country" v-if="form.platform=='gxy'||form.platform=='xyb'">
						<el-input v-model="form.country" size="mini"></el-input>
					</el-form-item>
					<!-- 工学云/校友帮/习讯云 -->
					<el-form-item label="省" prop="province"
						v-if="form.platform=='gxy'||form.platform=='xyb'||form.platform=='xxy'">
						<el-input v-model="form.province" size="mini"></el-input>
					</el-form-item>
					<!-- 工学云/校友帮/习讯云 -->
					<el-form-item label="市" prop="city"
						v-if="form.platform=='gxy'||form.platform=='xyb'||form.platform=='xxy'">
						<el-input v-model="form.city" size="mini"></el-input>
					</el-form-item>
					<!-- 工学云 -->
					<el-form-item label="区/县" prop="area" v-if="form.platform=='gxy'">
						<el-input v-model="form.area" size="mini"></el-input>
					</el-form-item>
					<!-- 校友帮 -->
					<el-form-item label="地区编码" prop="adcode" v-if="form.platform=='xyb'">
						<el-input v-model="form.adcode" size="mini"></el-input>
					</el-form-item>
					<!-- 职校家园/校友帮/学习通 -->
					<el-form-item label="公司地址" prop="addressOld"
						v-if="form.platform=='zxjy'||form.platform=='xyb'||form.platform=='xxy'||form.platform=='xxt'||form.platform=='hzj'">
						<el-input v-model="form.addressOld" size="mini"></el-input>
					</el-form-item>
					<!-- 黔直通 -->
					<el-form-item label="公司地址" prop="officialAddress" v-if="form.platform=='qzt'">
						<el-input v-model="form.officialAddress" size="mini"></el-input>
					</el-form-item>
					<!-- 工学云 -->
					<el-form-item label="公司地址" prop="jobAddress" v-if="form.platform=='gxy'">
						<el-input v-model="form.jobAddress" size="mini"></el-input>
					</el-form-item>
					<el-form-item label="打卡地址" prop="address" v-if="form.platform!='xxy'">
						<el-input v-model="form.address" size="mini"></el-input>
					</el-form-item>
					<el-form-item label="打卡地址" prop="address" v-if="form.platform=='xxy'">
						<el-select v-model="form.address" placeholder="请选择打卡地点" size="mini" filterable
							:remote-method="xxyAddressSearchPoi" :remote="xxyRemote.status" :loading="xxyRemote.loading"
							@change="(value)=>xxyCheckAddressChange(value)">
							<el-option v-for="item in xxyAddressPois" :key="item.textValue" :label="item.textValue"
								:value="item.value">
							</el-option>
						</el-select>
						<br>
						<el-switch v-model="xxyRemote.status" :width="60" inline-prompt active-text="在线搜索"
							inactive-text="离线搜索" />
					</el-form-item>
					<el-form-item label="经度" prop="lng">
						<el-input v-model="form.lng" size="mini"></el-input>
					</el-form-item>
					<el-form-item label="纬度" prop="lat">
						<el-input v-model="form.lat" size="mini"></el-input>
					</el-form-item>
					<!-- 职校家园 -->
					<el-form-item label="打卡时间" prop="check_time" v-if="form.platform=='zxjy'">
						<el-time-picker placeholder="选择时间" v-model="form.check_time" value-format="HH:mm:ss"
							size="mini"></el-time-picker>
					</el-form-item>
					<el-form-item label="上班打卡时间" prop="up_check_time" v-if="form.platform!='zxjy'">
						<el-time-picker placeholder="选择时间" v-model="form.up_check_time" value-format="HH:mm:ss"
							size="mini"></el-time-picker>
					</el-form-item>
					<el-form-item label="上班打卡类型" prop="up_remark" v-if="form.platform=='xxy'">
						<el-select v-model="form.up_remark" placeholder="请选择上班打卡类型" size="mini">
							<el-option v-for="item in mark_list" :key="item.key" :label="item.value" :value="item.key">
							</el-option>
						</el-select>
					</el-form-item>
					<el-form-item label="下班打卡" v-if="form.platform=='xxy'||form.platform=='hzj'" prop="down_check">
						<el-switch v-model="formOther.down_check" @change="downCheckChange" size="mini"
							active-color="#13ce66" inactive-color="#DCDFE6">
						</el-switch>
					</el-form-item>
					<el-form-item label="下班打卡时间" prop="down_check_time"
						v-if="(form.platform=='qzt' || form.platform=='gxy' || formOther.down_check)&&form.platform!='zxjy'">
						<el-time-picker placeholder="选择时间" v-model="form.down_check_time" value-format="HH:mm:ss"
							size="mini"></el-time-picker>
					</el-form-item>
					<el-form-item label="下班打卡类型" prop="down_remark" v-if="form.platform=='xxy'&&formOther.down_check">
						<el-select v-model="form.down_remark" placeholder="请选择下班打卡类型" size="mini">
							<el-option v-for="item in mark_list" :key="item.key" :label="item.value" :value="item.key">
							</el-option>
						</el-select>
					</el-form-item>
					<el-form-item label="打卡周期" prop="check_week">
						<el-checkbox-group v-model="form.check_week" size="mini">
							<el-checkbox label="0">周一</el-checkbox>
							<el-checkbox label="1">周二</el-checkbox>
							<el-checkbox label="2">周三</el-checkbox>
							<el-checkbox label="3">周四</el-checkbox>
							<el-checkbox label="4">周五</el-checkbox>
							<el-checkbox label="5">周六</el-checkbox>
							<el-checkbox label="6">周日</el-checkbox>
						</el-checkbox-group>
					</el-form-item>
					<el-form-item label="结束时间" prop="end_time">
						<el-date-picker type="date" placeholder="选择日期" v-model="form.end_time" value-format="yyyy-MM-dd"
							size="mini"></el-date-picker>
						<span style="color:#409eff;margin-left:10px;"> 日期当天是最后一天打卡</span>
					</el-form-item>
					<el-divider></el-divider>
					<el-form-item label="日报" prop="day_paper" v-if="form.platform!='qzt'">
						<el-switch v-model="form.day_paper" size="mini" active-color="#13ce66" inactive-color="#DCDFE6">
						</el-switch>
					</el-form-item>
					<el-form-item label="日报最小字数" prop="day_paper_minSize" v-if="form.platform!='qzt'&&form.day_paper">
						<el-input-number v-model="form.paperNumSetting.day.minSize" :step="100" step-strictly size="mini" 
							:min="0"></el-input-number>
						<span style="color:#409eff;margin-left:10px;"> 0表示不限制</span>
					</el-form-item>
					<el-form-item label="日报最大字数" prop="day_paper_maxSize" v-if="form.platform!='qzt'&&form.day_paper">
					    <el-input-number v-model="form.paperNumSetting.day.maxSize" :step="100" step-strictly size="mini" 
							:min="0"></el-input-number>
						<span style="color:#409eff;margin-left:10px;"> 0表示不限制</span>
					</el-form-item>
					<el-divider></el-divider>
					<el-form-item label="周报" prop="week_paper">
						<el-switch v-model="form.week_paper" size="mini" active-color="#13ce66"
							inactive-color="#DCDFE6">
						</el-switch>
					</el-form-item>
					<!-- 职校家园/黔直通 -->
					<el-form-item label="周报提交时间" prop="weekPaperSubmitWeek" v-if="form.week_paper">
						<el-input-number v-model="form.weekPaperSubmitWeek" :step="1" step-strictly size="mini" :max="7"
							:min="1"></el-input-number>
						<span style="color:#409eff;margin-left:10px;"> 周一到周七</span>
					</el-form-item>
					<el-form-item label="周报最小字数" prop="week_paper_minSize" v-if="form.platform!='qzt'&&form.week_paper">
						<el-input-number v-model="form.paperNumSetting.week.minSize" :step="100" step-strictly size="mini" 
							:min="0"></el-input-number>
						<span style="color:#409eff;margin-left:10px;"> 0表示不限制</span>
					</el-form-item>
					<el-form-item label="周报最大字数" prop="week_paper_maxSize" v-if="form.platform!='qzt'&&form.week_paper">
					    <el-input-number v-model="form.paperNumSetting.week.maxSize" :step="100" step-strictly size="mini" 
							:min="0"></el-input-number>
						<span style="color:#409eff;margin-left:10px;"> 0表示不限制</span>
					</el-form-item>
					<el-divider></el-divider>
					<el-form-item label="月报" prop="month_paper">
						<el-switch v-model="form.month_paper" size="mini" active-color="#13ce66"
							inactive-color="#DCDFE6">
						</el-switch>
					</el-form-item>
					<!-- 职校家园/黔直通 -->
					<el-form-item label="月报提交时间" prop="monthPaperSubmitMonth" v-if="form.month_paper">
						<el-input-number v-model="form.monthPaperSubmitMonth" :step="1" step-strictly size="mini"
							:max="28" :min="0"></el-input-number>
						<span style="color:#409eff;margin-left:10px;"> 0 代表为月底最后一天提交</span>
					</el-form-item>
					<el-form-item label="月报最小字数" prop="month_paper_minSize" v-if="form.platform!='qzt'&&form.month_paper">
						<el-input-number v-model="form.paperNumSetting.month.minSize" :step="100" step-strictly size="mini" 
							:min="0"></el-input-number>
						<span style="color:#409eff;margin-left:10px;"> 0表示不限制</span>
					</el-form-item>
					<el-form-item label="月报最大字数" prop="month_paper_maxSize" v-if="form.platform!='qzt'&&form.month_paper">
					    <el-input-number v-model="form.paperNumSetting.month.maxSize" :step="100" step-strictly size="mini" 
							:min="0"></el-input-number>
						<span style="color:#409eff;margin-left:10px;"> 0表示不限制</span>
					</el-form-item>
					<el-divider></el-divider>
					<el-form-item label="遵循法定节假日" prop="holiday">
						<el-switch v-model="form.holiday" size="mini" active-color="#13ce66" inactive-color="#DCDFE6">
						</el-switch>
					</el-form-item>
					<el-form-item label="预计扣费" prop="yuji">
						<el-input v-model="yuji" readonly size="mini" v-if="!edit"></el-input>
						<el-input v-model="editYuji" readonly size="mini" v-else></el-input>
					</el-form-item>
				</el-form>
				<div slot="footer" class="dialog-footer">
					<el-button @click="dialog = false">取 消</el-button>
					<el-button type="primary" @click="add" :loading="addloading">{{edit?'修改':'添加'}}</el-button>
				</div>
			</el-dialog>
			<el-dialog class="addeditDialog" :title="buPapersForm.phone+'补报告'" :visible.sync="bDialog"
				:before-close="BhandleClose" :width="addeditDialogWidth" :height="addeditDialogHeight" align-center>
				<el-form :model="buPapersForm" :rules="buPapersFormRules" ref="BForm" label-width="auto" size="mini"
					label-position="left">

					<el-form-item label="类型" prop="levelName">
                      <el-radio-group v-model="buPapersForm.levelName" size="mini">
                        <el-radio label="日报" v-if="buPapersForm.platform=='qzt'?false:true">日报</el-radio>
                        <el-radio label="周报">周报</el-radio>
                        <el-radio label="月报">月报</el-radio>
                      </el-radio-group>
                     </el-form-item>

					<el-form-item label="开始时间" prop="startTime">
							<el-date-picker type="date" placeholder="选择日期" v-model="buPapersForm.startTime" value-format="yyyy-MM-dd"
							size="mini"></el-date-picker>
					</el-form-item>


					<el-form-item label="结束时间" prop="startTime">
						<el-date-picker type="date" placeholder="选择日期" v-model="buPapersForm.endTime" value-format="yyyy-MM-dd"
							size="mini"></el-date-picker>
					</el-form-item>
				</el-form>
				<div slot="footer" class="dialog-footer">
					<el-button @click="bDialog = false">取 消</el-button>
					<el-button type="primary" @click="buPapers" :loading="buPapersloading">下单</el-button>
				</div>
			</el-dialog>
			<el-dialog class="addeditDialog" :title="bindWxForm.phone+'升级绑微'" :visible.sync="bindWxDialog"
				:before-close="bindWxhandleClose" :width="addeditDialogWidth" :height="addeditDialogHeight" align-center>
				<div>
				    <p>
				        1.绑定微信将会一次性<b style="color:red;font-size:14px">扣除本平台余额{{price*1000}}</b>作为绑定升级费用，且永久不可退。</br>
                        2.绑定后每日单价将会提升至目前单价的5倍，所以需要补齐剩余天数的4倍费用，提交将会<b style="color:red;font-size:14px">扣除下方预计补充费用</b>。
				    </p>
				    <p>
				        预计补充费用<el-input v-model="bindWxMoney" readonly size="mini"></el-input>
				    </p>
				</div>
				<div slot="footer" class="dialog-footer">
					<el-button @click="bindWxDialog = false">取 消</el-button>
					<el-button type="primary" @click="xybBindWx" :loading="bindWxLoading">升级绑微</el-button>
				</div>
			</el-dialog>
			<!-- wxpush -->
			<el-dialog class="addeditDialog" :title="form.phone+'绑定微信通知'" :visible.sync="wxpushDialog"
				:before-close="wxpushhandleClose" :width="addeditDialogWidth" :height="addeditDialogHeight" align-center>
				<div>
				    <el-image :src='formOther.wxpush_img' style='width:250px;height:250px' >
				        <div slot="error" class="image-slot">
                            <i class="el-icon-loading" style="font-size:64px"></i>
                        </div>
				    </el-image>
		            <p style="margin: 12px; font-size: 12px; color:red;">
            			上方二维码是wxpusher推送，但现在，我们更推荐showdoc推送，请关注showdoc微信公众号，并将公众号发送您的专属推送地址填入下方，最后点击使用showdoc按钮
            		</p>
        			地址<el-input v-model="form.wxpush" size="mini" placeholder="填写您的showdoc专属推送地址"></el-input>
				</div>
				<div slot="footer" class="dialog-footer">
					<el-button @click="wxpushDialog = false">取 消</el-button>
					<el-button type="primary" @click="useShowDoc" :loading="wxpushLoading">使用showdoc</el-button>
				</div>
			</el-dialog>
			<!-- 列表 -->
			<el-card class="box-card">
				<el-alert :size="allSize" title="TaiShan版本号：251225，欢迎下单，问题请反馈" type="success"></el-alert>
				<div style="height: 10px"></div>
				<?php if ($userrow['uid'] == 1) { ?>
					<el-button type="warning" :size="allSize" @click="update" :loading="updateLoading"
						plain>拉取进度</el-button>
				<?php } ?>
				<el-button type="warning" @click="()=>{dialog=true;edit=false;}" :size="allSize" plain>提交订单</el-button>
				<el-button type="warning" @click="picUpload" :size="allSize" plain>图片上传</el-button>
				<?php if ($userrow['uid'] == 1) { ?>
					<span>{{userrow.msg}}</span>
				<?php } ?>
				<div style="height: 20px"></div>
				<el-input placeholder="模糊查询" v-model="cx.qq" class="input-with-select" :size="allSize">
					<el-select v-model="cx.search" style="width: 80px" placeholder="条件" slot="prepend">
						<el-option label="账号" value="phone"></el-option>
						<el-option label="密码" value="password"></el-option>
					</el-select>
					<el-button slot="append" icon="el-icon-search" @click="query">搜索</el-button>
				</el-input>
				<div style="height: 20px" v-show="multipleSelection.length>0"></div>
				<div class="batch" v-show="multipleSelection.length>0">
					<span style="margin-right:10px;">已选择{{multipleSelection.length}}条</span>
					<el-select v-model="batchType" placeholder="请选择需要批量的操作" :size="allSize">
						<el-option v-for="item in batchList" :key="item.value" :label="item.label" :value="item.value">
						</el-option>
					</el-select>
					<el-button type="warning" @click="batchFuntion" :size="allSize" plain
						:loading="batchLoading">执行</el-button>
				</div>
				<div style="height: 10px"></div>
				<div class="card-body" style="padding:0px">
					<el-table :data="tableData" :row-class-name="tableRowClassName" v-loading="loading"
						element-loading-text="载入数据中" ref="multipleTable" :size="allSize" empty-text="啊哦！一条订单都没有哦！"
						highlight-current-row border style="font-size:12px" @selection-change="handleSelectionChange" style="font-size:12px">
						<el-table-column type="selection" width="55">
						</el-table-column>
						<el-table-column type="index" label="序号">
						</el-table-column>
						<el-table-column label="操作" fixed="left" width="50">
							<template slot-scope="scope">
								<el-dropdown @command="handleCommand">
									<span class="el-dropdown-link" style="color:#409EFF">
										更多<i class="el-icon-arrow-down el-icon--right"></i>
									</span>
									<el-dropdown-menu slot="dropdown">
									    <el-dropdown-item v-if="scope.row.platform=='xyb'&&scope.row.runType==2"
											:command="{item:scope.row,type:'bingWx'}">升级绑微</el-dropdown-item>
										<el-dropdown-item
											:command="{item:scope.row,type:'nowCheck'}">立即打卡</el-dropdown-item>
											<el-dropdown-item :command="{item:scope.row,type:'log'}">查看日志</el-dropdown-item>
											<el-dropdown-item
											:command="{item:scope.row,type:'buPapers'}">补报告</el-dropdown-item>
												<el-dropdown-item :command="{item:scope.row,type:'getAsyncTask'}">查看补报告进度</el-dropdown-item>
										
										
										<el-dropdown-item
											:command="{item:scope.row,type:'changeCheckCode'}">{{scope.row.code==1?"暂停":"启动"}}</el-dropdown-item>
										<el-dropdown-item
											:command="{item:scope.row,type:'showErWeiMa'}">微信推送</el-dropdown-item>
										<el-dropdown-item
											:command="{item:scope.row,type:'editOrderInfo'}">编辑</el-dropdown-item>
										<el-dropdown-item :command="{item:scope.row,type:'del'}">删除</el-dropdown-item>
									</el-dropdown-menu>
								</el-dropdown>
							</template>
						</el-table-column>
						<el-table-column label="平台">
							<template slot-scope="scope">
								{{
								scope.row.platform=="zxjy"?"职校家园":
								scope.row.platform=="qzt"?"黔职通":
								scope.row.platform=="xyb"?"校友帮":
								scope.row.platform=="xxy"?"习讯云&宁夏":
								scope.row.platform=="xxt"?"学习通":
								scope.row.platform=="gxy"?"工学云":
								scope.row.platform=="hzj"?"慧职教":"未知"
								}}
								{{
								scope.row.platform=="xyb"?scope.row.runType==1?"(抓包)":scope.row.runType==2?"(账密1号)":scope.row.runType==3?"(绑微1号)":"":""
								}}
							</template>
						</el-table-column>
						<el-table-column prop="phone" label="账号" width="150">
						</el-table-column>
						<el-table-column prop="password" label="密码" width="150">
						</el-table-column>
						<el-table-column label="状态" show-overflow-tooltip>
							<template slot-scope="scope">
								<el-tag type="primary" v-if="scope.row.code==1">
									运行中
								</el-tag>
								<el-tag type="warning" v-else> 未运行 </el-tag>
							</template>
						</el-table-column>
						<el-table-column prop="name" label="姓名" width="80">
						</el-table-column>
						<el-table-column prop="address" label="打卡地址" width="250">
						</el-table-column>
						<el-table-column label="微信推送" show-overflow-tooltip>
							<template slot-scope="scope">
								<el-tag type="primary" v-if="scope.row.wxpush"> 已开启 </el-tag>
								<el-tag type="warning" v-else> 未开启 </el-tag>
							</template>
						</el-table-column>
						<el-table-column label="上班打卡时间" width="120">
							<template slot-scope="scope">
								{{scope.row.check_time||scope.row.up_check_time}}
							</template>
						</el-table-column>
						<el-table-column prop="down_check_time" label="下班打卡时间" width="120">
						</el-table-column>
						<el-table-column label="打卡周期" width="180">
							<template slot-scope="scope">
								<el-tag type="success" v-if='scope.row.check_week.indexOf("0") > -1'>周一</el-tag>
								<el-tag type="success" v-if='scope.row.check_week.indexOf("1") > -1'>周二</el-tag>
								<el-tag type="success" v-if='scope.row.check_week.indexOf("2") > -1'>周三</el-tag>
								<el-tag type="success" v-if='scope.row.check_week.indexOf("3") > -1'>周四</el-tag>
								<el-tag type="success" v-if='scope.row.check_week.indexOf("4") > -1'>周五</el-tag>
								<el-tag type="success" v-if='scope.row.check_week.indexOf("5") > -1'>周六</el-tag>
								<el-tag type="success" v-if='scope.row.check_week.indexOf("6") > -1'>周日</el-tag>
							</template>
						</el-table-column>
						<el-table-column label="报告" width="180">
							<template slot-scope="scope">
								<el-tag type="success" v-if='scope.row.day_paper==1'>日报</el-tag>
								<el-tag type="success" v-if='scope.row.week_paper==1'>周报</el-tag>
								<el-tag type="success" v-if='scope.row.month_paper==1'>月报</el-tag>
							</template>
						</el-table-column>
						<el-table-column prop="end_time" label="到期时间" width="120">
						</el-table-column>
						<el-table-column prop="createTime" label="创建时间" width="150">
						</el-table-column>
					</el-table>
					<el-divider></el-divider><!--<by TaoYao 分页>-->
					<el-pagination :size="allSize" @size-change="sizechange" @current-change="pagechange"
						:current-page.sync="currentpage" :page-sizes="[10, 20, 50, 100, 200, 500]" :page-size="pagesize"
						layout="total,sizes, prev, pager, next, jumper" :total="pagecount" v-if="allSize=='small'">
					</el-pagination>
					<el-pagination v-else :size="allSize" @size-change="sizechange" @current-change="pagechange"
						:current-page.sync="currentpage" :page-sizes="[10, 20, 50, 100, 200, 500]" :page-size="pagesize"
						layout="total,sizes, prev, pager, next" :total="pagecount" small>
					</el-pagination>
					<el-divider></el-divider><!--<by TaoYao 分页>-->
				</div>
			</el-card>
		</div>
	</div>
	<script>
		var vm = new Vue({
			el: "#app",
			data() {
				return {
					addeditDialogWidth: "675px",
					addeditDialogHeight: "600px",
					allSize: "small",
					xxyRemote: {
						status: false,
						loading: false
					},
					loading: false,
					addloading: false,
					buPapersloading:false,
					bindWxLoading:false,
					wxpushLoading:false,
					getPhoneInfoLoading: false,
					updateLoading: false,
					xxtgetSchoolLoading: false,
					batchLoading: false,
					currentpage: 1, //默认在第几页
					pagesize: 20, //每页显示的数量
					pagecount: 20, //总数的默认值，后面会做调整，此数值无参考意义
					price: 0,
					dialog: false,
					bDialog: false,
					bindWxDialog: false,
					wxpushDialog:false,
					edit: false,
					userSchool: true,
					userrow:{
					    msg:"查询中...",
					},
					form: {
						id: 0, //id                          
						platform: "zxjy",  //平台                 
						runType:1,   //运行方式             校友帮
						school: "",//学校
						schoolName: "",//学校               慧职教
						schoolId: "",//学校id
						url: "",//学校私立服务器
						addressName: "",//习讯云地址名称
						up_remark: "0",//上班打卡类型
						down_remark: "8",//下班打卡类型
						phone: "",  //手机号                  
						password: "",//密码
						name: "",//姓名
						gwName: "",//岗位名称
						customizedGwName: "",
						phone_name: "HUAWEI|HUAWEIELE AL00|11",//手机型号            职校家园
						country: "中国",//国家              工学云 校友帮
						province: "",//省市                 工学云 校友帮
						city: "",//县                       工学云 校友帮
						area: "",//区县                     工学云
						adcode: "",//地区编码               校友帮
						addressOld: "",//公司地址            职校家园 校友帮 慧职教
						officialAddress: "",//公司地址      黔职通
						jobAddress: "",//公司地址           工学云
						address: "",//打卡地址
						lat: "",//经纬度
						lng: "",//经纬度
						reason: "",//打卡备注               校友帮
						projectName: "",//项目关键词         校友帮
						desctext: "",//打卡备注             工学云
						randomLocation: true,//浮动打卡     工学云
						check_time: "",//打卡时间            职校家园
						up_check_time: "",//上班打卡时间     
						down_check_time: "",//下班打卡时间
						check_week: [],//打卡周期
						end_time: "",//结束时间
						wxpush: "",//微信推送
						day_paper: false,//日报开关
						week_paper: false,//周报开关
						month_paper: false,//月报开关
						weekPaperSubmitWeek: 7,//周报提交时间     职校家园
						monthPaperSubmitMonth: 0,//月报提交时间   黔直通
						paper_router: 2,//报告库选择
						payType: 0,//一次新订单
						holiday: false,
						paperNumSetting: {
                    		day: {
                    			minSize: 0,
                    			maxSize: 0
                    		},
                    		week: {
                    			minSize: 0,
                    			maxSize: 0
                    		},
                    		month: {
                    			minSize: 0,
                    			maxSize: 0
                    		},
                    		summary:{
                    		    minSize: 0,
                    			maxSize: 0
                    		}
                    	}
					},
					buPapersForm:{
					    id:0,
					    phone:"",
					    startTime:"",
					    endTime:"",
					    levelName:"日报"
					},
					bindWxForm:{
					    id:0,
					    phone:"",
					    platform:"xyb",
					    check_week:[],
					    end_time:""
					},
					cx: { qq: "", search: "phone" },
					platFormList: [
						{
							value: "zxjy",
							label: "职校家园",
						},
						{
							value: "qzt",
							label: "黔职通",
						},
						{
							value: "gxy",
							label: "工学云",
						},
						{
							value: "xyb",
							label: "校友帮",
						},
						{
							value: "xxy",
							label: "习迅云&宁夏",
						},
						{
							value: "xxt",
							label: "学习通",
						},
						{
							value: "hzj",
							label: "慧职教",
						},
					],
					formOther: {
						down_check: false,
						wxpush_img:""
					},
					buPapersFormRules:{
					    levelName:[
					        { required: true, message: '请选择类型', trigger: 'blur' },
				        ],
				        startTime:[
				            { required: true, message: '请选择时间', trigger: 'change' }
				            ],
				            endTime:[
				            { required: true, message: '请选择时间', trigger: 'change' }
				            ]
					},
					addFormRules: {
						platform: [
							{ required: true, message: '请选择打卡平台', trigger: 'blur' },
						],
						schoolId: [
							{ required: true, message: '请选择学校', trigger: 'blur' }
						],
						phone: [
							{ required: true, message: '请输入打卡账号', trigger: 'blur' }
						],
						password: [
							{ required: true, message: '请输入打卡密码', trigger: 'blur' }
						],
						country: [
							{ required: true, message: '请输入国家', trigger: 'blur' }
						],
						province: [
							{ required: true, message: '请输入省份', trigger: 'blur' }
						],
						city: [
							{ required: true, message: '请输入区/县', trigger: 'blur' }
						],
						area: [
							{ required: true, message: '请输入区/县', trigger: 'blur' }
						],
						adcode: [
							{ required: true, message: '请输入地区编码', trigger: 'blur' }
						],
						addressOld: [
							{ required: true, message: '请输入公司地址', trigger: 'blur' }
						],
						officialAddress: [
							{ required: true, message: '请输入打卡地址', trigger: 'blur' }
						],
						jobAddress: [
							{ required: true, message: '请输入打卡地址', trigger: 'blur' }
						],
						address: [
							{ required: true, message: '请输入打卡地址', trigger: 'blur' }
						],
						lat: [
							{ required: true, message: '请输入纬度', trigger: 'blur' }
						],
						lng: [
							{ required: true, message: '请输入经度', trigger: 'blur' }
						],
						check_time: [
							{ required: true, message: '请选择时间', trigger: 'change' }
						],
						up_check_time: [
							{ required: true, message: '请选择时间', trigger: 'change' }
						],
						down_check_time: [
							{ required: true, message: '请选择时间', trigger: 'change' }
						],
						check_week: [
							{ type: 'array', required: true, message: '请至少选择一个打卡日期', trigger: 'change' }
						],
						end_time: [
							{ required: true, message: '请选择日期', trigger: 'change' }
						],
					},
					tableData: [],
					xxySchoolList: [],
					xxtSchoolList: [],
					hzjSchoolList: [],
					xxyAddressPois: [],
					mark_list: [
						{
							key: "0",
							value: "上班",
						},
						{
							key: "1",
							value: "因公外出",
						},
						{
							key: "2",
							value: "假期",
						},
						{
							key: "3",
							value: "请假",
						},
						{
							key: "4",
							value: "轮岗",
						},
						{
							key: "5",
							value: "回校",
						},
						{
							key: "6",
							value: "外宿",
						},
						{
							key: "7",
							value: "在家",
						},
						{
							key: "8",
							value: "下班",
						},
						{
							key: "9",
							value: "学习",
						},
						{
							key: "10",
							value: "毕业设计",
						},
						{
							key: "11",
							value: "院区轮转",
						},
						{
							key: "13",
							value: "集训",
						},
					],
					noticeList: [],
					multipleSelection: [],
					batchType: "",
					batchList: [
						{
							value: 'batchOpenOrder',
							label: '批量启动订单'
						}, {
							value: 'batchCloseOrder',
							label: '批量暂停订单'
						},
						{
							value: 'batchOpenHoliday',
							label: '批量绑定法定节假日'
						},
						{           
							value: 'batchCloseHoliday',
							label: '批量解绑法定节假日'
						},
						{           
							value: 'batchDelOrder',
							label: '批量删除订单'
						}

					],
				};
			},
			created() {
				this.handleResize();
			},
			beforeMount() { },
			mounted() {
			    this.get_userrow();
				this.get_price(this.form.platform);
				this.query();
				this.getNotice();
				this.xxyGetSchoolList()
				this.hzjGetSchoolList()
				this.onResize()
			},
			computed: {
				yuji: function () {
					if (this.form.end_time && this.form.check_week.length > 0) {
						let day = this.timeCalcTrueday(new Date().getTime(), this.form.end_time, this.form.check_week)
						if (this.form.platform=="xyb"&&this.form.runType==3){
						    day=day*5
						}
						return (this.price * day).toFixed(2) + "元";
					} else {
						return "0元"
					}
				},
				editYuji: function () {
					if (this.form.end_time && this.form.old_end_time && this.edit && this.form.check_week.length > 0) {
						let oldsjc = new Date(this.form.old_end_time + " 23:59:59").getTime();//01-01
						let tadysjc = new Date().getTime();//02-27
						let endSjc = new Date(this.form.end_time + " 23:59:59").getTime();//12-27
						let day;
						if (tadysjc >= oldsjc && tadysjc < endSjc) {
							//订单已过期续费
							let newDay = this.timeCalcTrueday(new Date().getTime(), this.form.end_time, this.form.check_week)
							day = newDay;
						} else {
							//订单未过期
							let nowsjc = oldsjc;//01-01

							if (endSjc <= nowsjc && this.form.old_check_week.length <= this.form.check_week) {
								return "0元"
							}
							let oldDay = this.timeCalcTrueday(new Date().getTime(), this.form.old_end_time, this.form.old_check_week)
							let newDay = this.timeCalcTrueday(new Date().getTime(), this.form.end_time, this.form.check_week)
							day = newDay - oldDay;
						}
						if (day < 0) {
							return "0元"
						}
						if (this.form.platform=='xyb'){
						    if (this.form.runType==3){
						        day=day*5
						    }
						}
						return (this.price * day).toFixed(2) + "元";
					} else {
						return "0元"
					}
				},
				bindWxMoney: function () {
					if (this.bindWxForm.end_time && this.bindWxForm.check_week.length > 0) {
						let day = this.timeCalcTrueday(new Date().getTime(), this.bindWxForm.end_time, this.bindWxForm.check_week)
						return (this.price * day * 4).toFixed(2) + "元";
					} else {
						return "0元"
					}
				},
			},
			methods: {
				//计算输入日期距离当前日期的天数差-仅计算传入参数二数组包含的周几天数
				timeCalcTrueday(nowsjc, end_time, check_week) {
					if (check_week) {
						// 打卡周期0-6排序
						check_week.sort(function (a, b) {
							return a - b;
						});
						// 获取当前时间戳
						// let nowsjc = new Date().getTime();
						// 获取结束时间戳
						let endSjc = new Date(end_time + " 23:59:59").getTime();
						if(endSjc<nowsjc){
						    return 0
						}
						// 获取当前周几
						let nowWeekDay = this.getDayOfWeek(nowsjc);
						// 获取当前周周末时间戳
						let weekEndSjc = new Date(
							this.format(nowsjc + (6 - nowWeekDay) * 86400 * 1000) + " 23:59:59"
						).getTime();
						//获取本周大于等于今天周几的天数，且在打卡周期内的
						let nowWeekLast = check_week.filter((item) => {
							return Number(item) >= nowWeekDay;
						});
						
						
						//获取结束当天周几
						let endWeekDay = this.getDayOfWeek(endSjc);
						//判断结束时间戳是否小于等于本周周末时间戳
						if (endSjc <= weekEndSjc) {
							//结束时间在本周内
							//通过打卡周期内本周大于今天周几的数组，来获取结束日期前的周几数组
							let lastWeekLast = nowWeekLast.filter((item) => {
								return Number(item) <= endWeekDay;
							});
							//返回本周可打卡天数
							return lastWeekLast.length;
						} else {
							//结束时间不在本周内
							//获取结束周小于等于结束时间周几的天数，且在打卡周期内
							let endWeekLast = check_week.filter((item) => {
								return Number(item) <= endWeekDay;
							});
							//获取整周总时间戳
							let intSjc = endSjc - (endWeekDay + 1) * 86400 * 1000 - weekEndSjc;
							//返回本周打卡周期内天数+整周打卡周期内天数+结束周打卡周期内天数
							return (
								nowWeekLast.length +
								(intSjc / 7 / 86400 / 1000) * check_week.length +
								endWeekLast.length
							);
						}
					}
				},
				//获取指定时间戳是周几
				getDayOfWeek(timestamp) {
					let date = new Date(timestamp);
					let dayOfWeek = date.getDay();
					let weekdays = [7, 1, 2, 3, 4, 5, 6];
					return weekdays[dayOfWeek] - 1;
				},
				//将时间戳转换为年月日时间
				format(dataString) {
					//dataString是整数，否则要parseInt转换
					var time = new Date(dataString);
					var year = time.getFullYear();
					var month = time.getMonth() + 1;
					var day = time.getDate();
					return (
						year +
						"-" +
						(month < 10 ? "0" + month : month) +
						"-" +
						(day < 10 ? "0" + day : day)
					);
				},
				handleSelectionChange(val) {
					this.multipleSelection = val;
					console.log(this.multipleSelection)
				},
				async batchFuntion() {
					if (this.multipleSelection.length == 0) {
						this.$message.error("未选择任何订单");
						return
					}
					if (this.batchType == "") {
						this.$message.error("未选择需要执行的操作");
						return
					}
					this.batchLoading = true;
					const orders = JSON.parse(JSON.stringify(this.multipleSelection));
					for (const [index, item] of orders.entries()) {

						let url = ""
						let title = ""
						let result = "失败"
						if (this.batchType == "batchOpenOrder") {
							item["code"] = 1
							title = "批量启动订单"
							url = "/sxdk/api.php?act=changeCheckCode"
						}
						else if (this.batchType == "batchCloseOrder") {
							item["code"] = 0
							title = "批量暂停订单"
							url = "/sxdk/api.php?act=changeCheckCode"
						}
						else if (this.batchType == "batchOpenHoliday") {
							item["code"] = 1
							title = "批量绑定法定节假日"
							url = "/sxdk/api.php?act=changeHolidayCode"
						}
						else if (this.batchType == "batchCloseHoliday") {
							item["code"] = 0
							title = "批量解绑法定节假日"
							url = "/sxdk/api.php?act=changeHolidayCode"
						}
						else if (this.batchType == "batchDelOrder") {
							title = "批量删除订单"
							url = "/sxdk/api.php?act=del"
						}
						else {
							return
						}

						let resp = await this.$http.post(
							url,
							{
								...item
							},
							{ emulateJSON: true }
						)
						if (resp.body.code == 0) result = "成功"
						console.log(resp.body.code)
						this.$notify({
							title: title,
							message: `当前进度${index + 1}/${this.multipleSelection.length}\n账号：${item.phone}执行完毕,执行结果：${result},执行日志：${resp.body.msg}`,
							duration: 3000
						});
					}




					this.query();
					this.batchLoading = false;
					this.$alert(`当前进度${this.multipleSelection.length}/${this.multipleSelection.length}\n全部执行完毕`, `批量执行完毕`, {

					});

				},
				resetForm() {
					this.form = {
						id: 0, //id                          
						platform: "zxjy",  //平台                 
						runType:1,
						school: "",//学校
						schoolName: "",//学校               慧职教
						schoolId: "",//学校id
						url: "",//学校私立服务器
						addressName: "",//习讯云地址名称
						up_remark: "0",//上班打卡类型
						down_remark: "8",//下班打卡类型
						phone: "",  //手机号                  
						password: "",//密码
						name: "",//姓名
						gwName: "",//岗位名称
						customizedGwName: "",
						phone_name: "HUAWEI|HUAWEIELE AL00|11",//手机型号            职校家园
						country: "中国",//国家              工学云 校友帮
						province: "",//省市                 工学云 校友帮
						city: "",//县                       工学云 校友帮
						area: "",//区县                     工学云
						adcode: "",//地区编码               校友帮
						addressOld: "",//公司地址            职校家园 校友帮 慧职教
						officialAddress: "",//公司地址      黔职通
						jobAddress: "",//公司地址           工学云
						address: "",//打卡地址
						lat: "",//经纬度
						lng: "",//经纬度
						reason: "",//打卡备注               校友帮
						projectName: "",//项目关键词         校友帮
						desctext: "",//打卡备注             工学云
						randomLocation: true,//浮动打卡     工学云
						check_time: "",//打卡时间            职校家园
						up_check_time: "",//上班打卡时间     
						down_check_time: "",//下班打卡时间
						check_week: [],//打卡周期
						end_time: "",//结束时间
						wxpush: "",//微信推送
						day_paper: false,//日报开关
						week_paper: false,//周报开关
						month_paper: false,//月报开关
						weekPaperSubmitWeek: 7,//周报提交时间     职校家园
						monthPaperSubmitMonth: 0,//月报提交时间   黔直通
						paper_router: 2,//报告库选择
						payType: 0,//一次新订单
						holiday: false,
						paperNumSetting : {
            				day: {
            					minSize: 0,
            					maxSize: 0
            				},
            				week: {
            					minSize: 0,
            					maxSize: 0
            				},
            				month: {
            					minSize: 0,
            					maxSize: 0
            				},
            				summary: {
            					minSize: 0,
            					maxSize: 0
            				}
            			}
					},
						this.formOther.down_check = false
				},
				BresetForm(){
				    this.form={
					    id:0,
					    phone:"",
					    startTime:"",
					    endTime:"",
					    levelName:"日报"
					}
				},
				bingWxresetForm(){
				    this.bindWxForm={
					    id:0,
					    phone:"",
					    platform:"xyb",
					    check_week:[],
					    end_time:""
					}
				},
				picUpload() {
					this.$confirm(`支持平台：<br/>校友帮(支持：打卡、立即打卡、日报、周报、月报、补报告)<br/>黔职通(支持：打卡、立即打卡、周报、月报、补报告)<br/>习讯云&宁夏(打卡、立即打卡、日报、周报、月报)<br/>工学云/蘑菇钉(打卡、立即打卡、日报、周报、月报)<br/><br/>每使用一张删除一张，图片可上传储备。<p style='color:red'>图片数据库网址：http://location.copilotai.top:4000/#/picDatabases</p>`, '打卡图片上传', {
						distinguishCancelAndClose: true,
						dangerouslyUseHTMLString: true,
						confirmButtonText: '跳转',
						cancelButtonText: '取消'
					})
						.then(() => {
							window.open('http://location.copilotai.top:4000/#/picDatabases', "_blank");
						})
						.catch(action => { });
				},
				xxyschoolChange(value) {
					let checkSchools = this.xxySchoolList.filter((item) => {
						return item.value == value
					})
					if (checkSchools.length > 0) {
						checkSchool = checkSchools[0]
						this.form.school = checkSchool.text
						this.form.url = checkSchool.url
						this.form.schoolId = checkSchool.value
					} else {
						this.$message.error("不存在的学校");
						this.form.school = ""
						this.form.url = ""
						this.form.schoolId = ""
					}
				},
				xxyCheckAddressChange(value) {
					let checkAddresss = this.xxyAddressPois.filter((item) => {
						return item.value == value
					})
					if (checkAddresss.length > 0) {
						checkAddress = checkAddresss[0]
						this.form.address = checkAddress.value
						this.form.addressName = checkAddress.addressName
						this.form.lat = checkAddress.lat
						this.form.lng = checkAddress.lng
					} else {
						this.$message.error("不存在的地址");
					}
				},
				xxtschoolChange(value) {
					let checkSchools = this.xxtSchoolList.filter((item) => {
						return item.value == value
					})
					if (checkSchools.length > 0) {
						checkSchool = checkSchools[0]
						this.form.school = checkSchool.text
						this.form.schoolId = checkSchool.value
					} else {
						this.$message.error("不存在的学校");
						this.form.school = ""
						this.form.schoolId = ""
					}
				},
				xxtLoginTypeChange(value) {
					if (value) {
						this.form.schoolId = ""
						this.form.school = ""
					}
				},
				hzjschoolChange(value) {
					let checkSchools = this.hzjSchoolList.filter((item) => {
						return item.value == value
					})
					if (checkSchools.length > 0) {
						checkSchool = checkSchools[0]
						this.form.schoolName = checkSchool.text
						this.form.schoolId = checkSchool.value
					} else {
						this.$message.error("不存在的学校");
						this.form.schoolName = ""
						this.form.schoolId = ""
					}
				},
				handleCommand: async function (data) {
					if (data.type == "log") {
						this.getLog(data.item)
					}
					if (data.type == "buPapers") {
					    this.buPapersForm.id=data.item.id
					    this.buPapersForm.phone=data.item.phone
					    this.buPapersForm.platform=data.item.platform
					    if(this.buPapersForm.platform=="qzt"){
					        this.buPapersForm.levelName="周报"
					    }
						this.bDialog=true
					}
					if (data.type == "bingWx") {
					    let sourceOrder = await this.querySourceOrder(data.item)
					    if (sourceOrder) {
					        console.log(sourceOrder)
    					    this.bindWxForm.id=data.item.id
    					    this.bindWxForm.phone=sourceOrder.phone
    					    this.bindWxForm.platform=data.item.platform
    					    this.bindWxForm.end_time=sourceOrder.end_time
    					    this.bindWxForm.check_week=sourceOrder.check_week.split(",")
    					    this.get_price(data.item.platform)
    						this.bindWxDialog=true
					    }
					}
					
					if (data.type == "getAsyncTask") {
						this.getAsyncTask(data.item)
					}
					if (data.type == "del") {
						this.deleteOrder(data.item)
					}
					if (data.type == "nowCheck") {
						this.nowCheck(data.item)
					}
					if (data.type == "changeCheckCode") {
						this.changeCheckCode(data.item)
					}
					if (data.type == "showErWeiMa") {
						this.getWxPush(data.item)
						this.form.id=data.item.id
                        const regex = /^https:\/\/push\.showdoc\.com\.cn\/server\/api\/push\/.+$/;
                    	if (regex.test(data.item.wxpush)) {
                    		this.form.wxpush = data.item.wxpush;
                    	} else {
                    		this.form.wxpush=""
                    	}
						this.wxpushDialog=true
					}
					if (data.type == "editOrderInfo") {
                        this.get_price(data.item.platform)
						let sourceOrder = await this.querySourceOrder(data.item)
						if (sourceOrder) {
							this.form = {
								...sourceOrder,
								platform: data.item.platform,
							}
							this.form.id = data.item.id
							this.form.old_end_time = sourceOrder.end_time;
							this.form.old_check_week = sourceOrder.check_week.split(",");
							this.form.check_week = sourceOrder.check_week.split(",");
							this.form.holiday = sourceOrder.holiday == 1;
							this.form.day_paper = sourceOrder.day_paper == 1;
							this.form.week_paper = sourceOrder.week_paper == 1;
							this.form.month_paper = sourceOrder.month_paper == 1;
							this.form.weekPaperSubmitWeek = sourceOrder.weekPaperSubmitWeek + 1;
							if(sourceOrder.paperNumSetting){
							    try {
                        			const paperNumSetting = JSON.parse(sourceOrder.paperNumSetting);
                        			this.form.paperNumSetting = paperNumSetting;
                        		} catch (e) {
                        			this.form.paperNumSetting = {
                        				day: {
                        					minSize: 0,
                        					maxSize: 0
                        				},
                        				week: {
                        					minSize: 0,
                        					maxSize: 0
                        				},
                        				month: {
                        					minSize: 0,
                        					maxSize: 0
                        				},
                        				summary: {
                        					minSize: 0,
                        					maxSize: 0
                        				}
                        			};
                        		}
							}else{
							    this.form.paperNumSetting = {
                        				day: {
                        					minSize: 0,
                        					maxSize: 0
                        				},
                        				week: {
                        					minSize: 0,
                        					maxSize: 0
                        				},
                        				month: {
                        					minSize: 0,
                        					maxSize: 0
                        				},
                        				summary: {
                        					minSize: 0,
                        					maxSize: 0
                        				}
                        			};
							}
							
							if (sourceOrder.down_check_time) {
								this.formOther.down_check = true
							}
							this.edit = true;
							
							this.dialog = true;
							console.log(this.form)
						}
					}
				},
				getPhoneInfo: function () {
					this.getPhoneInfoLoading = true
					this.$http
						.post(
							"/sxdk/api.php?act=searchPhoneInfo",
							{
								...this.form
							},
							{ emulateJSON: true }
						)
						.then((res) => {
							if (res.data.code == 0) {
								if (this.form.platform == "zxjy") {
									this.form.name = res.data.data.name
									this.form.address = res.data.data.address
									this.form.gwName = res.data.data.gwName
									this.form.customizedGwName = res.data.data.gwName
									this.form.lat = res.data.data.lat
									this.form.lng = res.data.data.lng
									this.form.addressOld = res.data.data.addressOld
									this.$message.success(res.data.msg + "\n地址、经纬度自动配置成功");
								} else if (this.form.platform == "qzt") {
									this.form.name = res.data.data.name;
									this.form.up_check_time = res.data.data.checkInTime;
									this.form.down_check_time = res.data.data.checkOutTime;
									this.form.end_time = res.data.data.endTime;
									this.form.lng = res.data.data.longitude;
									this.form.lat = res.data.data.latitude;
									this.form.officialAddress = res.data.data.address;
									this.form.address = res.data.data.officialAddress;
									this.form.gwName = res.data.data.gwName;
									let weekList = res.data.data.weekList.split(",");
									weekList = weekList.map((item) => {
										if (Number(item) > 1) {
											return (Number(item) - 2).toString();
										} else if (Number(item) == 1) {
											return "6";
										}
									});
									this.form.check_week = weekList;
									this.$message.success(res.data.msg + "\n上下班打卡时间、实习结束时间、打卡周期、地址、经纬度自动配置成功");
								} else if (this.form.platform == "gxy") {
									this.form.name = res.data.data.name;
									this.form.gwName = res.data.data.jobName;
									this.form.lng = res.data.data.lng;
									this.form.lat = res.data.data.lat;
									this.form.jobAddress = res.data.data.jobAddress;
									this.form.address = res.data.data.address;
									this.form.province = res.data.data.jobProvince;
									this.form.city = res.data.data.jobCity;
									this.form.area = res.data.data.jobArea;
									this.form.projectName = res.data.data.planName;
									this.$message.success(res.data.msg + "\n地址、经纬度自动配置成功");
								} else if (this.form.platform == "xyb") {
									this.form.addressOld = res.data.data.addressOld;
									this.form.address = res.data.data.address;
									this.form.lng = res.data.data.lng;
									this.form.lat = res.data.data.lat;
									this.form.adcode = res.data.data.adcode;
									this.form.province = res.data.data.province;
									this.form.country = res.data.data.country;
									this.form.city = res.data.data.city;
									this.form.day_paper = res.data.data.checkOrder.needDailyBlogs;
									this.form.week_paper = res.data.data.checkOrder.needWeeklyBlogs;
									this.form.month_paper = res.data.data.checkOrder.needMonthlyBlogs;
									this.form.gwName = res.data.data.checkOrder.gwName;
									this.form.name = res.data.data.checkOrder.name;
									this.form.projectName = res.data.data.checkOrder.planName;
									if (res.data.data.checkOrder.clockRuleType == 1) {
										this.formOther.down_check = true;
										this.$confirm('官方信息获取成功，该账号需要进行下班打卡，请不要忘记设置。', '提示', {
											confirmButtonText: '确定',
											cancelButtonText: '取消',
											type: 'warning'
										}).then(() => {
											this.$message.success(res.data.msg + "\n日周月报、地址、经纬度自动配置成功");
										}).catch(() => {
											this.$message.success(res.data.msg + "\n日周月报、地址、经纬度自动配置成功");
										});
									} else {
										this.form.down_check_time = "";
										this.formOther.down_check = false;
										this.$confirm('官方信息获取成功，该账号无需进行下班打卡。', '提示', {
											confirmButtonText: '确定',
											cancelButtonText: '取消',
											type: 'warning'
										}).then(() => {
											this.$message.success(res.data.msg + "\n日周月报、地址、经纬度自动配置成功");
										}).catch(() => {
											this.$message.success(res.data.msg + "\n日周月报、地址、经纬度自动配置成功");
										});
									}
								} else if (this.form.platform == "xxy") {
									this.form.name = res.data.data.name;
									this.form.lng = res.data.data.lng;
									this.form.lat = res.data.data.lat;
									this.form.province = res.data.data.province;
									this.form.city = res.data.data.city;
									this.form.addressOld = res.data.data.addressOld;
									this.form.gwName = res.data.data.gwName;
									this.form.city = res.data.data.city;
									this.form.day_paper = res.data.data.day_paper;
									this.form.week_paper = res.data.data.week_paper;
									this.form.month_paper = res.data.data.month_paper;
									this.xxyAddressPois = res.data.data.addressPois.map((item) => {
										return {
											...item,
											textValue: item.value + item.text,
										};
									});
									if (res.data.data.addressPois.length > 0) {
										this.form.address = res.data.data.address;
										this.form.addressName = res.data.data.addressName;
									}
									this.$message.success(res.data.msg + "\n日周月报、地址、经纬度自动配置成功");
								} else if (this.form.platform == "xxt") {
									this.form.name = res.data.data.name;
									this.form.lng = res.data.data.lng;
									this.form.lat = res.data.data.lat;
									this.form.address = res.data.data.address;
									this.form.addressOld = res.data.data.addressOld;
									this.form.gwName = res.data.data.gwName;
									this.form.day_paper = res.data.data.day_paper;
									this.form.week_paper = res.data.data.week_paper;
									this.form.month_paper = res.data.data.month_paper;
									this.form.projectName = res.data.data.planName;
									if (res.data.data.up_check_time && res.data.data.down_check_time) {
										this.formOther.down_check = true;
										this.form.up_check_time = res.data.data.up_check_time;
										this.form.down_check_time = res.data.data.down_check_time;
										this.$message.success(res.data.msg + "\n日周月报、地址、经纬度,上下班打卡时间自动配置成功");
									} else {
										this.form.down_check_time = "";
										this.formOther.down_check = false;
										this.$message.success(res.data.msg + "\n日周月报、地址、经纬度自动配置成功，无需下班打卡");
									}
								} else if (this.form.platform == "hzj") {
									this.form.name = res.data.data.name;
									this.form.lng = res.data.data.lng;
									this.form.lat = res.data.data.lat;
									this.form.address = res.data.data.address;
									this.form.addressOld = res.data.data.addressOld;
									this.form.gwName = res.data.data.gwName;
									this.form.day_paper = res.data.data.day_paper == 1;
									this.form.week_paper = res.data.data.week_paper == 1;
									this.form.month_paper = res.data.data.month_paper == 1;
									this.$message.success(res.data.msg + "\n日周月报、地址、经纬度自动配置成功");
								} else {
									this.$message.error("未匹配到项目");
								}
							} else {
								this.$message.error(res.data.msg);
							}
							this.getPhoneInfoLoading = false
						}).catch((e) => {
							this.$message.error("网络错误，超时");
							this.getPhoneInfoLoading = false
						});
				},
				query: function () {
					this.loading = true;
					data = { cx: this.cx, page: this.currentpage, size: this.pagesize };
					this.$http
						.post("/sxdk/api.php?act=order", data, { emulateJSON: true })
						.then(function (data) {
							this.loading = false;
							if (data.data.code == "0") {
								this.pagecount = Number(data.body.count);
								this.tableData = data.body.data;
							} else {
								this.$message.error(data.data.msg);
							}
						});
				},
				get_price(platform) {
					if (this.form.platform == "xyb" || this.form.platform == "zxjy" || this.form.platform == "xxy" || this.form.platform == "xxt" || this.form.platform == "hzj") {
						this.formOther.down_check = false
						this.form.down_check_time = ""
					} else {
						this.formOther.down_check = true
						this.form.down_check_time = ""
					}
					this.$http
						.post(
							"/sxdk/api.php?act=price",
							{ platform: platform },
							{ emulateJSON: true }
						)
						.then(function (data) {
							this.price = data.data.data;
						});
				},
				get_userrow(){
				    this.$http
						.post(
							"/sxdk/api.php?act=get_userrow",
							{ },
							{ emulateJSON: true }
						)
						.then(function (data) {
							this.userrow = data.data.data;
						});
				},
				add: function () {
					this.$refs["Form"].validate((valid) => {
						if (valid) {
							this.addloading = true;
							if (this.edit) {
								this.$http.post(
									"/sxdk/api.php?act=edit",
									{
										form: {
											...this.form,
											paperNumSetting: JSON.stringify(this.form.paperNumSetting),
											check_week: this.form.check_week.join(","),
											holiday: this.form.holiday ? 1 : 2,
											day_paper: this.form.day_paper ? 1 : 2,
											week_paper: this.form.week_paper ? 1 : 2,
											month_paper: this.form.month_paper ? 1 : 2,
											weekPaperSubmitWeek: this.form.weekPaperSubmitWeek - 1,
											randomLocation: this.form.randomLocation ? 1 : 2,
										}
									},
									{ emulateJSON: true }
								)
									.then(function (data) {
										this.addloading = false;
										this.query();
										if (data.data.code == 0) {
											vm.$message.success(data.data.msg);
											this.resetForm()
											this.dialog = false;
										} else {
											vm.$message.error(data.data.msg);
										}
									});
							} else {
								this.$http.post(
									"/sxdk/api.php?act=add",
									{
										form: {
											...this.form,
											paperNumSetting: JSON.stringify(this.form.paperNumSetting),
											check_week: this.form.check_week.join(","),
											holiday: this.form.holiday ? 1 : 2,
											day_paper: this.form.day_paper ? 1 : 2,
											week_paper: this.form.week_paper ? 1 : 2,
											month_paper: this.form.month_paper ? 1 : 2,
											weekPaperSubmitWeek: this.form.weekPaperSubmitWeek - 1,
											randomLocation: this.form.randomLocation ? 1 : 2,
										}
									},
									{ emulateJSON: true }
								)
									.then(function (data) {
										this.addloading = false;
										this.query();
										if (data.data.code == 0) {
											vm.$message.success(data.data.msg);
											this.resetForm()
											this.dialog = false;
										} else {
											vm.$message.error(data.data.msg);
										}
									});
							}
						} else {
							this.$message.error("请填写全部必写内容");
							return false;
						}
					});

				},
				sizechange: function (val) {
					this.pagesize = val;
					this.query();
				},
				pagechange: function (val) {
					this.currentpage = val;
					this.query();
				},
				handleClose: function (done) {
					this.resetForm()
					done();
				},
				BhandleClose: function (done) {
					this.BresetForm()
					done();
				},
				bindWxhandleClose: function (done) {
					this.bingWxresetForm()
					done();
				},
				wxpushhandleClose: function (done) {
					this.resetForm()
					this.formOther.wxpush_img=""
					done();
				},
				downCheckChange(val) {
					if (!val) this.form.down_check_time = ""
				},
				deleteOrder: function (item) {
					this.$confirm(`确定要删除订单：${item.phone}`, '删除警告', {
						distinguishCancelAndClose: true,
						confirmButtonText: '删除',
						cancelButtonText: '取消'
					})
						.then(() => {
							this.$http.post(
								"/sxdk/api.php?act=del",
								{
									...item
								},
								{ emulateJSON: true }
							)
								.then(function (data) {
									this.query();
									if (data.data.code == 0) {
										vm.$message.success(data.data.msg);
									} else {
										vm.$message.error(data.data.msg);
									}
								});
						})
						.catch(action => { });
				},
				getLog: function (item) {
					this.$http.post(
						"/sxdk/api.php?act=getLog",
						{
							...item
						},
						{ emulateJSON: true }
					)
						.then(function (data) {
							if (data.data.code == 0) {
								let htmlText = data.data.data.map((i) => {
									return `<p>时间：${i.logTime}<br/>类型：${i.logType}<br/>内容：${i.logText}</p><hr/>`
								}).join("")
								this.$alert(htmlText, `${item.phone}的最近10条日志`, {
									dangerouslyUseHTMLString: true
								});
							} else {
								vm.$message.error(data.data.msg);
							}
						});
				},
				nowCheck: function (item) {
					this.$confirm(`账号：${item.phone}，现在要立即打卡么？打卡成功将会扣除该项目一天的余额`, '立即打卡提示', {
						distinguishCancelAndClose: true,
						confirmButtonText: '立即打卡',
						cancelButtonText: '取消'
					})
						.then(() => {
							this.$http.post(
								"/sxdk/api.php?act=nowCheck",
								{
									...item
								},
								{ emulateJSON: true }
							)
								.then(function (data) {
									this.query();
									if (data.data.code == 0) {
										vm.$message.success(data.data.msg);
									} else {
										vm.$message.error(data.data.msg);
									}
								});
						})
						.catch(action => { });

				},
				buPapers:function(){
				    	this.$refs["BForm"].validate((valid) => {
						if (valid) {
							this.buPapersloading = true;
							this.$http.post(
								"/sxdk/api.php?act=buPapers",
								{
									...this.buPapersForm
								},
								{ emulateJSON: true }
							)
								.then(function (data) {
									if (data.data.code == 0) {
    								    this.buPapersloading = false;
    							        this.BresetForm()
										vm.$message.success(data.data.msg);
										this.bDialog = false;
									} else {
										vm.$message.error(data.data.msg);
									}
								});
						} else {
							this.$message.error("请填写全部必写内容");
							return false;
						}
					});
				    
				},
				xybBindWx:function(){
					this.bindWxLoading = true;
					this.$http.post(
						"/sxdk/api.php?act=xybBindWx",
						{
							...this.bindWxForm
						},
						{ emulateJSON: true }
					)
					.then(function (data) {
					    this.bindWxLoading = false;
						if (data.data.code == 0) {
						    this.BresetForm()
							vm.$message.success(data.data.msg);
							this.bindWxDialog=false
						} else {
							vm.$message.error(data.data.msg);
						}
					});
				},
				useShowDoc:function(){
					this.wxpushLoading = true;
					this.$http.post(
						"/sxdk/api.php?act=useShowDoc",
						{
							form:this.form,
						},
						{ emulateJSON: true }
					)
					.then(function (data) {
					    this.wxpushLoading = false;
						if (data.data.code == 0) {
						    this.resetForm()
					        this.formOther.wxpush_img=""
							vm.$message.success(data.data.msg);
							this.wxpushDialog=false
						} else {
							vm.$message.error(data.data.msg);
						}
					});
				},
				getAsyncTask:function(item){
				    this.$http.post(
						"/sxdk/api.php?act=getAsyncTask",
						{
							...item
						},
						{ emulateJSON: true }
					)
						.then(function (data) {
							if (data.data.code == 0) {
								let htmlText=data.data.data.filter((item) => {
                    				return item.taskName != "";
                    			}).map((item => {
                    				let taskData=JSON.parse(item.taskData);
                    				return `<li style="font-size: 12px; border-bottom: 1px solid #ddd; padding: 5px 2px;text-align:left">类型：${taskData.levelName}<br/>日期范围：${taskData.startTime}至${taskData.endTime}<br/>状态：${item.code==1?'待执行':item.code==2?'执行中':'已关闭'}<br/>下单时间：${item.createTime}<br/>结束时间：${item.endTime}<br/>更新时间：${item.updateTime}<br/>运行日志：${item.taskMsg}</li>`;
                    			})).join("");
								this.$alert(htmlText, `${item.phone}的补报告记录`, {
									dangerouslyUseHTMLString: true
								});
							} else {
								vm.$message.error(data.data.msg);
							}
						});
				},
				changeCheckCode: function (item) {
					this.$confirm(`账号：${item.phone}，是否要${item.code != 1 ? '启动' : '暂停'}订单？`, `${item.code != 1 ? '启动' : '暂停'}订单`, {
						distinguishCancelAndClose: true,
						confirmButtonText: `${item.code != 1 ? '启动' : '暂停'}`,
						cancelButtonText: '取消'
					})
						.then(() => {
							const form = JSON.parse(JSON.stringify(item));
							if (form["code"] == 1) {
								form["code"] = 2;
							} else {
								form["code"] = 1;
							}
							this.$http.post(
								"/sxdk/api.php?act=changeCheckCode",
								{
									...form
								},
								{ emulateJSON: true }
							)
								.then(function (data) {
									this.query();
									if (data.data.code == 0) {
										vm.$message.success(data.data.msg);
									} else {
										vm.$message.error(data.data.msg);
									}
								});
						})
						.catch(action => { });
				},
				getWxPush: function (item) {
					this.$http.post(
						"/sxdk/api.php?act=getWxPush",
						{
							...item
						},
						{ emulateJSON: true }
					)
						.then(function (data) {
							if (data.data.code == 0) {
							    this.formOther.wxpush_img=data.data.data.url
							} else {
								vm.$message.error(data.data.msg);
							}
						});
				},
				async querySourceOrder(item) {
					let res = await this.$http.post(
						"/sxdk/api.php?act=querySourceOrder",
						{
							form: item,
						},
						{ emulateJSON: true }
					);
					if (res.data.code == 0) {
						return res.data.data;
					} else {
						vm.$message.error(res.data.msg);
						return false;
					}
				},
				update() {
					this.updateLoading = true
					this.$http.post(
						"/sxdk/api.php?act=yunOrder",
						{},
						{ emulateJSON: true }
					).then((res) => {
						if (res.data.code == 0) {
							vm.$message.success(res.data.msg);
							this.query()
						} else {
							vm.$message.error(res.data.msg);

						}
						this.updateLoading = false
					})

				},
				getNotice() {
					this.$http.post(
						"/sxdk/api.php?act=getNotice",
						{},
						{ emulateJSON: true }
					)
						.then(function (data) {
							if (data.data.code == 0) {
								this.noticeList = data.data.data
								this.showNotice()
							}
						});
				},
				showNotice() {
					if (this.noticeList.length <= 0) {
						return;
					}
					let item = this.noticeList.shift();

					this.$alert(item["content"], "公告", {
						dangerouslyUseHTMLString: true,
						callback: action => {
							if (this.noticeList.length > 0) {
								this.showNotice()
							}

						}
					});

				},
				xxyGetSchoolList() {
					this.$http.post(
						"/sxdk/api.php?act=xxyGetSchoolList",
						{},
						{ emulateJSON: true }
					)
						.then(function (data) {
							if (data.data.code == 20000) {
								data.data.data.forEach((par) => {
									par.schools.forEach((item) => {
									    let al=this.xxySchoolList.filter((alItem)=>{
									        return alItem.value==item.school_id
									    })
									    if (al.length>0){
									        item.school_id=item.school_id+" "
									    }
										this.xxySchoolList.push({
											text: item.school_name,
											value: item.school_id,
											url: item.differ_api,
										});
									});
								});
							} else {
								vm.$message.error("学校列表加载失败，请向上级反馈");
							}
						});
				},

				xxtGetSchoolList(value) {
					this.xxtgetSchoolLoading = true
					this.$http.post(
						"/sxdk/api.php?act=xxtGetSchoolList",
						{
							filter: value
						},
						{ emulateJSON: true }
					)
						.then(function ({ data }) {
							if (data.result) {
								this.xxtSchoolList = data.froms.map((item) => {
									return {
										text: item.name,
										value: item.schoolid,
									};
								});
							} else {
								this.xxtSchoolList = []
								vm.$message.error("请输入正确的学校名称");
							}
							this.xxtgetSchoolLoading = false
						});
				},
				hzjGetSchoolList() {
					this.$http.post(
						"/sxdk/api.php?act=hzjGetSchoolList",
						{},
						{ emulateJSON: true }
					)
						.then(function (data) {
							if (data.data.res == "success") {
								let obj = data.data.data
								for (let key in obj) {
									obj[key].forEach((item) => {
										this.hzjSchoolList.push({
											text: item.schoolName,
											value: item.schoolId,
										});
									});
								}
							} else {
								vm.$message.error("学校列表加载失败，请向上级反馈");
							}
						});
				},
				xxyAddressSearchPoi(query) {
					if (query) {
						this.xxyRemote.loading = true
						this.$http
							.post(
								"/sxdk/api.php?act=xxyAddressSearchPoi",
								{
									form: {
										text: query,
										lat: this.form.lat,
										lng: this.form.lng,
									}
								},
								{ emulateJSON: true }
							)
							.then((res) => {
								this.xxyRemote.loading = false
								if (res.data.code == 0) {
									this.xxyAddressPois = res.data.data.addressPois.map((item) => {
										return {
											...item,
											textValue: item.value + item.text,
										};
									});
								} else {
									this.$message.error(res.data.msg);
									this.xxyRemote.loading = false
									this.xxyAddressPois = []
								}
							}).catch((e) => {
								this.$message.error(res.data.msg);
								this.xxyRemote.loading = false
								this.xxyAddressPois = []
							});
					} else {
						this.xxyAddressPois = []
					}
				},
				handleResize() {
					window.addEventListener('resize', this.onResize);
				},
				onResize() {
					if (window.innerWidth <= 675) {
						this.addeditDialogWidth = "100%"
						this.allSize = "mini"
					} else {
						this.addeditDialogWidth = "675px"
						this.allSize = "small"
					}
					if (window.innerHeight <= 600) {
						this.addeditDialogHeight = "100%"
					} else {
						this.addeditDialogHeight = "500px"
					}
				},
				tableRowClassName({ row, rowIndex }) {
					if (row.code == 1) {
						return 'success-row';
					} else {
						return 'warning-row';
					}
				},
			},
			beforeDestroy() {
				window.removeEventListener('resize', this.onResize);
			},

		});
	</script>
	<style>
		.el-table .warning-row {
			background: oldlace;
		}

		.el-table .success-row {
			background: white;
		}

		.el-message-box {
			width: 320px;

		}

		.el-message-box .el-message-box__content {
			max-height: 600px;
			overflow: auto;
		}

		.addeditDialog .el-dialog {
			margin-top: 5vh !important;
		}

		.addeditDialog .el-dialog__body {
			padding: 30px 2vw;
		}

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
			height: 65vh;
			overflow: auto;
		}
	</style>
</div>