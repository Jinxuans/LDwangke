<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
    NavBar,
    Button,
    Field,
    Form,
    showToast,
    CellGroup,
    RadioGroup,
    Radio,
} from "vant";
import {
    getProductDetail,
    getPayChannels,
    createPay,
    queryCourse,
} from "../api";

const route = useRoute();
const router = useRouter();
const tid = route.params.tid as string;
const cid = Number(route.params.cid);

const product = ref<any>(null);
const loading = ref(true);
const submitting = ref(false);
const querying = ref(false);
const account = ref("");
const password = ref("");
const remark = ref("");

// 查课结果
const queryResult = ref<any>(null);
const queryDone = ref(false);
const selectedCourses = ref<any[]>([]);

function toggleCourse(item: any) {
    const idx = selectedCourses.value.findIndex((c) => c.id === item.id);
    if (idx >= 0) {
        selectedCourses.value.splice(idx, 1);
    } else {
        selectedCourses.value.push(item);
    }
}
function isCourseSelected(item: any) {
    return selectedCourses.value.some((c) => c.id === item.id);
}

interface Channel {
    type: string;
    name: string;
}
const channels = ref<Channel[]>([]);
const payType = ref("");

onMounted(async () => {
    try {
        const [prod, chs] = await Promise.all([
            getProductDetail(cid),
            getPayChannels(),
        ]);
        product.value = prod;
        channels.value = (chs as unknown as Channel[]) || [];
        if (channels.value.length > 0) {
            payType.value = channels.value[0]!.type;
        }
    } catch (e: any) {
        showToast(e?.message || "加载失败");
    } finally {
        loading.value = false;
    }
});

async function handleQuery() {
    if (!account.value.trim()) {
        showToast("请填写账号");
        return;
    }
    if (!password.value.trim()) {
        showToast("请填写密码");
        return;
    }
    querying.value = true;
    queryResult.value = null;
    queryDone.value = false;
    try {
        const res: any = await queryCourse({
            cid,
            userinfo: `${account.value.trim()} ${password.value.trim()}`,
        });
        queryResult.value = res;
        queryDone.value = true;
        selectedCourses.value = [];
        // 若只有一门课，自动选中
        if (res?.data?.length === 1) {
            selectedCourses.value = [res.data[0]];
        }
    } catch (e: any) {
        showToast(e?.message || "查课失败");
    } finally {
        querying.value = false;
    }
}

async function handleOrder() {
    if (!queryDone.value) {
        showToast("请先查课");
        return;
    }
    if (queryResult.value?.data?.length > 0 && selectedCourses.value.length === 0) {
        showToast("请选择要购买的课程");
        return;
    }
    if (!payType.value) {
        showToast("请选择支付方式");
        return;
    }
    submitting.value = true;
    try {
        const res: any = await createPay({
            cid,
            account: account.value.trim(),
            password: password.value.trim(),
            pay_type: payType.value,
            remark: remark.value,
            course_id: selectedCourses.value.map((c) => c.id).join(","),
            course_name: selectedCourses.value.map((c) => c.name).join(","),
            course_kcjs: selectedCourses.value.map((c) => c.kcjs || "").join(","),
        });
        if (res?.pay_url && res?.out_trade_no) {
            // 存订单号，支付完成后结果页用于检测
            localStorage.setItem("pending_out_trade_no", res.out_trade_no);
            localStorage.setItem("pending_pay_url", res.pay_url);
            router.push(`/${tid}/pay-result?out_trade_no=${res.out_trade_no}`);
        } else {
            showToast({ type: "success", message: "下单成功", duration: 1500 });
            setTimeout(() => router.push(`/${tid}/orders`), 1500);
        }
    } catch (e: any) {
        showToast(e?.message || "提交失败");
    } finally {
        submitting.value = false;
    }
}

const payIcons: Record<string, string> = {
    alipay: "https://img.icons8.com/color/48/alipay.png",
    wxpay: "https://img.icons8.com/color/48/wechat.png",
    qqpay: "https://img.icons8.com/color/48/qq.png",
};

// Generate gradient based on product name
function getGradient(name: string): string {
    const gradients = [
        "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
        "linear-gradient(135deg, #f093fb 0%, #f5576c 100%)",
        "linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)",
        "linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)",
        "linear-gradient(135deg, #fa709a 0%, #fee140 100%)",
        "linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)",
        "linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)",
        "linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)",
    ];
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
        hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    return gradients[Math.abs(hash) % gradients.length]!;
}
</script>

