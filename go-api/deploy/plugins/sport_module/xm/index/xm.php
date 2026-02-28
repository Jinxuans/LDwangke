<?php
$mod='blank';
$title='XM运动';
require_once('head.php');
?>
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>XM运动 · 下单中心</title>
    <link rel="stylesheet" href="https://unpkg.com/element-plus/dist/index.css" />
    <style>
        body {
            background: #f5f5f5;
            padding: 30px;
            font-family: "Helvetica Neue", Arial, sans-serif;
        }
        .order-form-container {
            max-width: 1200px;
            width: 95%;
            margin: 40px auto;
            background: #fff;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 12px rgba(0,0,0,0.1);
            min-height: 700px;
        }
        .form-submit {
            text-align: right;
            margin-top: 20px;
        }
        .form-title {
            font-size: 26px;
            color: #409EFF;
            font-weight: 600;
            letter-spacing: 2px;
        }

        @media (max-width: 768px) {
            .order-form-container {
                padding: 15px;
                margin: 20px auto;
                min-height: auto;
            }
            .form-title {
                font-size: 20px;
                letter-spacing: 1px;
            }
            .el-row {
                display: block;
            }
            .el-col {
                width: 100% !important;
                max-width: 100% !important;
            }
            .el-button-group {
                flex-wrap: wrap;
            }
            .el-button-group .el-button {
                margin-bottom: 10px;
                width: 100%;
            }
        }
    </style>
</head>
<body>
<div id="app"></div>

<!-- Vue -->
<script src="https://unpkg.com/vue@3/dist/vue.global.prod.js"></script>
<!-- Element Plus -->
<script src="https://unpkg.com/element-plus/dist/index.full.min.js"></script>
<!-- Element Plus 中文 -->
<script src="https://unpkg.com/element-plus/dist/locale/zh-cn.min.js"></script>

