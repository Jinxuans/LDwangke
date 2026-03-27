<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { NavBar, NoticeBar, Empty, showToast } from "vant";
import { getShopInfo, getProducts, triggerCustomerService } from "../api";

const route = useRoute();
const router = useRouter();
const tid = String(route.params.tid || "");
const basePath = tid ? `/${tid}` : "";
const shop = ref<any>(null);
const products = ref<any[]>([]);
const loading = ref(true);
const activeCategory = ref("全部");
const showCategories = computed(() => shop.value?.mall_config?.show_categories !== false);
const popupVisible = ref(false);
const popupHtml = computed(() => String(shop.value?.mall_config?.popup_notice_html || "").trim());
const customerService = computed(() => shop.value?.mall_config?.customer_service || null);
const customerPopup = ref<null | { title: string; label: string; type: string; value: string }>(null);

// Extract unique categories from products
const categories = computed(() => {
    const cats = new Set<string>();
    products.value.forEach(p => {
        const category = p.fenlei_name || p.fenlei;
        if (category) cats.add(category);
    });
    return ["全部", ...Array.from(cats)];
});

// Filter products by selected category
const filteredProducts = computed(() => {
    if (!showCategories.value) return products.value;
    if (activeCategory.value === "全部") return products.value;
    return products.value.filter(p => (p.fenlei_name || p.fenlei) === activeCategory.value);
});

