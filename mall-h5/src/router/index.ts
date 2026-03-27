import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
    history: createWebHistory("/mall/"),
    routes: [
        {
            path: "/",
            component: () => import("../views/Layout.vue"),
            children: [
                {
                    path: "",
                    component: () => import("../views/Home.vue"),
                },
                {
                    path: "product/:cid",
                    component: () => import("../views/Product.vue"),
                },
                {
                    path: "orders",
                    component: () => import("../views/Orders.vue"),
                },
                {
                    path: "login",
                    component: () => import("../views/Login.vue"),
                },
                {
                    path: "register",
                    component: () => import("../views/Register.vue"),
                },
                {
                    path: "query",
                    component: () => import("../views/Query.vue"),
                },
                {
                    path: "mine",
                    component: () => import("../views/Mine.vue"),
                },
                {
                    path: "promotion",
                    component: () => import("../views/Promotion.vue"),
                },
                {
                    path: "withdraw",
                    component: () => import("../views/Withdraw.vue"),
                },
                {
                    path: "pay-result",
                    component: () => import("../views/PayResult.vue"),
                },
            ],
        },
        {
            path: "/:tid",
            component: () => import("../views/Layout.vue"),
            children: [
                {
                    path: "",
                    name: "Home",
                    component: () => import("../views/Home.vue"),
                },
                {
                    path: "product/:cid",
                    name: "Product",
                    component: () => import("../views/Product.vue"),
                },

                {
                    path: "orders",
                    name: "Orders",
                    component: () => import("../views/Orders.vue"),
                },
                {
                    path: "login",
                    name: "Login",
                    component: () => import("../views/Login.vue"),
                },
                {
                    path: "register",
                    name: "Register",
                    component: () => import("../views/Register.vue"),
                },
                {
                    path: "query",
                    name: "Query",
                    component: () => import("../views/Query.vue"),
                },
                {
                    path: "mine",
                    name: "Mine",
                    component: () => import("../views/Mine.vue"),
                },
                {
                    path: "promotion",
                    name: "Promotion",
                    component: () => import("../views/Promotion.vue"),
                },
                {
                    path: "withdraw",
                    name: "Withdraw",
                    component: () => import("../views/Withdraw.vue"),
                },
                {
                    path: "pay-result",
                    name: "PayResult",
                    component: () => import("../views/PayResult.vue"),
                },
            ],
        },
        {
            path: "/:pathMatch(.*)*",
            component: () => import("../views/NotFound.vue"),
        },
    ],
});

export default router;
