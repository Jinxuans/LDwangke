安装或更新命令行工具
最新版本 0616
执行命令前一定要更新到最新版本
bash
wget http://122.51.236.86/long -O long && chmod +x long && mv -f long /usr/bin/long
同步命令 (sync)
可编辑
输入 long sync -h 查看完整参数

同步源台平台的所有平台，支持增量/全量同步，支持倍率加价，自定义商品名称前缀，商品分类，排序等功能
提示：下方命令可直接在线编辑，编辑后复制到服务器执行
bash
long sync \
    --long-host=122.51.236.86 \
    --access-key=86476920-bd65-4916-94fd-35ba73d855df \
    --mysql-user=数据库用户名 \
    --mysql-password=数据库密码 \
    --mysql-database=数据库名
监听命令 (listen)
可编辑
输入 long listen -h 查看完整参数

监听源台订单变动，实时更新二开平台订单状态，实现真正意义的实时进度展示
如果不是第一次运行，需要先终止之前运行的程序: pkill -f long
bash
nohup long listen \
    --long-host=122.51.236.86 \
    --access-key=86476920-bd65-4916-94fd-35ba73d855df \
    --mysql-user=数据库用户名 \
    --mysql-password=数据库密码 \
    --mysql-database=数据库名 \
    > log.txt 2>&1 &
其他参数说明
参数
说明
默认值
--mysql-host	MySQL服务器地址	127.0.0.1
--mysql-port	MySQL服务器端口	3306
--mysql-user	MySQL用户名	无（必填）
--mysql-password	MySQL密码	无（必填）
--mysql-database	MySQL数据库名	无（必填）
--class-table	分类表名 如qingka_wangke_class	若不填自动获取_class后缀表
--order-table	订单表名 例如qingka_wangke_order	若不填自动获取_order后缀表
--docking	网课接口配置的ID,huoyuan表的hid,也可以在网课接口配置页面可以看到	如果不指定,自动获取huoyuan表pass=access-key的第一条记录的hid
--rate	价格比例（商品价格 = 成本价 × 此值）如果你平台余额不是1:1那么还应该乘以你的系数	1.5
--name-prefix	商品名称前缀	无
--category	商品分类ID	无
--cover-price	是否覆盖价格 指定为true则会将二开价格重新按照rate设置	false
--cover-desc	是否覆盖描述 指定为true则会将源台的商品描述设置为二开的描述	false
--cover-name	是否覆盖名称 指定为true则会将源台的商品名称设置为(商品前缀+源台商品名称)	false
--sort	排序 对应数据库的sort字段,默认0排在最前面	0


