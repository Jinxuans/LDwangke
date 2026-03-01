<?php
/**
 * 扩展菜单示例页面
 * 通过 /php-api/ext/index.php 访问
 */
?>
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>扩展页面</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', Arial, sans-serif;
            background: linear-gradient(135deg, #f0f4ff 0%, #f5f7fa 50%, #faf0ff 100%);
            color: #333;
            padding: 20px;
            min-height: 100vh;
        }
        .container { max-width: 640px; margin: 0 auto; }
        .page-header {
            text-align: center;
            padding: 24px 0 16px;
        }
        .page-header h1 {
            font-size: 20px;
            font-weight: 600;
            color: #1a1a2e;
            letter-spacing: 0.5px;
        }
        .page-header .subtitle {
            font-size: 13px;
            color: #999;
            margin-top: 4px;
        }
        .card {
            background: #fff;
            border-radius: 12px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.03);
            padding: 20px;
            margin-bottom: 12px;
            border: 1px solid rgba(0,0,0,0.04);
            transition: box-shadow 0.2s;
        }
        .card:hover { box-shadow: 0 2px 6px rgba(0,0,0,0.06), 0 8px 20px rgba(0,0,0,0.04); }
        .card-header {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 12px;
            padding-bottom: 10px;
            border-bottom: 1px solid #f5f5f5;
        }
        .card-icon {
            width: 32px; height: 32px;
            border-radius: 8px;
            display: flex; align-items: center; justify-content: center;
            font-size: 16px;
            flex-shrink: 0;
        }
        .card-icon.blue { background: #e8f4fd; color: #1890ff; }
        .card-icon.green { background: #e6f7ed; color: #52c41a; }
        .card-header h2 {
            font-size: 15px;
            font-weight: 600;
            color: #1a1a2e;
        }
        .card p {
            color: #666;
            line-height: 1.7;
            font-size: 13px;
        }
        .info-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 10px 0;
            border-bottom: 1px solid #fafafa;
            gap: 12px;
        }
        .info-row:last-child { border-bottom: none; }
        .info-label {
            color: #999;
            font-size: 13px;
            flex-shrink: 0;
        }
        .info-value {
            font-weight: 500;
            font-size: 13px;
            color: #333;
            text-align: right;
            word-break: break-all;
            min-width: 0;
        }

        @media (max-width: 480px) {
            body { padding: 12px 10px; }
            .page-header { padding: 12px 0 10px; }
            .page-header h1 { font-size: 17px; }
            .page-header .subtitle { font-size: 12px; }
            .card { padding: 14px; margin-bottom: 10px; border-radius: 10px; }
            .card-header { margin-bottom: 10px; padding-bottom: 8px; }
            .card-icon { width: 28px; height: 28px; font-size: 14px; border-radius: 6px; }
            .card-header h2 { font-size: 14px; }
            .card p { font-size: 12px; line-height: 1.6; }
            .info-row { padding: 8px 0; }
            .info-label { font-size: 12px; }
            .info-value { font-size: 12px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="page-header">
            <h1>扩展页面</h1>
            <div class="subtitle">PHP Extension Page</div>
        </div>
        <div class="card">
            <div class="card-header">
                <div class="card-icon blue">⚡</div>
                <h2>PHP 扩展页面</h2>
            </div>
            <p>这是一个通过扩展菜单嵌入的 PHP 单页示例。你可以在此基础上开发自定义功能。</p>
        </div>
        <div class="card">
            <div class="card-header">
                <div class="card-icon green">🖥</div>
                <h2>服务器信息</h2>
            </div>
            <div class="info-row">
                <span class="info-label">PHP 版本</span>
                <span class="info-value"><?php echo phpversion(); ?></span>
            </div>
            <div class="info-row">
                <span class="info-label">服务器时间</span>
                <span class="info-value"><?php echo date('Y-m-d H:i:s'); ?></span>
            </div>
            <div class="info-row">
                <span class="info-label">服务器软件</span>
                <span class="info-value"><?php echo $_SERVER['SERVER_SOFTWARE'] ?? 'N/A'; ?></span>
            </div>
            <div class="info-row">
                <span class="info-label">请求来源</span>
                <span class="info-value"><?php echo $_SERVER['HTTP_REFERER'] ?? '直接访问'; ?></span>
            </div>
        </div>
    </div>
</body>
</html>
