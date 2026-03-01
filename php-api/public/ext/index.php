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
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif; background: #f5f7fa; color: #333; padding: 24px; }
        .container { max-width: 800px; margin: 0 auto; }
        .card { background: #fff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.06); padding: 24px; margin-bottom: 16px; }
        .card h2 { font-size: 18px; margin-bottom: 12px; color: #1890ff; }
        .card p { color: #666; line-height: 1.8; }
        .info-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #f0f0f0; }
        .info-row:last-child { border-bottom: none; }
        .info-label { color: #999; }
        .info-value { font-weight: 500; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <h2>PHP 扩展页面</h2>
            <p>这是一个通过扩展菜单嵌入的 PHP 单页示例。你可以在此基础上开发自定义功能。</p>
        </div>
        <div class="card">
            <h2>服务器信息</h2>
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