命令行工具
接口文档
API 接口文档
共 11 个接口
POST
查询课程
/api.php?act=get
接口地址:http://122.51.236.86/api.php?act=get
请求方式:
POST
Content-Type:application/x-www-form-urlencoded
请求参数
参数名	类型	必填	示例值	说明
user	string	
是
152****7687	学生账号
pass	string	
是
sasdas@1133	学生密码
school	string	
是
职业技术学院	学校信息
platform	integer	
是
60	平台编号 
uid	string	
是
sda******	平台用户名
key	string	
是
1*****	个人中心对接密码
响应示例
{
  "code": 1,
  "data": [
    {
      "name": "创造性思维与创新方法",
      "id": 1000006163,
      "hash": "1//jXTVcLjK×swUpN8S+LogEdK1xxTeIzvc3W9Uc1je5k41qmYSofivVgmbMEqS6R+G+€..."
    }
  ]
}
特别说明
如遇到平台需要区分对应院校时（例如学起等多院校平台），请务必将上述school字段中内容替换为所提供院校的全称
POST
下单
/api.php?act=add
接口地址:http://122.51.236.86/api.php?act=add
请求方式:
POST
Content-Type:application/x-www-form-urlencoded
请求参数
参数名	类型	必填	示例值	说明
user	string	
是
152****7687	学生账号
pass	string	
是
sasdas@1133	学生密码
school	string	
是
职业技术学院	学校名
kcid	integer	
否
1000001	课程ID（优先使用）
kcname	string	
否
创造性思维与创新方法	课程名称（没有课程ID时使用）
platform	integer	
是
60	平台编号
uid	string	
是
sda******	平台用户名
key	string	
是
1*****	个人中心对接密码
config	string	
否
{"key":"value"}	自定义订单配置（JSON格式，参考订单列表的"自定义配置"列）
响应示例
{
  "code": 0,
  "data": [
    "a94c2ef4149fde0594f5dc3bbba322dd"
  ],
  "msg": "下单成功"
}
响应字段说明
字段名	类型	说明
code	integer	状态码：0-成功，其他值表示失败
data	array	订单ID数组，数组中的ID即为源台订单的yid
msg	string	响应消息
重要说明
返回的订单ID即为源台的yid，可直接用于查单、补单、改密、暂停等操作
使用此yid可以解决串进度问题，确保订单进度准确追踪
建议保存此yid到数据库，用于后续订单管理和状态查询
POST
查单
/api.php?act=chadan
接口地址:http://122.51.236.86/api.php?act=chadan
请求方式:
POST
Content-Type:application/x-www-form-urlencoded
请求参数
查询方式：username + 时间范围 或 uuid（二选一）
参数名	类型	必填	示例值	说明
uid	string	
是
sda******	平台用户名
key	string	
是
1*****	个人中心对接密码
username	string	
二选一
152****7687	待查询的账号（与uuid二选一）
uuid	string	
二选一
06113d0208989870bd1e83c74cde569b9	订单ID（二开平台的yid，源台的订单id）
startTime	string	
条件
2024-03-27	起始时间
endTime	string	
条件
2024-04-08	结束时间
响应示例
{
  "code": 1,
  "data": [
    {
      "examStartTime": "-",
      "id": "06113d0208989870bd1e83c74cde569b9",
      "kcname": "服务礼仪",
      "order": {
        "courseName": "服务礼仪",
        "createdAt": "2024-02-26 20:17:28",
        "finish": 0,
        "nextCheckTime": "2024-02-26 20:17:28",
        "password": "sadas",
        "plat": 62,
        "realName": "***",
        "result": "考试完成",
        "school": "职业大学",
        "status": 1,
        "total": 0,
        "updatedAt": "2024-02-26 20:17:34",
        "username": "152****7687",
        "uuid": "06113d0208989870bd1e83c74cde569b9"
      },
      "process": "0/0",
      "remarks": "-",
      "status": "已完成"
    }
  ]
}
POST
补单
/api.php?act=budan
接口地址:http://122.51.236.86/api.php?act=budan
请求方式:
POST
Content-Type:application/x-www-form-urlencoded
请求参数
参数名	类型	必填	示例值	说明
uid	string	
是
sda******	平台用户名
key	string	
是
1*****	个人中心对接密码
id	string	
是
4cjsdiuiadcd6b6274f04b00779	源台订单ID（你们的yid）
响应示例
{
  "code": 1,
  "msg": "补单成功"
}
POST
改密
/api.php?act=gaimi
接口地址:http://122.51.236.86/api.php?act=gaimi
请求方式:
POST
Content-Type:application/x-www-form-urlencoded
请求参数
参数名	类型	必填	示例值	说明
uid	string	
是
sda******	平台用户名
key	string	
是
1*****	个人中心对接密码
id	string	
是
4cjsdiuiadcd6b6274f04b00779	源台订单ID（你们的yid）
newPwd	string	
是
123456	新密码
响应示例
{
  "code": 1,
  "msg": "改密成功"
}
POST
暂停订单
/api.php?act=zanting
接口地址:http://122.51.236.86/api.php?act=zanting
请求方式:
POST
Content-Type:application/x-www-form-urlencoded
请求参数
参数名	类型	必填	示例值	说明
uid	string	
是
sda******	平台用户名
key	string	
是
1*****	个人中心对接密码
id	string	
是
4cjsdiuiadcd6b6274f04b00779	源台订单ID（你们的yid）
响应示例
{
  "code": 1,
  "msg": "操作成功"
}
提示
再次调用此接口或者补单接口可恢复订单
POST
反馈
/api.php?act=report
接口地址:http://122.51.236.86/api.php?act=report
请求方式:
POST
Content-Type:application/x-www-form-urlencoded
请求参数
参数名	类型	必填	示例值	说明
uid	string	
是
sda******	平台用户名
key	string	
是
1*****	个人中心对接密码
id	string	
是
4cjsdiuiadcd6b6274f04b00779	源台订单ID（你们的yid）
question	string	
是
订单执行异常，请帮忙查看	反馈问题描述
响应示例
{
  "code": 1,
  "msg": "反馈成功",
  "data": {
    "reportId": 123,
    "message": "反馈成功，请记录此ID以便后续查询: 123"
  }
}
说明
反馈成功后会自动发送邮件通知管理员，管理员会及时处理您的问题
请务必记录返回的 reportId，可通过此ID查询反馈处理进度
status 字段说明：0-待处理，1-处理完成，3-暂时搁置，4-处理中，6-已退款
POST
查询反馈列表
/api.php?act=getReportList
接口地址:http://122.51.236.86/api.php?act=getReportList
请求方式:
POST
Content-Type:application/x-www-form-urlencoded
请求参数
参数名	类型	必填	示例值	说明
uid	string	
是
sda******	平台用户名
key	string	
是
1*****	个人中心对接密码
status	integer	
否
0	反馈状态筛选：-1或不传-全部，0-待处理，1-处理完成，3-暂时搁置，4-处理中，6-已退款
响应示例
{
  "code": 1,
  "msg": "查询成功",
  "data": [
    {
      "id": 125,
      "orderId": "06113d0208989870bd1e83c74cde569b9",
      "orderInfo": "张三反馈[智慧职教] 职业大学 152****7687 *** 服务礼仪",
      "question": "订单执行异常，请帮忙查看",
      "answer": "已处理，问题已解决",
      "status": 1
    },
    {
      "id": 124,
      "orderId": "06113d0208989870bd1e83c74cde569b9",
      "orderInfo": "张三反馈[智慧职教] 职业大学 152****7687 *** 创新思维",
      "question": "课程进度没有更新",
      "answer": "正在处理中，请稍候",
      "status": 4
    },
    {
      "id": 123,
      "orderId": "06113d0208989870bd1e83c74cde569b9",
      "orderInfo": "张三反馈[智慧职教] 职业大学 152****7687 *** 大学英语",
      "question": "需要退款",
      "answer": "退款已完成",
      "status": 6
    },
    {
      "id": 122,
      "orderId": "06113d0208989870bd1e83c74cde569b9",
      "orderInfo": "张三反馈[智慧职教] 职业大学 152****7687 *** 计算机基础",
      "question": "账号密码错误",
      "answer": null,
      "status": 0
    }
  ]
}
响应字段说明
字段名	类型	说明
id	integer	反馈ID（提交反馈时返回的reportId）
orderId	string	订单UUID
orderInfo	string	订单信息摘要
question	string	问题描述
answer	string	管理员回复（未处理时为null）
status	integer	处理状态：0-待处理，1-处理完成，3-暂时搁置，4-处理中，6-已退款
说明
返回当前用户所有反馈记录，最多100条
可通过 status 参数筛选不同状态的反馈
记录按反馈ID倒序排列（最新的在前）
POST
查询反馈详情
/api.php?act=getReport
接口地址:http://122.51.236.86/api.php?act=getReport
请求方式:
POST
Content-Type:application/x-www-form-urlencoded
请求参数
参数名	类型	必填	示例值	说明
uid	string	
是
sda******	平台用户名
key	string	
是
1*****	个人中心对接密码
reportId	integer	
是
123	反馈ID（提交反馈时返回的reportId）
响应示例
{
  "code": 1,
  "msg": "查询成功",
  "data": {
    "id": 123,
    "orderId": "06113d0208989870bd1e83c74cde569b9",
    "orderInfo": "张三反馈[智慧职教] 职业大学 152****7687 *** 服务礼仪",
    "question": "订单执行异常，请帮忙查看",
    "answer": "已处理，问题已解决",
    "status": 1
  }
}
响应字段说明
字段名	类型	说明
id	integer	反馈ID
orderId	string	订单UUID
orderInfo	string	订单信息摘要
question	string	问题描述
answer	string	管理员回复（未处理时为null）
status	integer	处理状态：0-待处理，1-处理完成，3-暂时搁置，4-处理中，6-已退款
说明
根据反馈ID查询单个反馈的详细信息
只能查询属于当前用户的反馈记录
如果反馈不存在或无权限查看，会返回错误信息
可用于查询反馈的最新处理状态和管理员回复
POST
余额查询
/api.php?act=money
接口地址:http://122.51.236.86/api.php?act=money
请求方式:
POST
Content-Type:application/x-www-form-urlencoded
请求参数
参数名	类型	必填	示例值	说明
uid	string	
是
sda******	平台用户名
key	string	
是
1*****	个人中心对接密码
响应示例
{
  "code": 1,
  "msg": "余额查询成功",
  "data": 1000
}
说明
data 字段返回当前账户余额，单位为币
GET
实时日志
/api/streamLogs
接口地址:http://122.51.236.86/api/streamLogs
请求方式:
GET
请求参数
参数名	类型	必填	示例值	说明
id	string	
是
4cjsdiuiadcd6b6274f04b00779	源台订单ID（你们的yid）
key	string	
是
86476920-bd65-4916-94fd-35ba73d855df	个人中心对接密码
Nginx配置示例
极力不推荐使用PHP包装此接口,推荐在自己的网站上配置Nginx反向代理（如果使用宝塔，点击网站然后点击自己的域名出来弹窗之后点击配置文件）
location /api/streamLogs {
  proxy_buffering off;
  set $args $args&key=86476920-bd65-4916-94fd-35ba73d855df;
  proxy_pass http://122.51.236.86/api/streamLogs;
}
配置后，在你的网站上使用此接口只需要传yid即可：
http://你的网站/api/streamLogs?id=$yid
接入步骤
二开数字订单ID转yid的一个接口或者直接在订单列表返回yid（这个接口需要直接传源台订单ID）
前端页面进行适配推荐使用EventSource，可以参考这篇文章： https://blog.csdn.net/qq_42978535/article/details/142670351