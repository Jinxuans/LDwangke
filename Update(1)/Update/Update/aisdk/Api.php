<?php

class Api
{
    private $url = 'http://www.zwgflw.top/';

    /**
     * 登录Token
     * @var string
     */
    private $token = '';


    /**
     * @var string 接口账号
     */
    public $username = '';


    /**
     * @var string 接口密码
     */
    public $password = '';

    /**
     * 构造方法
     */
    public function __construct($conf)
    {
        $this->username = $conf['lunwen_api_username'];
        $this->password = $conf['lunwen_api_password'];
        $this->token = $this->login();
    }

    /**
     * 登录
     * @return string
     * @throws \Exception
     */
    protected function login($isReset = false)
    {
        
        if(!empty($_SESSION['lunwentoken']) && !$isReset){
            $this->token = $_SESSION['lunwentoken'];
            return $_SESSION['lunwentoken'];
        }
        $username = $this->username;
        $password = $this->password;
        if (empty($username) || empty($password)) {
            return false;
        }
        $postData = json_encode([
            'username' => $username,
            'password' => $password
        ]);

        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $this->url . 'prod-api/login');
        curl_setopt($ch, CURLOPT_POST, 1);
        curl_setopt($ch, CURLOPT_POSTFIELDS, $postData);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json'
        ]);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

        $response = curl_exec($ch);
        if (curl_errno($ch)) {
            throw new \Exception(curl_error($ch));
        }
        curl_close($ch);

        $ret = json_decode($response, true);
        if ($ret['code'] == 200) {
            $this->token = $ret['token'];
            $_SESSION["lunwentoken"] = $ret['token'];
            return $ret['token'];
        }
        throw new \Exception($ret['msg']);
    }
    /**
     * 发送GET请求
     * @param $path
     * @param $data
     * @return array
     */
    public function get($path, $data = [])
    {

        $url = $this->url . $path;
        if ($data) {
            $url .= '?' . http_build_query($data);
        }
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_POST, 0);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Authorization: Bearer ' . $this->token,
        ]);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        //$response = curl_exec($ch);
        $maxRetries = 3;
        $retryDelay = 2; // 秒
        for ($i = 0; $i < $maxRetries; $i++) {
            $response = curl_exec($ch);
            if (!curl_errno($ch)) break;
            sleep($retryDelay);
        }
        if (curl_errno($ch)) {
            throw new \Exception(curl_error($ch));
        }
        curl_close($ch);

        $ret = json_decode($response, true);
        if ($ret['code'] == 200) {
            return $ret;
        }
         if ($ret['code'] == 401) {
             $this->login(true);
             $this->get($path,$data);
         }
        throw new \Exception($ret['msg']);
    }

    /**
     * 发送POST请求
     * @param $path
     * @param $data
     * @return array
     */
    public function post($path, $data = [])
    {
        $url = $this->url . $path;
        $ch = curl_init();

        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_POST, 1);
        if ($data) {
            $postData = json_encode($data);
            curl_setopt($ch, CURLOPT_POSTFIELDS, $postData);
        }

        $headers = [
            'Content-Type: application/json',
            'Authorization: Bearer ' . $this->token,
        ];
        curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        $response = curl_exec($ch);
        if (curl_errno($ch)) {
            throw new \Exception(curl_error($ch));
        }
        curl_close($ch);

        $ret = json_decode($response, true);
        if ($ret['code'] == 200) {
            return $ret;
        }
         if ($ret['code'] == 401) {
             $this->login(true);
             $this->post($path, $data);
         }
        throw new \Exception($ret['msg']);
    }

    /**
     * 发送Stream请求
     * @param $path
     * @param $data
     * @return array
     */
    public function streamRequest($path, $params = [])
    {
        // 彻底禁用所有输出缓冲
        while (ob_get_level()) {
            ob_end_clean();
        }

        // 设置立即刷新
        ob_implicit_flush(true);

        $url = $this->url . $path;

        // 设置头信息
        $headers = [
            'Content-Type: application/json',
            'Accept: text/event-stream',
            'Cache-Control: no-cache',
            'Connection: keep-alive',
            'X-Accel-Buffering: no', // 特别针对Nginx
            'Authorization: Bearer ' . $this->token,
        ];

        // 初始化cURL
        $ch = curl_init();

        // 设置cURL选项
        curl_setopt_array($ch, [
            CURLOPT_URL => $url,
            CURLOPT_POST => true,
            CURLOPT_POSTFIELDS => json_encode($params),
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_RETURNTRANSFER => false,
            CURLOPT_HEADER => false,
            CURLOPT_WRITEFUNCTION => function ($ch, $data) {
                // 立即输出数据
                echo $data;
                ob_flush();
                flush();
                return strlen($data);
            },
            // 关键选项 - 禁用cURL内部缓冲
            CURLOPT_BUFFERSIZE => 128, // 小缓冲区
            CURLOPT_NOPROGRESS => false,
            CURLOPT_PROGRESSFUNCTION => function (
                $resource,
                $download_size,
                $downloaded,
                $upload_size,
                $uploaded
            ) {
                // 强制立即处理数据
                return 0;
            },
            CURLOPT_TIMEOUT => 0, // 无超时限制
        ]);

        // 设置客户端响应头
        header('Content-Type: text/event-stream');
        header('Cache-Control: no-cache, must-revalidate');
        header('Connection: keep-alive');
        header('X-Accel-Buffering: no');

        // 执行请求
        $response = curl_exec($ch);

        // 错误处理
        if ($response === false) {
            echo "event: error\ndata: " . json_encode([
                'error' => curl_error($ch),
                'errno' => curl_errno($ch)
            ]) . "\n\n";
            ob_flush();
            flush();
        }

        // 关闭cURL资源
        curl_close($ch);
    }


    /**
     * 获取商品列表
     * @return array
     * @throws Exception
     */
    public function getShopList()
    {
        $res = $this->get('prod-api/wk/ShopInfo/getShopList', ['type' => 1]);
        return $res;
    }

    /**
     * 获取商品价格
     * @param $ShopCode
     * @return array
     * @throws Exception
     */
    public function getShopPrice($ShopCode = 'ktbg')
    {
        $res = $this->get('prod-api/wk/userPrice/getShopPrice', ['ShopCode' => $ShopCode]);
        return $res;
    }

    /**
     * 获取附加模板
     * @param $pageNum
     * @param $pageSize
     * @return array
     * @throws Exception
     */
    public function getTemplate($params)
    {
        $res = $this->get('prod-api/system/template/list', $params);
        return $res;
    }


    /**
     * 生成任务书
     * @param array $params 包含'id'参数的数组
     * @return array
     * @throws Exception
     */
    public function generateTask($params)
    {
        $res = $this->post('prod-api/system/lunwen/generate-task', $params);
        return $res;
    }

    public function systemtemplate($params)
    {
        $res = $this->post('prod-api/system/template', $params);
        return $res;
    }
    /**
     * 生成开题报告
     * @param array $params 包含'id'参数的数组
     * @return array
     * @throws Exception
     */
    public function generateProposal($params)
    {
        $res = $this->post('prod-api/system/lunwen/generate-proposal', $params);
        return $res;
    }
    /**
     * 获取论文列表
     * @param $data
     * @return array
     * @throws Exception
     */
    public function getList($data = [])
    {
        $res = $this->get('prod-api/system/lunwen/list', $data);
        return $res;
    }

    /**
     * 生成论文标题
     * @param $direction
     * @return array
     * @throws Exception
     */
    public function generateTitles($params)
    {

        $res = $this->post('prod-api/system/lunwen/generate-titles', $params);
        return $res;
    }

    /**
     * 生成论文大纲
     * @param $title
     * @param $wordCount
     * @param string $customRequirements
     * @return array
     * @throws Exception
     */
    public function generateOutline($params)
    {
        $res = $this->post('prod-api/system/lunwen/generate-outline', $params);
        return $res;
    }

    /**
     * 获取论文大纲状态
     * @param $orderid
     * @return array
     * @throws Exception
     */
    public function outlineStatus($orderId)
    {
        $res = $this->get("prod-api/system/lunwen/outline-status/{$orderId}");
        return $res;
    }

    /**
     * 上传文件
     * @param $path
     * @param $file
     * @return void
     */
    public function file($path, $fileInfo, $params = [])
    {
        $url = $this->url . $path;
        $ch = curl_init();
        $filePath = $fileInfo['tmp_name'];
        // 使用 CURLFile 包装文件
        $file = new CURLFile(
            $filePath,                  // 文件路径
            mime_content_type($filePath), // MIME 类型（如 'application/octet-stream'）
            $fileInfo['name']         // 文件名
        );


        $params['file'] = $file;

        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_POST, 1);
        curl_setopt($ch, CURLOPT_POSTFIELDS, $params);
        $headers = [
            'Content-Type: multipart/form-data',
            'Authorization: Bearer ' . $this->token,
        ];
        curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

        $response = curl_exec($ch);

        if (curl_errno($ch)) {
            throw new \Exception(curl_error($ch));
        }
        curl_close($ch);
        $ret = json_decode($response, true);
        if ($ret['code'] == 200) {
            return $ret;
        }
         if ($ret['code'] == 401) {
             $this->login(true);
             $this->file($path, $fileInfo, $params);
         }
        throw new \Exception($ret['msg']);
    }

    /**
     * 论文下单
     * @param $orderid
     * @return array
     * @throws Exception
     */
    public function paperOrder($params)
    {
        $res = $this->post('prod-api/system/lunwen/xiadan', $params);
        return $res;
    }


    /**
     * 段落修改
     *
     * @param [type] $data
     * @return void
     */
    public function paperParaEditApi($params)
    {
        $res = $this->streamRequest("prod-api/system/lunwen/xiugai/stream", $params);
        return $res;
    }

    /**
     * 论文下载
     * @param $orderid
     * @return array
     * @throws Exception
     */
    public function paperDownload($orderid, $fileName)
    {
        $urlencodeFileName = urlencode($fileName);
        $res = $this->get("prod-api/system/lunwen/download/{$orderid}?fileName={$urlencodeFileName}");
        return $res;
    }

    /**
     * 文本降重
     * @param $orderid
     * @return array
     * @throws Exception
     */
    public function textPaperRewrite($params)
    {
        $res = $this->streamRequest("prod-api/system/lunwen/rewrite/stream", $params);
        return $res;
    }
    /**
     * 降低AIGC
     * @param $orderid
     * @return array
     * @throws Exception
     */
    public function textRewriteAigc($params)
    {
        $res = $this->streamRequest("prod-api/system/lunwen/rewrite-aigc/stream", $params);
        return $res;
    }
    /**
     * 统计WORDS字数
     * @param $orderid
     * @return array
     * @throws Exception
     */
    public function fileCountWords($file)
    {
        $res = $this->file("prod-api/system/lunwen/countWords", $file);
        return $res;
    }
    /**
     * 上传模板文件
     * @param $orderid
     * @return array
     * @throws Exception
     */
    public function fileuploadCover($file)
    {
        $res = $this->file("prod-api/system/template/uploadCover", $file);
        return $res;
    }
    /**
     * 文件降重
     * @param $orderid
     * @return array
     * @throws Exception
     */
    public function fileDedup($file, $params)
    {
        $res = $this->file("prod-api/system/lunwen/jiangchong", $file, $params);
        return $res;
    }
}
