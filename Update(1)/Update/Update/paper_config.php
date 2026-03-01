<?php
$title = '论文配置';
include('head.php');

if($userrow['uid']!=1){
    alert("你来错地方了","index.php");
}
?>
<div class="app-content-body">
    <div class="wrapper-md control" id="PaperConfig">
        <div class="panel panel-default" style="box-shadow: 3px 3px 8px #d1d9e6, -3px -3px 8px #d1d9e6;border-radius: 6px;">
            <div class="panel-body">
                <form class="form-horizontal devform" id="form-web">
                    <div class="card">
                        <ul class="nav nav-tabs" role="tablist">
                            <li class="active">
                                <a data-toggle="tab" href="#apis">接口配置</a>
                            </li>
                            <li class="nav-item">
                                <a data-toggle="tab" href="#price">价格配置(实际扣费为本页面价格*用户费率)</a>
                            </li>
                        </ul>
                        <div class="tab-content">
                            <div class="tab-pane fade active in" id="apis">
                                <div class="form-group">
                                    <label class="col-sm-2 control-label">登录账号</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_username" value="<?= $conf['lunwen_api_username'] ?>" placeholder="请输入登录账号" required>
                                    </div>
                                </div>

                                <div class="form-group">
                                    <label class="col-sm-2 control-label">登录密码</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_password" value="<?= $conf['lunwen_api_password'] ?>" placeholder="请输入登录密码" required>
                                    </div>
                                </div>
                            </div>
                            <div class="tab-pane fade" id="price">
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">论文6000字价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_6000_price" value="<?= $conf['lunwen_api_6000_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">论文8000字价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_8000_price" value="<?= $conf['lunwen_api_8000_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">论文10000字价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_10000_price" value="<?= $conf['lunwen_api_10000_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">论文12000字价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_12000_price" value="<?= $conf['lunwen_api_12000_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">论文15000字价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_15000_price" value="<?= $conf['lunwen_api_15000_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">任务书价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_rws_price" value="<?= $conf['lunwen_api_rws_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">开题报告价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_ktbg_price" value="<?= $conf['lunwen_api_ktbg_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">降低AIGC痕迹价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_jdaigchj_price" value="<?= $conf['lunwen_api_jdaigchj_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">修改段落千字价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_xgdl_price" value="<?= $conf['lunwen_api_xgdl_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">降重率千字价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_jcl_price" value="<?= $conf['lunwen_api_jcl_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                                <div class="form-group">
                                    <label class="col-sm-3 control-label">降低AIGC率千字价格</label>
                                    <div class="col-sm-9">
                                        <input type="text" class="layui-input" name="lunwen_api_jdaigcl_price" value="<?= $conf['lunwen_api_jdaigcl_price'] ?>" placeholder="请输入价格" required>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="col-sm-offset-2 col-sm-4">
                        <input type="button" @click="onSubmit" value="立即修改" class="layui-btn" />
                    </div>

                </form>


            </div>
        </div>
    </div>
</div>


<?php require_once("footer.php"); ?>


<script type="text/javascript">
    var app = new Vue({
        el: "#PaperConfig",
        data: {

        },
        methods: {
            //提交
            onSubmit() {
                var loading = layer.load(2);
                this.$http.post("/index/aisdk/config.php?act=webset", {
                    data: $("#form-web").serialize()
                }, {
                    emulateJSON: true
                }).then(function(data) {
                    layer.close(loading);
                    if (data.data.code == 1) {
                        layer.alert(data.data.msg, {
                            icon: 1,
                            title: "温馨提示"
                        }, function() {
                            setTimeout(function() {
                                window.location.href = ""
                            });
                        });
                    } else {
                        layer.alert(data.data.msg, {
                            icon: 2,
                            title: "温馨提示"
                        });
                    }
                });
            },
        },
        mounted() {

        }
    });
</script>