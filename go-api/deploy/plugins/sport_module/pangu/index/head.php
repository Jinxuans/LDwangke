<?php
/**
 * 精简版 head.php — 用于 iframe 嵌入的盘古模块
 * 去掉了反 iframe、反调试、侧边栏/导航栏等不需要的内容
 * 部署位置：服务器根目录/pangu/view/head.php
 */
include('../../confing/common.php');
if($islogin!=1){exit("<script>window.location.href='/index/login';</script>");}

?>
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title><?php echo isset($title) ? $title : '盘古模块'; ?> - <?= $conf['sitename'] ?></title>
  <link rel="shortcut icon" href="<?= $conf['logo']; ?>">
  <link rel="stylesheet" href="/assets/css/oneui.min.css">
  <style>
    body { background: #f5f5f5; padding: 10px; }
  </style>
</head>
<body>
  <div id="page-container" class="main-content-boxed">
    <main id="main-container">
      <div class="content">