<template>
    <div class="product-page">
        <NavBar
            :title="product?.name || '商品详情'"
            left-arrow
            @click-left="router.back()"
        />

        <div v-if="loading" class="loading-container">
            <div class="loading-spinner"></div>
            <p class="loading-text">加载中...</p>
        </div>

        <template v-else-if="product">
            <!-- Gradient header with product name -->
            <div 
                class="product-cover"
                :style="{ background: getGradient(product.name || '商品') }"
            >
                <div class="cover-content">
                    <span class="cover-char">{{ (product.name || '商')[0] }}</span>
                    <h2 class="cover-title">{{ product.name }}</h2>
                    <p v-if="product.noun" class="cover-desc">{{ product.noun }}</p>
                </div>
            </div>

            <div class="product-info-card animate-fade-in-up">
                <div class="price-row">
                    <span class="price-symbol">¥</span>
                    <span class="price-value">{{ product.retail_price }}</span>
                </div>
                <h1 class="product-title">{{ product.name }}</h1>
                <p v-if="product.noun" class="product-desc">{{ product.noun }}</p>
                <div class="product-tags">
                    <span class="tag">正品保证</span>
                    <span class="tag">极速发货</span>
                    <span class="tag">售后无忧</span>
                </div>
            </div>

            <div class="divider"></div>

            <div class="form-section animate-fade-in-up">
                <div class="form-title">填写信息</div>
                <Form>
                    <CellGroup inset>
                        <Field
                            v-model="account"
                            label="账号"
                            placeholder="请输入购买账号"
                            required
                            clearable
                        />
                        <Field
                            v-model="password"
                            label="密码"
                            type="password"
                            placeholder="请输入账号密码"
                            required
                            clearable
                        />
                        <Field
                            v-model="remark"
                            label="备注"
                            type="textarea"
                            rows="2"
                            placeholder="选填，如特殊需求等"
                        />
                    </CellGroup>

                    <!-- 查课按钮 -->
                    <div class="query-wrap">
                        <Button
                            type="default"
                            block
                            round
                            :loading="querying"
                            @click="handleQuery"
                            class="query-btn"
                        >
                            {{ querying ? "查课中..." : "查课" }}
                        </Button>
                    </div>
                </Form>

                <!-- 查课结果 -->
                <div v-if="queryDone && queryResult" class="query-result">
                    <div class="query-result-header">
                        <span class="query-username">{{
                            queryResult.userName || queryResult.userinfo
                        }}</span>
                        <span v-if="queryResult.msg" class="query-msg">{{
                            queryResult.msg
                        }}</span>
                    </div>
                    <div
                        v-if="queryResult.data && queryResult.data.length > 0"
                        class="course-list"
                    >
                        <div
                            v-for="item in queryResult.data"
                            :key="item.id"
                            class="course-item"
                            :class="{ selected: isCourseSelected(item) }"
                            @click="toggleCourse(item)"
                        >
                            <span
                                class="course-select-dot"
                                :class="{ active: isCourseSelected(item) }"
                            ></span>
                            <span class="course-name">{{ item.name }}</span>
                            <span
                                v-if="item.complete"
                                class="course-complete"
                                :class="item.complete === '已完成' ? 'done' : ''"
                                >{{ item.complete }}</span
                            >
                        </div>
                        <div v-if="selectedCourses.length > 0" class="selected-hint">
                            已选 {{ selectedCourses.length }} 门课程
                        </div>
                    </div>
                    <div v-else-if="queryResult.msg === '查询成功' || queryResult.msg === '此课程无需查课，直接下单即可'" class="query-success-direct">无需选课，可直接下单</div>
                    <div v-else class="query-empty">暂无课程数据</div>
                </div>

                <!-- 支付方式 -->
                <div v-if="channels.length > 0" class="pay-section">
                    <div class="form-title">选择支付方式</div>
                    <RadioGroup v-model="payType">
                        <div class="pay-channels">
                            <label
                                v-for="ch in channels"
                                :key="ch.type"
                                class="pay-channel-item"
                                :class="{ active: payType === ch.type }"
                                @click="payType = ch.type"
                            >
                                <img
                                    v-if="payIcons[ch.type]"
                                    :src="payIcons[ch.type]"
                                    class="pay-icon"
                                    alt=""
                                />
                                <span class="pay-name">{{ ch.name }}</span>
                                <Radio :name="ch.type" class="pay-radio" />
                            </label>
                        </div>
                    </RadioGroup>
                </div>

                <div class="submit-wrap">
                    <Button
                        type="primary"
                        block
                        round
                        :loading="submitting"
                        :disabled="!queryDone"
                        class="submit-btn"
                        @click="handleOrder"
                    >
                        {{
                            submitting
                                ? "跳转支付中..."
                                : queryDone
                                  ? `立即支付 ¥${product.retail_price}`
                                  : "请先查课再支付"
                        }}
                    </Button>
                </div>
            </div>

            <div class="bottom-tip safe-area-bottom">
                <van-icon name="shield-check" size="14" color="#94a3b8" />
                <span>安全支付 · 放心购物</span>
            </div>
        </template>
    </div>
</template>

