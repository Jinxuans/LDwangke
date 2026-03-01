<?php
$title = '论文下单';
include('head.php');
?>

<link href="css/element.css?v=1.4" rel="stylesheet">
<link href="css/tailwind.css?v=1.11" rel="stylesheet">
<div class="app-content-body" style="padding: 15px;" id="PaperDedup">
    <el-card class="box-card" style="margin-left:10px;">
        <div slot="header">
            <div class="flex justify-between items-center">
                <div>
                    论文降重服务
                </div>
                <div>
                    <el-radio-group v-model="navigationActive" @input="navigationActiveInput">
                        <el-radio-button :label="1">文件降重</el-radio-button>
                        <el-radio-button :label="2">文本降重</el-radio-button>

                    </el-radio-group>
                </div>
                <div>
                    <el-tag
                        type="info"
                        effect="plain">
                        降重单价：<?= $conf['lunwen_api_jcl_price'] * $userrow['addprice'] ?>元/千字
                    </el-tag>
                    <el-tag
                        type="info"
                        effect="plain">
                        AI降重单价：<?= $conf['lunwen_api_jcl_price'] * $userrow['addprice'] ?>元/千字
                    </el-tag>
                </div>
            </div>
        </div>

        <div v-if="navigationActive == 1">
            <el-upload
                class="upload-components"
                drag
                action="aisdk/http.php?act=countWords"
                accept=".docx"
                :limit="1"
                :on-change="handleFileChange">
                <i class="el-icon-upload"></i>

                <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
                <div class="el-upload__tip" slot="tip">只能上传 docx 格式文件<br />文件大小不能超过 20MB<br /><br />注意：降重内容由大模型生成，生成的内容不代表本平台观点<br />降重过程中会保留原文档的格式，但部分复杂格式无法保留<br />大模型生成的内容可能会存在部分错误，这是正常现象，生成内容请自行检查<br />生成的内容可能会对原文内容长度进行增减</div>
            </el-upload>
            <div v-if="isFileInfoShow">
                <el-descriptions class="mt-6" title="文件信息" :column="1" border>

                    <el-descriptions-item label="文件名称">
                        {{fileName}}
                    </el-descriptions-item>
                    <el-descriptions-item label="文件大小">
                        {{fileSize}}
                    </el-descriptions-item>
                    <el-descriptions-item label="字数统计">
                        {{fileForm.wordCount}}
                    </el-descriptions-item>
                </el-descriptions>
                <el-row class="border p-16 my-5">
                    <el-col :span="10">
                        论文降重
                    </el-col>
                    <el-col :span="2">
                        ¥<?= $conf['lunwen_api_jcl_price'] * $userrow['addprice'] ?>
                    </el-col>
                    <el-col :span="3">
                        <el-switch
                            v-model="fileForm.jiangchong"
                            :inactive-value="0"
                            :active-value="1">
                        </el-switch>
                    </el-col>
                    <el-col :span="2">
                        <el-tag type="info"><?= $conf['lunwen_api_jcl_price'] * $userrow['addprice'] ?>元/千字</el-tag>
                    </el-col>
                </el-row>
                <el-row class="border p-16 my-5">
                    <el-col :span="10">
                        降低AIGC痕迹
                    </el-col>
                    <el-col :span="2">
                        ¥<?= $conf['lunwen_api_jdaigcl_price'] * $userrow['addprice'] ?>
                    </el-col>
                    <el-col :span="3">
                        <el-switch
                            v-model="fileForm.aigc"
                            :inactive-value="0"
                            :active-value="1">
                        </el-switch>
                    </el-col>
                    <el-col :span="2">
                        <el-tag type="info"><?= $conf['lunwen_api_jdaigcl_price'] * $userrow['addprice'] ?>元/千字</el-tag>
                    </el-col>
                </el-row>
            </div>
            <div class="flex justify-center mt-10">
                <el-button type="primary" @click="fileSubmit" :disabled="fileForm.jiangchong == 0 && fileForm.aigc == 0" :loading="fileSubmitLoad">提交降重</el-button>
            </div>
        </div>
        <div v-if="navigationActive == 2">
            <el-form ref="form" :model="form" label-width="80px" size="medium">
                <el-form-item label="原文内容">
                    <el-input type="textarea" v-model="form.content" :rows="8" @input="contentInput" placeholder="在使用降低AIGC率的时候，尽可能多粘贴一些内容（250字以上），以便大模型生成效果更好的内容，尽量不要同时粘贴多个段落"></el-input>
                </el-form-item>
                <el-form-item>
                    <el-button plain size="small">字数统计：{{contentCount}}字</el-button>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="handleTextPaperRewrite('textPaperRewrite')" :disabled="form.content==''" :loading="textPaperRewriteLoad">降低重复率</el-button>
                    <el-button type="success" @click="handleTextPaperRewrite('textRewriteAigc')" :disabled="form.content==''" :loading="textPaperRewriteLoad">降低AIGC率</el-button>
                    <el-button type="warning" @click="clearContent" :disabled="form.content==''">清空内容</el-button>
                </el-form-item>
            </el-form>
        </div>
        <div v-if="isResultContentShow">
            <div class="processing-status">
                <div class="status-text">{{statusText}}</div>
            </div>
            <div class="">
                <div class="flex justify-between items-center mb-6"><span>降重结果</span><el-link type="primary" @click="copyResult">复制结果</el-link></div>
                <el-input type="textarea" v-model="resultContent" :rows="8" :readonly="true" v-copy></el-input>
            </div>
        </div>
    </el-card>


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
        el: "#PaperDedup",
        data: {
            navigationActive: 1,
            contentCount: 0,
            resultContent: '',
            statusText: '',
            isResultContentShow: false,
            fileName: '',
            fileSize: '',
            isFileInfoShow: false,
            fileForm: {
                file: '',
                wordCount: 0,
                aigc: 0,
                jiangchong: 0
            },
            form: {
                content: ''
            },
            textPaperRewriteLoad: false,
            fileSubmitLoad: false,
        },
        methods: {
            handleFileChange(file, fileList) {
                const fileName = file.name;
                const fileSizeKb = (file.size / 1024).toFixed(2);
                this.fileName = fileName;
                this.fileSize = fileSizeKb + 'KB';
                if (file.response && file.response.code == 200) {
                    this.fileForm.wordCount = file.response.data;
                    this.isFileInfoShow = true;
                    this.fileForm.file = file.raw;
                }
            },
            // 文件降重提交（处理普通JSON响应）
            fileSubmit() {
                this.fileSubmitLoad = true;
                const formData = new FormData();
                formData.append('file', this.fileForm.file);
                formData.append('wordCount', this.fileForm.wordCount);
                formData.append('aigc', this.fileForm.aigc);
                formData.append('jiangchong', this.fileForm.jiangchong);

                this.$http.post("aisdk/http.php?act=fileDedup", formData, {
                    headers: {
                        'Content-Type': 'multipart/form-data'
                    }
                }).then((res) => {
                    this.fileSubmitLoad = false;
                    console.log("文件降重接口响应:", res.body);
                    
                    if (res.body.code === 200) {
                        this.isFileInfoShow = false;
                        this.$message.success(res.body.msg);
                    } else {
                        const errorCode = res.body.code;
                        const errorMsg = res.body.data || res.body.msg || "接口返回异常";
                        
                        if ((errorCode === -1 || errorCode === 500) && errorMsg.includes("余额不足")) {
                            this.$message.error("源台余额不足，请联系管理员");
                        } else {
                            this.$message.error(`操作失败 (${errorCode}): ${errorMsg}`);
                        }
                    }
                }).catch((err) => {
                    this.fileSubmitLoad = false;
                    this.$message.error(`网络请求失败: ${err.message}`);
                    console.error("文件降重接口异常:", err);
                });
            },
            contentInput(value) {
                this.contentCount = value.length;
            },
            clearContent() {
                this.form.content = '';
                this.contentCount = 0;
            },
            // 文本降重（处理SSE流式响应）
            async handleTextPaperRewrite(act) {
                const content = this.form.content;
                this.resultContent = '';
                this.textPaperRewriteLoad = true;
                
                try {
                    const response = await fetch('aisdk/http.php?act=' + act, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ content })
                    });
                    
                    if (!response.ok) {
                        throw new Error(`HTTP错误: ${response.status}`);
                    }
                    
                    // 先检查JSON响应错误
                    let responseJson;
                    try {
                        responseJson = await response.clone().json();
                        if (responseJson.code && responseJson.code !== 200) {
                            const errorMsg = responseJson.data || responseJson.msg || "未知错误";
                            if ((responseJson.code === -1 || responseJson.code === 500) && errorMsg.includes("余额不足")) {
                                this.$message.error("源台余额不足，请联系管理员");
                            } else {
                                this.$message.error(`接口错误: ${responseJson.code}\n${errorMsg}`);
                            }
                            return;
                        }
                    } catch (jsonError) {
                        console.log("非JSON响应，处理流式数据");
                    }
                    
                    // 处理流式响应
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
                            
                            // 处理流式错误事件
                            if (eventData.event === 'error') {
                                const errorMsg = JSON.parse(eventData.data);
                                if (errorMsg.includes("余额不足")) {
                                    this.$message.error("源台余额不足，请联系管理员");
                                    this.textPaperRewriteLoad = false;
                                    this.isResultContentShow = false;
                                    return;
                                } else {
                                    this.$message.error(`接口错误: ${errorMsg}`);
                                    this.textPaperRewriteLoad = false;
                                    this.isResultContentShow = false;
                                    return;
                                }
                            }
                            
                            // 处理正常数据
                            if (eventData.event === 'chunk') {
                                this.resultContent += JSON.parse(eventData.data);
                            } else if (eventData.event === 'status') {
                                this.statusText = JSON.parse(eventData.data);
                            }
                        }
                    }
                    
                } catch (error) {
                    this.textPaperRewriteLoad = false;
                    if (error.name !== 'AbortError') {
                        this.$message.error(error.message);
                        console.error("文本降重错误:", error);
                    }
                } finally {
                    this.textPaperRewriteLoad = false;
                }
            },
            copyResult() {
                copyToClipboard(this.resultContent, this);
            },
            navigationActiveInput(value) {
                if (value == 1) {
                    this.form = { content: '' };
                    this.isResultContentShow = false;
                    this.resultContent = '';
                    this.contentCount = 0;
                } else {
                    this.fileName = '';
                    this.fileSize = '';
                    this.isFileInfoShow = false;
                    this.fileForm = { file: '', wordCount: 0, aigc: 0, jiangchong: 0 };
                }
            }
        },
        mounted() {}
    });
</script>
<style>
    .upload-components .el-upload {
        width: 100%;
    }

    .upload-components .el-upload-dragger {
        width: 100%;
    }

    .upload-components input {
        display: none;
    }

    .el-descriptions__table th {
        width: 100px;
    }
</style>