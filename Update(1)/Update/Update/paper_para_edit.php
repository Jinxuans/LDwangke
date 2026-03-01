<?php
$title = '段落修改';
include('head.php');
?>
<link href="css/element.css" rel="stylesheet">
<link href="css/tailwind.css?v=1.4" rel="stylesheet">
<div id="content" class="wrapper-md control" role="main">
    <div class="app-content-body" style="padding: 15px;" id="PaperParaEdit">
        <el-card class="box-card" style="margin-left:10px">
            <div slot="header" class="flex justify-between">
                <span>文本段落修改</span>
                <el-tag
                    type="info"
                    effect="plain">
                    降重单价：<?= $conf['lunwen_api_xgdl_price'] * $userrow['addprice'] ?>元/千字
                </el-tag>
            </div>
            <div>
                <el-form ref="form" :model="form" label-width="80px" size="medium">
                    <el-form-item label="原文内容">
                        <el-input type="textarea" v-model="form.content" :rows="5" @input="contentInput"></el-input>
                    </el-form-item>
                    <el-form-item label="修改意见">
                        <el-input type="textarea" v-model="form.yijian" :rows="5"></el-input>
                    </el-form-item>
                    <el-form-item>
                        <el-button plain size="small">字数统计：{{contentCount}}字</el-button>
                        <el-button type="danger" plain size="small">文本字数不能少于100字</el-button>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" @click="onSubmit" :disabled="form.content==''" :loading="onSubmitLoad">提交修改</el-button>
                    </el-form-item>
                </el-form>
            </div>
            <div v-if="isResultContentShow">
                <div class="processing-status">
                    <div class="status-text">{{statusText}}</div>
                </div>
                <div class="">
                    <div class="flex justify-between items-center mb-6"><span>修改结果</span><el-link type="primary" @click="copyResult">复制结果</el-link></div>
                    <el-input type="textarea" v-model="resultContent" :rows="8" :readonly="true" v-copy></el-input>
                </div>
            </div>
        </el-card>
    </div>
</div>

<script src="js/vue.min.js"></script>
<script src="js/vue-resource.min.js"></script>
<script src="js/element.js"></script>

<script type="text/javascript">
    // 复制到剪贴板函数
    function copyToClipboard(text, vm) {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        document.body.appendChild(textarea);
        textarea.select();
        try {
            const successful = document.execCommand('copy');
            const msg = successful ? '复制成功!' : '复制失败!';
            vm.$message.success(msg);
        } catch (err) {}

        document.body.removeChild(textarea);
    }
    var app = new Vue({
        el: "#PaperParaEdit",
        data: {
            resultContent: '', //结果内容
            statusText: '', //状态文字
            isResultContentShow: false, //显示结果
            contentCount: 0, //字数统计
            form: {
                content: '',
                delay: 50,
                yijian: '',
            },
            onSubmitLoad: false,
        },
        methods: {
            //原文内容
            contentInput(value) {
                this.contentCount = value.length;
            },
            //提交
            async onSubmit() {
                const form = this.form;
                this.error = null;
                this.onSubmitLoad = true;

                try {
                    const response = await fetch('aisdk/http.php?act=paperParaEdit', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(form),
                    });

                    if (!response.ok) {
                        throw new Error(`HTTP错误: ${response.status}`);
                    }

                    this.isResultContentShow = true;
                    const reader = response.body.getReader();
                    const decoder = new TextDecoder('utf-8');
                    let buffer = '';

                    while (true) {
                        const { done, value } = await reader.read();
                        if (done) break;

                        buffer += decoder.decode(value);
                        const events = buffer.split('\n\n');
                        buffer = events.pop() || '';

                        for (const event of events) {
                            const lines = event.split('\n');
                            const eventData = {};
                            
                            lines.forEach(line => {
                                const [key, val] = line.split(':');
                                if (key) eventData[key.trim()] = val.trim();
                            });

                            // 错误处理
                            if (eventData.event === 'error') {
                                const errorMsg = JSON.parse(eventData.data);
                                console.log("【接口错误】", errorMsg);
                                
                                // 余额不足错误处理
                                if (errorMsg.includes("余额不足")) {
                                    this.$message.error("源台余额不足，请联系管理员");
                                    this.onSubmitLoad = false;
                                    this.isResultContentShow = false; // 可选：隐藏结果区域
                                    return;
                                } else {
                                    this.$message.error(`接口错误: ${errorMsg}`);
                                    this.onSubmitLoad = false;
                                    this.isResultContentShow = false; // 可选：隐藏结果区域
                                    return;
                                }
                            }

                            // 正常数据处理
                            if (eventData.event === 'chunk') {
                                this.resultContent += JSON.parse(eventData.data);
                            } else if (eventData.event === 'status') {
                                this.statusText = JSON.parse(eventData.data);
                            }
                        }
                    }

                } catch (error) {
                    this.onSubmitLoad = false;
                    if (error.name !== 'AbortError') {
                        console.error('请求失败:', error);
                        this.$message.error(error.message);
                    }
                } finally {
                    this.onSubmitLoad = false;
                }
            },
            //复制结果
            copyResult() {
                copyToClipboard(this.resultContent, this);
            }
        },
        mounted() {

        }
    });
</script>