// Generate a color based on name
function getAvatarColor(name: string): string {
    const colors = [
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
    return colors[Math.abs(hash) % colors.length]!;
}

onMounted(async () => {
    try {
        const [shopRes, prodRes] = await Promise.all([
            getShopInfo(),
            getProducts(),
        ]);
        shop.value = shopRes;
        document.title = shopRes?.shop_name || "精选商城";
        products.value = Array.isArray(prodRes) ? prodRes : [];
        const popupKey = tid ? `mall_popup_notice_seen:${tid}` : `mall_popup_notice_seen:${window.location.host}`;
        if (popupHtml.value && !sessionStorage.getItem(popupKey)) {
            sessionStorage.setItem(popupKey, "1");
            popupVisible.value = true;
        }
    } catch (e) {
        console.error(e);
    } finally {
        loading.value = false;
    }
});

async function contactService() {
    const res = await triggerCustomerService(customerService.value || undefined);
    if ((res as any)?.mode === "popup") {
        customerPopup.value = res as any;
        return;
    }
    showToast((res as any)?.message || "操作失败");
}

async function copyCustomerValue() {
    if (!customerPopup.value?.value) return;
    await navigator.clipboard.writeText(customerPopup.value.value);
    showToast(customerPopup.value.type === "qq" ? "QQ号已复制" : customerPopup.value.type === "phone" ? "电话号码已复制" : "客服微信已复制");
}

function callCustomerPhone() {
    if (!customerPopup.value?.value) return;
    window.location.href = `tel:${customerPopup.value.value}`;
}
</script>

<template>
    <div class="home-page">
        <NavBar :title="shop?.shop_name || '精选商城'" class="home-nav">
            <template #right>
                <span
                    class="nav-btn"
                    @click="router.push(`${basePath}/query`)"
                >
                    <van-icon name="search" size="20" />
                </span>
            </template>
        </NavBar>

        <div class="shop-bar animate-fade-in-down">
            <van-image
                round
                width="40"
                height="40"
                :src="
                    shop?.shop_logo ||
                    'https://fastly.jsdelivr.net/npm/@vant/assets/cat.jpeg'
                "
            />
            <div class="shop-info">
                <span class="shop-name">{{ shop?.shop_name || "精选商城" }}</span>
                <span class="shop-sub">好物优选 · 品质保证</span>
            </div>
            <van-button
                size="mini"
                plain
                type="primary"
                round
                @click="router.push(`${basePath}/query`)"
                >查进度</van-button
            >
        </div>

        <NoticeBar
            v-if="shop?.shop_desc"
            :text="shop.shop_desc"
            left-icon="volume-o"
            scrollable
            class="shop-notice animate-fade-in-up"
        />

        <div v-if="customerService?.enabled" class="service-card animate-fade-in-up">
            <div>
                <div class="service-title">有问题？联系在线客服</div>
                <div class="service-desc">支付、进度、售后问题都可以直接咨询。</div>
            </div>
            <van-button size="small" type="primary" round @click="contactService">
                {{ customerService?.label || "联系客服" }}
            </van-button>
        </div>

        <!-- Category Tabs -->
        <div v-if="showCategories && products.length > 0" class="category-tabs">
            <div class="category-scroll">
                <span
                    v-for="cat in categories"
                    :key="cat"
                    class="category-tab"
                    :class="{ active: activeCategory === cat }"
                    @click="activeCategory = cat"
                >
                    {{ cat }}
                </span>
            </div>
        </div>

        <div class="section-header">
            <span class="section-title">{{ activeCategory === '全部' ? '全部商品' : activeCategory }}</span>
            <span class="section-count" v-if="filteredProducts.length"
                >共 {{ filteredProducts.length }} 件</span
            >
        </div>

        <div v-if="loading" class="loading-container">
            <div class="loading-spinner"></div>
            <p class="loading-text">加载中...</p>
        </div>
        <Empty
            v-else-if="!filteredProducts.length"
            :description="activeCategory === '全部' ? '暂无商品' : '该分类暂无商品'"
            style="padding-top: 80px"
        />
        <div v-else class="product-grid">
            <div
                v-for="(p, i) in filteredProducts"
                :key="p.cid"
                class="product-card animate-fade-in-up"
                :style="{ animationDelay: `${i * 0.05}s` }"
                @click="router.push(`${basePath}/product/${p.cid}`)"
            >
                <div class="card-img-wrap">
                    <img
                        v-if="p.cover_url"
                        :src="p.cover_url"
                        :alt="p.name"
                        class="product-cover-image"
                    />
                    <div
                        v-else
                        class="product-placeholder"
                        :style="{ background: getAvatarColor(p.name || '商品') }"
                    >
                        <span class="placeholder-char">{{ (p.name || '商')[0] }}</span>
                    </div>
                </div>
                <div class="card-body">
                    <div v-if="p.fenlei_name || p.fenlei" class="card-category">
                        {{ p.fenlei_name || p.fenlei }}
                    </div>
                    <p class="card-name">{{ p.name }}</p>
                    <p v-if="p.description || p.noun" class="card-desc">
                        {{ p.description || p.noun }}
                    </p>
                    <div class="card-footer">
                        <span class="price-tag">¥{{ p.retail_price }}</span>
                        <van-icon name="arrow" size="12" color="#94a3b8" />
                    </div>
                </div>
            </div>
        </div>
        <div class="page-bottom safe-area-bottom"></div>

        <div v-if="popupVisible" class="notice-popup-mask" @click.self="popupVisible = false">
            <div class="notice-popup animate-fade-in-up">
                <div class="notice-popup-header">
                    <span>公告</span>
                    <button class="notice-popup-close" @click="popupVisible = false">×</button>
                </div>
                <div class="notice-popup-body" v-html="popupHtml"></div>
                <van-button type="primary" block round class="notice-popup-btn" @click="popupVisible = false">
                    我知道了
                </van-button>
            </div>
        </div>

        <div v-if="customerPopup" class="notice-popup-mask" @click.self="customerPopup = null">
            <div class="notice-popup animate-fade-in-up">
                <div class="notice-popup-header">
                    <span>{{ customerPopup.title }}</span>
                    <button class="notice-popup-close" @click="customerPopup = null">×</button>
                </div>
                <div class="notice-popup-body">
                    <div class="customer-value">{{ customerPopup.value }}</div>
                    <div class="customer-tip">若无法直接跳转，可手动复制后联系。</div>
                </div>
                <div class="customer-actions">
                    <van-button plain round block @click="copyCustomerValue">复制{{ customerPopup.type === 'phone' ? '号码' : '账号' }}</van-button>
                    <van-button
                        v-if="customerPopup.type === 'phone'"
                        type="primary"
                        round
                        block
                        @click="callCustomerPhone"
                    >
                        立即拨号
                    </van-button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.home-page {
    min-height: 100vh;
    background: var(--bg-primary);
    padding-bottom: 60px;
}
.home-nav {
    background: var(--bg-secondary) !important;
}
.nav-btn {
    display: flex;
    align-items: center;
    padding: 6px;
    color: var(--text-secondary);
    cursor: pointer;
}
.nav-btn:active {
    color: var(--primary-color);
}
.shop-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 16px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
}
.shop-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
}
.shop-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.shop-sub {
    font-size: 12px;
    color: var(--text-muted);
}
.shop-notice {
    margin: 12px 12px 0;
    border-radius: 14px;
    overflow: hidden;
}
.service-card {
    margin: 12px 12px 0;
    padding: 14px 14px;
    border-radius: 16px;
    background: linear-gradient(135deg, #eff6ff 0%, #ffffff 100%);
    border: 1px solid #bfdbfe;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
}
.service-title {
    font-size: 14px;
    font-weight: 700;
    color: #1e3a8a;
}
.service-desc {
    margin-top: 4px;
    font-size: 12px;
    color: #64748b;
}
.category-tabs {
    background: var(--bg-secondary);
    padding: 8px 0;
    border-bottom: 1px solid var(--border-light);
}
.category-scroll {
    display: flex;
    gap: 8px;
    padding: 0 12px;
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
}
.category-scroll::-webkit-scrollbar {
    display: none;
}
.category-tab {
    flex-shrink: 0;
    padding: 6px 14px;
    font-size: 13px;
    color: var(--text-secondary);
    background: var(--bg-primary);
    border-radius: var(--radius-full);
    cursor: pointer;
    transition: all 0.2s ease;
    border: 1px solid var(--border-color);
}
.category-tab.active {
    background: var(--primary-color);
    color: #fff;
    border-color: var(--primary-color);
}
.section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px 10px;
}
.section-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
}
.section-count {
    font-size: 12px;
    color: var(--text-muted);
}
.product-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
    padding: 0 12px;
}
.product-card {
    background: var(--bg-secondary);
    border-radius: var(--radius-lg);
    overflow: hidden;
    border: 1px solid var(--border-color);
    cursor: pointer;
    transition:
        box-shadow 0.2s ease,
        transform 0.15s ease;
}
.product-card:active {
    transform: scale(0.97);
    box-shadow: var(--shadow-md);
}
.card-img-wrap {
    overflow: hidden;
    background: var(--bg-primary);
}
.product-cover-image {
    width: 100%;
    height: 150px;
    object-fit: cover;
    display: block;
}
.product-placeholder {
    width: 100%;
    height: 150px;
    display: flex;
    align-items: center;
    justify-content: center;
}
.placeholder-char {
    font-size: 48px;
    font-weight: 700;
    color: rgba(255, 255, 255, 0.9);
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
.card-body {
    padding: 10px 10px 12px;
}
.card-category {
    display: inline-flex;
    max-width: 100%;
    padding: 2px 8px;
    margin-bottom: 8px;
    border-radius: 999px;
    background: rgba(99, 102, 241, 0.08);
    color: var(--primary-color);
    font-size: 11px;
    font-weight: 600;
}
.card-name {
    font-size: 13px;
    color: var(--text-primary);
    line-height: 1.4;
    margin-bottom: 8px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}
.card-desc {
    font-size: 12px;
    color: var(--text-muted);
    line-height: 1.5;
    margin-bottom: 10px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}
.card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.page-bottom {
    height: 24px;
}
.notice-popup-mask {
    position: fixed;
    inset: 0;
    z-index: 50;
    background: rgba(15, 23, 42, 0.45);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
}
.notice-popup {
    width: min(100%, 420px);
    max-height: 80vh;
    overflow: hidden;
    border-radius: 20px;
    background: #fff;
    box-shadow: 0 20px 50px rgba(15, 23, 42, 0.24);
    display: flex;
    flex-direction: column;
}
.notice-popup-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 18px 12px;
    border-bottom: 1px solid #e2e8f0;
    font-size: 16px;
    font-weight: 700;
    color: #0f172a;
}
.notice-popup-close {
    border: none;
    background: transparent;
    font-size: 24px;
    line-height: 1;
    color: #64748b;
}
.notice-popup-body {
    padding: 16px 18px;
    overflow: auto;
    font-size: 14px;
    line-height: 1.8;
    color: #334155;
}
.notice-popup-btn {
    margin: 0 18px 18px;
}
.customer-value {
    padding: 12px 14px;
    border-radius: 14px;
    background: #f8fafc;
    color: #0f172a;
    font-size: 15px;
    font-weight: 700;
    word-break: break-all;
}
.customer-tip {
    margin-top: 10px;
    font-size: 12px;
    color: #64748b;
}
.customer-actions {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 0 18px 18px;
}
</style>
