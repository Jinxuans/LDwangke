axios = axios.create({
  timeout: 180000,
});
axios.defaults.headers.post["Content-Type"] =
  "application/x-www-form-urlencoded";
axios.interceptors.response.use(
  function (response) {
    return response.data;
  },
  function (error) {
    return Promise.reject(error);
  }
);

// 消息提示
const toast = {
  success: (message = "操作成功") => {
    ElMessage({
      message,
      type: "success",
    });
  },
  error: (message = "操作失败") => {
    ElMessage({
      message,
      type: "error",
    });
  },
};

// 加载方法
const FullLoading = {
  show: (text = "加载中...") => {
    return ElLoading.service({
      lock: true,
      text,
      background: "rgba(0, 0, 0, 0.7)",
    });
  },
  close: () => {
    ElLoading.service().close();
  },
};

// 引入 ElementPlus
const { ElMessage, ElLoading, ElMessageBox, ElNotification } = ElementPlus;
// 引入 Vue 钩子
const {
  createApp,
  ref,
  onBeforeMount,
  onMounted,
  reactive,
  nextTick,
  watch,
  computed,
} = Vue;

const app = createApp({
  setup() {
    const tableData = ref([]);
    const pagination = ref({
      total: 0,
      limit: 10,
      page: 1,
    });
    const search = ref({ type: "", keywords: "" });
    const tableLoading = ref(true);

    // 获取数据
    const loadData = (page = 1) => {
      tableLoading.value = true;
      axios
        .post("/flash/api.php?act=orders", {
          page,
          limit: pagination.value.limit,
          type: search.value.type,
          keywords: search.value.keywords,
        })
        .then((res) => {
          tableLoading.value = false;
          if (res.code === 0) {
            tableData.value = res.data || [];
            pagination.value = res.pagination;
          }
        });
    };

    // 切换每页数据
    const handleSizeChange = (val) => {
      pagination.value.limit = val;
      get(pagination.value.page);
    };

    // 翻页
    const handleCurrentChange = (val) => {
      get(val);
    };

    // 获取价格
    const price = ref(0);
    const getPrice = async () => {
      const { data } = await axios.post("/flash/api.php?act=get_price");
      price.value = data || 0;
    };

    // 处理菜单点击事件
    const handleMenu = (command) => {
      const { item, type } = command;
      if (type == "log") {
        logData.value.sdxy_order_id = item.sdxy_order_id;
        loadLogData();
      } else if (type == "delay") {
        handleDelayTask(item);
      } else if (type == "refund") {
        handleRefundOrder(item);
      }
    };

    // ======================================查看日志======================================
    const logDialogVisible = ref(false);
    const logData = ref({
      sdxy_order_id: "",
      list: [],
      pagination: {
        page_num: 1,
        page_size: 10,
        total: 0,
      },
    });
    const loadLogData = () => {
      FullLoading.show();

      const params = {
        sdxy_order_id: logData.value.sdxy_order_id,
        page_num: logData.value.pagination.page_num,
        page_size: logData.value.pagination.page_size,
      };

      axios.post("/flash/api.php?act=log", params).then((res) => {
        FullLoading.close();
        if (res.code === 0) {
          logDialogVisible.value = true;
          logData.value.list = res.data.list || [];
          logData.value.pagination.total = res.data.total || 0;
        }
      });
    };
    // 日志切换每页数据
    const handleLogSizeChange = (val) => {
      logData.value.pagination.page_size = val;
      loadLogData();
    };
    // 日志翻页
    const handleLogCurrentChange = (val) => {
      logData.value.pagination.page_num = val;
      loadLogData();
    };
    // 修改任务时间
    const changeTaskTimeForm = ref({
      sdxy_order_id: "",
      run_task_id: "",
      start_time: "",
    });
    const handleChangeTaskTime = (row) => {
      changeTaskTimeForm.value.sdxy_order_id = row.order_id || "";
      changeTaskTimeForm.value.run_task_id = row.run_task_id || "";
      changeTaskTimeForm.value.start_time = row.start_time || "";
    };
    const handleSaveTaskTime = () => {
      FullLoading.show();
      axios
        .post("/flash/api.php?act=change_task_time", {
          sdxy_order_id: changeTaskTimeForm.value.sdxy_order_id,
          run_task_id: changeTaskTimeForm.value.run_task_id,
          start_time: changeTaskTimeForm.value.start_time,
        })
        .then((res) => {
          FullLoading.close();
          if (res.code === 0) {
            toast.success(res.msg || "修改成功");
            handleCancelChangeTaskTime();
            loadLogData();
          } else {
            toast.error(res.msg || "修改失败");
          }
        });
    };
    const handleCancelChangeTaskTime = () => {
      changeTaskTimeForm.value.run_task_id = "";
      changeTaskTimeForm.value.start_time = "";
    };
    // ===================================================================================

    // ======================================延迟任务======================================
    const handleDelayTask = (row) => {
      FullLoading.show();
      axios
        .post("/flash/api.php?act=delay_task", {
          agg_order_id: row.agg_order_id,
          run_task_id: row.run_task_id,
        })
        .then((res) => {
          FullLoading.close();
          if (res.code === 0) {
            toast.success(res.msg || "延迟成功");
            loadData(pagination.value.page);
          } else {
            toast.error(res.msg || "延迟失败");
          }
        });
    };
    // ====================================================================================

    // ======================================退款订单======================================
    const handleRefundOrder = (row) => {
      ElMessageBox.confirm("您确认要退款该订单吗？", "提示", {
        type: "warning",
      })
        .then((_) => {
          FullLoading.show();
          axios
            .post("/flash/api.php?act=refund", {
              agg_order_id: row.agg_order_id,
            })
            .then((res) => {
              FullLoading.close();
              if (res.code === 0) {
                toast.success(res.msg || "退款成功");
                loadData(pagination.value.page);
              } else {
                toast.error(res.msg || "退款失败");
              }
            });
        })
        .catch((_) => {});
    };
    // =====================================================================================

    // ======================================切换跑步状态====================================
    const changePause = (row) => {
      FullLoading.show();
      axios
        .post("/flash/api.php?act=pause", {
          agg_order_id: row.agg_order_id,
          pause: row.pause,
        })
        .then((res) => {
          FullLoading.close();
          if (res.code === 0) {
            toast.success(res.msg);
            loadData(pagination.value.page);
          } else {
            toast.error(res.msg);
          }
        });
    };
    // ====================================================================================

    // ======================================提交订单=======================================
    const addDialogVisible = ref(false);
    const addLoading = ref(false);
    const addForm = ref({
      task_list: [],
      phone: "",
      dis: 1.0,
      zone_id: "",
      run_type: "SUN",
      student_id: "",
      password: "",
      run_rule_id: "",
    });
    const userInfo = ref({});
    const userInfoForm = ref({
      loginType: "password",
      phone: "",
      password: "",
      code: "",
    });

    // 通过密码查询用户信息
    const getUserInfoByPassword = () => {
      if (!userInfoForm.value.phone) return toast.error("请输入手机号");

      FullLoading.show();
      const form = {
        phone: userInfoForm.value.phone,
        password: userInfoForm.value.password,
      };
      axios
        .post("/flash/api.php?act=get_user_info_by_password", { form })
        .then((res) => {
          FullLoading.close();
          if (res.code === 0) {
            userInfo.value = res.data || {};
            initAddForm();
          } else {
            toast.error(res.msg || "查询用户信息失败");
          }
        });
    };

    // 发送验证码
    const sendCode = () => {
      if (!userInfoForm.value.phone) return toast.error("请输入手机号");

      FullLoading.show();
      const form = {
        phone: userInfoForm.value.phone,
      };
      axios.post("/flash/api.php?act=send_code", { form }).then((res) => {
        FullLoading.close();
        if (res.code === 0) {
          toast.success(res.msg || "验证码发送成功");
        } else {
          toast.error(res.msg || "验证码发送失败");
        }
      });
    };

    // 通过验证码查询用户信息
    const getUserInfoByCode = () => {
      if (!userInfoForm.value.phone) return toast.error("请输入手机号");
      if (!userInfoForm.value.code) return toast.error("请输入验证码");

      FullLoading.show();
      const form = {
        phone: userInfoForm.value.phone,
        code: userInfoForm.value.code,
      };
      axios
        .post("/flash/api.php?act=get_user_info_by_code", { form })
        .then((res) => {
          FullLoading.close();
          if (res.code === 0) {
            userInfo.value = res.data || {};
            initAddForm();
          } else {
            toast.error(res.msg || "查询用户信息失败");
          }
        });
    };

    // 更新跑步计划
    const updateRunRule = () => {
      if (!userInfo.value.student.student_id)
        return toast.error("请先查询用户信息");

      FullLoading.show();
      axios
        .post("/flash/api.php?act=update_run_rule", {
          student_id: userInfo.value.student.student_id,
        })
        .then((res) => {
          FullLoading.close();
          if (res.code === 0) {
            toast.success(res.msg || "更新计划成功");
          } else {
            toast.error(res.msg || "更新计划失败");
          }
        });
    };

    // 初始化提交表单
    const initAddForm = () => {
      addForm.value = {
        task_list: [],
        phone: userInfoForm.value.phone,
        dis: userInfo.value?.student?.run_rule?.min_dis || 1.0,
        zone_id: userInfo.value?.student?.default_zone?.zone_id || "",
        run_type: "SUN",
        student_id: userInfo.value?.student?.student_id || "",
        password:
          userInfoForm.value.password ||
          userInfo.value?.student?.password ||
          "",
        run_rule_id: userInfo.value?.student?.run_rule?.run_rule_id || "",
      };
    };

    // 跑步计划
    const runRuleList = computed(() => {
      return userInfo.value?.student?.run_rule_lst || [];
    });

    // 跑步区域
    const zoneList = computed(() => {
      return userInfo.value?.zone_list || [];
    });

    // 设置时间段
    const planTimeConfig = ref([]);
    const planTimeList = ref([]);

    // 跑步星期
    const weekOptions = computed(() => {
      return [
        {
          value: 1,
          label: "周一",
        },
        {
          value: 2,
          label: "周二",
        },
        {
          value: 3,
          label: "周三",
        },
        {
          value: 4,
          label: "周四",
        },
        {
          value: 5,
          label: "周五",
        },
        {
          value: 6,
          label: "周六",
        },
        {
          value: 0,
          label: "周日",
        },
      ];
    });

    // 生成任务列表的方法
    const generateTaskList = (time_range, start_date, num, week) => {
      const task_list = [];
      const startDate = new Date(start_date);
      const [startTimeStr, endTimeStr] = time_range;

      // 解析开始和结束时间为分钟数
      const parseTimeToMinutes = (timeStr) => {
        const [hours, minutes] = timeStr?.split(":")?.map(Number) || [0, 0];
        return hours * 60 + minutes;
      };

      const startTimeMinutes = parseTimeToMinutes(startTimeStr);
      const endTimeMinutes = parseTimeToMinutes(endTimeStr);

      // 格式化日期为YYYY-MM-DD格式
      const formatDate = (date) => {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, "0");
        const day = String(date.getDate()).padStart(2, "0");
        return `${year}-${month}-${day}`;
      };

      // 生成随机时间
      const generateRandomTime = () => {
        const randomMinutes =
          Math.floor(Math.random() * (endTimeMinutes - startTimeMinutes + 1)) +
          startTimeMinutes;
        const hours = Math.floor(randomMinutes / 60);
        const minutes = randomMinutes % 60;
        const seconds = Math.floor(Math.random() * 60);
        return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(
          2,
          "0"
        )}:${String(seconds).padStart(2, "0")}`;
      };

      let currentDate = new Date(startDate);
      let count = 0;

      // 遍历日期，直到生成足够数量的任务
      while (count < num) {
        const dayOfWeek = currentDate.getDay(); // 0表示星期日，1-6表示星期一到星期六

        // 检查日期是否在允许的星期列表中
        if (week.includes(dayOfWeek)) {
          const formattedDate = formatDate(currentDate);
          const randomTime = generateRandomTime();

          task_list.push({
            start_time: `${formattedDate} ${randomTime}`,
          });

          count++;
        }

        // 移动到下一天
        currentDate.setDate(currentDate.getDate() + 1);
      }

      return task_list;
    };

    // 生成计划时间列表
    watch(
      () => planTimeConfig.value,
      (newVal, oldVal) => {
        // 使用nextTick并确保在下次DOM更新周期才修改值
        nextTick(() => {
          planTimeList.value = newVal.map((plan, index) => {
            const task_list = generateTaskList(
              plan.time_range,
              plan.start_date,
              plan.num,
              plan.week
            );
            return { task_list };
          });
        });
      },
      {
        deep: true,
        immediate: false,
      }
    );

    // 处理时间提前函数
    const handleTimeAdvance = (timeStr) => {
      if (!timeStr) return timeStr;
      // 将时间字符串转换为Date对象
      const [hours, minutes] = timeStr.split(":").map(Number);
      const date = new Date();
      date.setHours(hours);
      date.setMinutes(minutes);

      // 提前15分钟
      date.setMinutes(date.getMinutes() - 30);

      // 格式化时间为HH:MM格式
      const newHours = String(date.getHours()).padStart(2, "0");
      const newMinutes = String(date.getMinutes()).padStart(2, "0");
      return `${newHours}:${newMinutes}`;
    };

    // 新增时段
    const addPlanTime = () => {
      const index = planTimeConfig.value.length;
      const beginTime = runRuleInfo.value?.time_slot[index]?.beg || "08:00";
      let endTime = runRuleInfo.value?.time_slot[index]?.end || "23:59";
      // 结束时间提前 30分钟
      endTime = handleTimeAdvance(endTime);
      const week = runRuleInfo.value?.time_slot[index]?.week || [
        1, 2, 3, 4, 5, 6, 0,
      ];

      planTimeConfig.value.push({
        time_range: [beginTime, endTime],
        start_date: new Date().toISOString().split("T")[0],
        num: 1,
        week: week,
      });
      planTimeList.value.push({ task_list: [] });
    };

    // 删除时段
    const removePlanTime = (index) => {
      planTimeConfig.value.splice(index, 1);
      planTimeList.value.splice(index, 1);
    };

    // 计划任务信息
    const runRuleInfo = computed(() => {
      return runRuleList.value.find(
        (item) => item.run_rule_id === addForm.value.run_rule_id
      );
    });

    // 时段显示信息
    const planTimeShow = (index) => {
      const beginTime = runRuleInfo.value?.time_slot[index]?.beg;
      const endTime = runRuleInfo.value?.time_slot[index]?.end;
      const minDis = runRuleInfo.value?.min_dis;
      // 把week中的数字转换为星期几
      const week = runRuleInfo.value?.time_slot[index]?.week
        ?.map(
          (item) =>
            weekOptions.value.find((weekItem) => weekItem.value === item)
              ?.label || ""
        )
        .join("");
      if (!beginTime || !endTime || !minDis || !week) return "";
      return `(${beginTime}至${endTime} | ${minDis}KM | ${week})`;
    };

    // 提交订单
    const handleAdd = () => {
      addForm.value.task_list = [];
      planTimeList.value.map((item) => {
        addForm.value.task_list.push(...item.task_list);
      });
      if (!addForm.value.run_rule_id) return toast.error("请选择跑步计划");
      if (!addForm.value.zone_id) return toast.error("请选择跑步区域");
      if (!addForm.value.phone) return toast.error("请输入手机号");
      if (!addForm.value.dis) return toast.error("请输入公里数");

      addLoading.value = true;
      axios
        .post("/flash/api.php?act=add", { form: addForm.value })
        .then((res) => {
          addLoading.value = false;
          if (res.code === 0) {
            toast.success(res.msg || "下单成功");
            addDialogVisible.value = false;
            resetAllForm();
            loadData(pagination.value.page);
          } else {
            toast.error(res.msg || "下单失败");
          }
        });
    };

    // 重置所有表单数据
    const resetAllForm = () => {
      addForm.value = {
        run_rule_id: "",
        zone_id: "",
        phone: "",
        dis: "",
        task_list: [],
      };
      planTimeConfig.value = [];
      planTimeList.value = [];
    };

    // 任务总数
    const taskTotal = computed(() => {
      return planTimeList.value.reduce(
        (total, item) => total + item.task_list.length,
        0
      );
    });

    // ====================================================================================

    onBeforeMount(() => {
      loadData();
      getPrice();
    });

    onMounted(() => {});

    return {
      tableData,
      tableLoading,
      pagination,
      search,
      price,
      loadData,
      handleMenu,
      handleSizeChange,
      handleCurrentChange,
      logDialogVisible,
      logData,
      handleLogSizeChange,
      handleLogCurrentChange,
      changeTaskTimeForm,
      handleChangeTaskTime,
      handleSaveTaskTime,
      handleCancelChangeTaskTime,
      handleDelayTask,
      handleRefundOrder,
      changePause,
      addDialogVisible,
      addLoading,
      addForm,
      userInfo,
      userInfoForm,
      getUserInfoByPassword,
      sendCode,
      getUserInfoByCode,
      updateRunRule,
      runRuleList,
      zoneList,
      planTimeConfig,
      planTimeList,
      weekOptions,
      removePlanTime,
      addPlanTime,
      runRuleInfo,
      planTimeShow,
      handleAdd,
      taskTotal,
    };
  },
});

// 安装 element-plus
app.use(ElementPlus, {
  locale: ElementPlusLocaleZhCn,
});

// 安装 element-plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component);
}

app.mount("#app");