<style scoped>
.product-page {
    min-height: 100vh;
    background: var(--bg-primary);
    padding-bottom: 24px;
}
.product-cover {
    height: 200px;
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    overflow: hidden;
}
.product-cover::before {
    content: '';
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.1);
}
.cover-content {
    position: relative;
    z-index: 1;
    text-align: center;
    padding: 20px;
}
.cover-char {
    display: inline-block;
    width: 80px;
    height: 80px;
    line-height: 80px;
    font-size: 40px;
    font-weight: 700;
    color: rgba(255, 255, 255, 0.95);
    background: rgba(255, 255, 255, 0.2);
    border-radius: 50%;
    backdrop-filter: blur(4px);
    margin-bottom: 12px;
}
.cover-title {
    font-size: 20px;
    font-weight: 600;
    color: #fff;
    margin: 0;
    text-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}
.cover-desc {
    font-size: 13px;
    color: rgba(255, 255, 255, 0.85);
    margin: 8px 0 0;
    max-width: 280px;
}
.product-info-card {
    margin: 0;
    padding: 16px;
    background: var(--bg-secondary);
}
.price-row {
    display: flex;
    align-items: baseline;
    gap: 2px;
    margin-bottom: 8px;
}
.price-symbol {
    font-size: 16px;
    font-weight: 600;
    color: #ef4444;
}
.price-value {
    font-size: 28px;
    font-weight: 700;
    color: #ef4444;
}
.product-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    line-height: 1.5;
    margin-bottom: 8px;
}
.product-desc {
    font-size: 13px;
    color: var(--text-secondary);
    line-height: 1.5;
    margin-bottom: 12px;
}
.product-tags {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
}
.tag {
    background: var(--primary-bg);
    color: var(--primary-color);
    padding: 2px 8px;
    border-radius: var(--radius-full);
    font-size: 12px;
    font-weight: 500;
}
.divider {
    height: 8px;
    background: var(--bg-primary);
}
.form-section {
    background: var(--bg-secondary);
    padding: 16px 0;
}
.form-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-secondary);
    padding: 0 16px 12px;
}
.pay-section {
    margin-top: 12px;
    padding-top: 4px;
}
.pay-channels {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 0 16px;
}
.pay-channel-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px 16px;
    border-radius: 10px;
    border: 1.5px solid var(--border-color, #e5e7eb);
    background: var(--bg-secondary);
    cursor: pointer;
    transition: border-color 0.2s;
}
.pay-channel-item.active {
    border-color: var(--primary-color, #6366f1);
    background: var(--primary-bg, #eef2ff);
}
.pay-icon {
    width: 28px;
    height: 28px;
    object-fit: contain;
}
.pay-name {
    flex: 1;
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
}
.pay-radio {
    pointer-events: none;
}
.submit-wrap {
    padding: 20px 16px 8px;
}
.submit-btn {
    height: 46px !important;
    font-size: 15px !important;
    font-weight: 600 !important;
}
.query-wrap {
    padding: 12px 16px 4px;
}
.query-btn {
    height: 40px !important;
    font-size: 14px !important;
    border: 1.5px solid var(--primary-color, #6366f1) !important;
    color: var(--primary-color, #6366f1) !important;
}
.query-result {
    margin: 8px 16px 4px;
    background: var(--bg-primary);
    border-radius: 10px;
    padding: 12px;
    border: 1px solid var(--border-color, #e5e7eb);
}
.query-result-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
}
.query-username {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
}
.query-msg {
    font-size: 12px;
    color: var(--text-secondary);
}
.course-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
}
.course-item {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 14px;
    color: var(--text-primary);
    padding: 12px;
    border-radius: 8px;
    border: 1.5px solid transparent;
    cursor: pointer;
    transition:
        border-color 0.15s,
        background 0.15s;
}
.course-item.selected {
    border-color: var(--primary-color, #6366f1);
    background: var(--primary-bg, #eef2ff);
}
.course-select-dot {
    flex-shrink: 0;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    border: 2px solid var(--border-color, #d1d5db);
    background: #fff;
    transition:
        border-color 0.15s,
        background 0.15s;
}
.course-select-dot.active {
    border-color: var(--primary-color, #6366f1);
    background: var(--primary-color, #6366f1);
    box-shadow: inset 0 0 0 3.5px #fff;
}
.course-name {
    flex: 1;
}
.course-complete {
    font-size: 12px;
    color: var(--text-muted);
    white-space: nowrap;
    margin-left: 8px;
}
.course-complete.done {
    color: #22c55e;
}
.selected-hint {
    font-size: 12px;
    color: var(--primary-color, #6366f1);
    padding: 6px 12px 2px;
    font-weight: 500;
}
.query-success-direct {
    font-size: 13px;
    color: #22c55e;
    text-align: center;
    padding: 8px 0;
    font-weight: 600;
}
.query-empty {
    font-size: 13px;
    color: var(--text-muted);
    text-align: center;
    padding: 8px 0;
}
.bottom-tip {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 16px;
    color: var(--text-muted);
    font-size: 12px;
}
</style>