<script>
    const { createApp, ref, reactive, watch, onMounted } = Vue;
    const zhCn = ElementPlusLocaleZhCn;

    const OrderForm = {
        template: `
          <div class="order-form-container">
            <el-divider content-position="center">
              <span class="form-title">XM运动 · 下单中心</span>
            </el-divider>

            <el-form :model="orderForm" :rules="rules" label-width="120px" ref="orderFormRef">
              <el-form-item label="项目" prop="project_id">
                <el-select v-model="orderForm.project_id" placeholder="请选择项目" clearable>
                  <el-option
                      v-for="item in projectOptions"
                      :key="item.id"
                      :label="item.name + ' - ' + item.price + ' 元'"
                      :value="item.id">
                  </el-option>
                </el-select>
              </el-form-item>

              <el-form-item label="项目描述" v-if="projectDescription">
                <div>{{ projectDescription }}</div>
              </el-form-item>

              <el-form-item label="价格" v-if="projectPrice !== null">
                <div>{{ projectPrice }} 元</div>
              </el-form-item>

              <el-row :gutter="10">
                <el-col :span="12">
                  <el-form-item label="账号" prop="account">
                    <el-input v-model="orderForm.account" placeholder="请输入账号"></el-input>
                  </el-form-item>
                </el-col>
                <el-col :span="12" v-if="needPassword">
                  <el-form-item label="密码" prop="password">
                    <el-input v-model="orderForm.password" placeholder="请输入密码"></el-input>
                  </el-form-item>
                </el-col>
                <el-col v-if="showQueryButton" :span="24">
                  <el-button type="primary" @click="handleQueryAccount" :loading="queryLoading" style="width: 100%;">
                    查询账号信息
                  </el-button>
                </el-col>
              </el-row>

              <el-form-item label="学校" prop="school">
                <el-input v-model="orderForm.school" placeholder="请输入学校"></el-input>
              </el-form-item>

              <el-form-item label="跑步类型" v-if="runRoleOptions.length > 0">
                <el-button-group>
                  <el-button
                      v-for="(item, index) in runRoleOptions"
                      :key="index"
                      :type="selectedRunRoleIndex === index ? 'primary' : 'default'"
                      @click="() => { selectedRunRoleIndex = index; applyRunRole(item.raw); }"
                  >{{ item.label }}</el-button>
                </el-button-group>
              </el-form-item>

              <el-form-item label="时间段" v-if="selectedRunRoleIndex !== null">
                <el-button-group>
                  <el-button
                      v-for="(time, index) in runRoleOptions[selectedRunRoleIndex].raw.run_times"
                      :key="index"
                      :type="selectedTimeIndex === index ? 'success' : 'default'"
                      @click="() => { selectedTimeIndex = index; applyRunTime(time); }"
                  >{{ time.start_time }} - {{ time.end_time }}</el-button>
                </el-button-group>
              </el-form-item>

              <el-form-item :label="runType === '次数' ? '下单次数' : '下单公里'" prop="total_km">
                <el-input-number v-model="orderForm.total_km" :min="0" controls-position="right"></el-input-number>
              </el-form-item>

              <el-form-item label="开始日期" prop="start_day">
                <el-date-picker
                    v-model="orderForm.start_day"
                    type="date"
                    placeholder="请选择开始日期"
                    format="YYYY年MM月DD日"
                    value-format="YYYY-MM-DD"
                ></el-date-picker>
              </el-form-item>

              <el-form-item label="开始时间" prop="start_time">
                <el-time-picker
                    v-model="orderForm.start_time"
                    placeholder="请选择开始时间"
                    format="HH:mm"
                    value-format="HH:mm"
                ></el-time-picker>
              </el-form-item>

              <el-form-item label="结束时间" prop="end_time">
                <el-time-picker
                    v-model="orderForm.end_time"
                    placeholder="请选择结束时间"
                    format="HH:mm"
                    value-format="HH:mm"
                ></el-time-picker>
              </el-form-item>

              <el-form-item label="跑步周期" prop="run_date">
                <el-checkbox
                    :indeterminate="isIndeterminate"
                    v-model="checkAll"
                    @change="handleCheckAllChange"
                >全选</el-checkbox>
                <el-checkbox-group
                    v-model="orderForm.run_date"
                    @change="handleCheckedWeekChange"
                >
                  <el-checkbox
                      v-for="item in weekOptions"
                      :key="item.value"
                      :label="item.value"
                  >{{ item.label }}</el-checkbox>
                </el-checkbox-group>
              </el-form-item>

              <el-form-item class="form-submit">
                <el-button type="primary" @click="submitOrder">
                  提交
                </el-button>
              </el-form-item>
            </el-form>
          </div>
        `,
        setup() {
            const orderForm = reactive({
                project_id: null,
                school: '',
                account: '',
                password: '',
                total_km: 0,
                start_day: '',
                start_time: '',
                end_time: '',
                run_date: [],
                type: 0
            });

            const orderFormRef = ref(null);

            const rules = reactive({
                project_id: [{ required: true, message: '请选择项目', trigger: 'change' }],
                account: [{ required: true, message: '请输入账号', trigger: 'blur' }],
                password: [],
                school: [{ required: true, message: '请输入学校名称', trigger: 'blur' }],
                total_km: [{ required: true, message: '请输入下单公里', type: 'number', trigger: 'blur' }],
                start_day: [{ required: true, message: '请选择开始日期', trigger: 'change' }],
                start_time: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
                end_time: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
                run_date: [{ required: true, message: '请选择跑步周期', trigger: 'change' }]
            });

            const projectOptions = ref([]);
            const runRoleOptions = ref([]);
            const selectedRunRoleIndex = ref(null);
            const selectedTimeIndex = ref(null);
            const projectDescription = ref('');
            const projectPrice = ref(null);
            const runType = ref('公里');
            const showQueryButton = ref(false);
            const queryLoading = ref(false);
            const needPassword = ref(false);

            const weekOptions = [
                { label: '星期一', value: 1 },
                { label: '星期二', value: 2 },
                { label: '星期三', value: 3 },
                { label: '星期四', value: 4 },
                { label: '星期五', value: 5 },
                { label: '星期六', value: 6 },
                { label: '星期日', value: 7 }
            ];

            const checkAll = ref(false);
            const isIndeterminate = ref(false);

            watch(() => orderForm.project_id, (newVal) => {
                const selected = projectOptions.value.find(p => p.id === newVal);
                showQueryButton.value = Number(selected?.query) === 1;
                needPassword.value = Number(selected?.password) === 1;
                projectDescription.value = selected?.description || '';
                projectPrice.value = selected?.price ?? null;
            });

            onMounted(() => {
                fetchProjectList();
            });

            function fetchProjectList() {
                fetch('/xm_apis.php?act=getProjects', {
                    method: 'GET',
                    credentials: 'include'
                })
                    .then(res => res.json())
                    .then(res => {
                        if (res.code === 1 && Array.isArray(res.data)) {
                            projectOptions.value = res.data;
                        } else {
                            ElementPlus.ElMessage.error(res.msg || '获取项目列表失败');
                        }
                    })
                    .catch(err => {
                        console.error(err);
                        ElementPlus.ElMessage.error('请求接口失败');
                    });
            }

            function handleQueryAccount() {
                if (!orderForm.account) {
                    ElementPlus.ElMessage.error('请输入账号后再查询！');
                    return;
                }
                if (needPassword.value && !orderForm.password) {
                    ElementPlus.ElMessage.error('该项目需要密码，请输入密码后再查询！');
                    return;
                }
                if (!orderForm.project_id) {
                    ElementPlus.ElMessage.error('请先选择项目！');
                    return;
                }

                queryLoading.value = true;

                const payload = {
                    account: orderForm.account,
                    password: orderForm.password,
                    project_id: orderForm.project_id
                };

                fetch('/xm_apis.php?act=query_run', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify(payload)
                })
                    .then(res => res.json())
                    .then(res => {
                        queryLoading.value = false;

                        if (res.code === 200 && Array.isArray(res.data) && res.data.length > 0) {
                            const info = res.data[0];

                            if (info.status === -1) {
                                ElementPlus.ElMessage.error(info.error || '查询失败！');
                                return;
                            }

                            orderForm.school = info.school;

                            if (info.run_roles && info.run_roles.length > 0) {
                                runRoleOptions.value = info.run_roles.map((role, index) => ({
                                    label: role.run_type || `跑步方案 ${index + 1}`,
                                    value: index,
                                    raw: role
                                }));
                                selectedRunRoleIndex.value = 0;
                                applyRunRole(runRoleOptions.value[0].raw);
                            } else {
                                runRoleOptions.value = [];
                            }

                            ElementPlus.ElMessage.success('账号信息查询成功！');
                        } else {
                            ElementPlus.ElMessage.error(res.msg || '查询失败！');
                        }
                    })
                    .catch(err => {
                        queryLoading.value = false;
                        console.error(err);
                        ElementPlus.ElMessage.error('查询请求失败！');
                    });
            }

            function applyRunRole(role) {
                orderForm.total_km = role.total_km;
                orderForm.run_date = role.run_date;
                orderForm.start_day = role.start_day;
                orderForm.type = Number(role.type) ?? 0;
                runType.value = role.run_type || '公里';
                orderForm.start_time = '';
                orderForm.end_time = '';
                selectedTimeIndex.value = null;
            }

            function applyRunTime(time) {
                orderForm.start_time = time.start_time;
                orderForm.end_time = time.end_time;
            }

            function handleCheckAllChange(val) {
                orderForm.run_date = val ? weekOptions.map(item => item.value) : [];
                isIndeterminate.value = false;
            }

            function handleCheckedWeekChange(value) {
                const checkedCount = value.length;
                checkAll.value = checkedCount === weekOptions.length;
                isIndeterminate.value = checkedCount > 0 && checkedCount < weekOptions.length;
            }

            function submitOrder() {
                rules.password = [];
                if (needPassword.value) {
                    rules.password.push({ required: true, message: '请输入密码', trigger: 'blur' });
                }

                orderFormRef.value.validate((valid) => {
                    if (!valid) {
                        return;
                    }

                    const payload = {
                        project_id: orderForm.project_id,
                        school: orderForm.school,
                        account: orderForm.account,
                        password: orderForm.password,
                        total_km: orderForm.total_km,
                        run_date: orderForm.run_date,
                        start_day: orderForm.start_day,
                        start_time: orderForm.start_time,
                        end_time: orderForm.end_time,
                        type: orderForm.type
                    };

                    fetch('/xm_apis.php?act=add_order', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        credentials: 'include',
                        body: JSON.stringify(payload)
                    })
                        .then(res => res.json())
                        .then(res => {
                            if (res.code === 200) {
                                ElementPlus.ElMessage.success(res.msg || '下单成功');
                            } else {
                                ElementPlus.ElMessage.error(res.msg || '下单失败');
                            }
                        })
                        .catch(err => {
                            console.error(err);
                            ElementPlus.ElMessage.error('下单请求失败');
                        });
                });
            }

            return {
                orderForm,
                orderFormRef,
                rules,
                projectOptions,
                projectDescription,
                projectPrice,
                showQueryButton,
                queryLoading,
                runRoleOptions,
                selectedRunRoleIndex,
                selectedTimeIndex,
                weekOptions,
                runType,
                checkAll,
                isIndeterminate,
                handleQueryAccount,
                applyRunRole,
                applyRunTime,
                handleCheckAllChange,
                handleCheckedWeekChange,
                submitOrder,
                needPassword
            };
        }
    };

    const app = createApp({
        components: { OrderForm },
        template: `<el-config-provider :locale="locale"><order-form/></el-config-provider>`,
        setup() {
            return { locale: zhCn };
        }
    });

    app.use(ElementPlus);
    app.mount('#app');
</script>
</body>
</html>
