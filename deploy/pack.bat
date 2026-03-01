@echo off
chcp 65001 >nul
echo ========================================
echo   一键打包部署包（Windows 本地执行）
echo ========================================

set ROOT=%~dp0..
set OUT=%ROOT%\deploy\release

:: 清理旧包
if exist "%OUT%" rd /s /q "%OUT%"
mkdir "%OUT%"
mkdir "%OUT%\go-api"
mkdir "%OUT%\go-api\config"
mkdir "%OUT%\go-api\migrations"
mkdir "%OUT%\php-api"

echo.
echo [1/4] 编译 Go 后端...
cd /d "%ROOT%\go-api"
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -o "%OUT%\go-api\go-api-linux" ./cmd/server/
if errorlevel 1 (
    echo ❌ Go 编译失败！
    pause
    exit /b 1
)
copy config\config.yaml "%OUT%\go-api\config\" >nul
xcopy migrations\*.sql "%OUT%\go-api\migrations\" /q >nul
echo ✅ Go 后端编译完成

echo.
echo [2/4] 构建前端...
cd /d "%ROOT%\vben-admin"
call pnpm build:antd
if errorlevel 1 (
    echo ❌ 前端构建失败！
    pause
    exit /b 1
)
echo ✅ 前端构建完成

echo.
echo [3/4] 复制文件...
:: 前端
xcopy "%ROOT%\vben-admin\apps\web-antd\dist\*" "%OUT%\dist\" /s /e /q >nul
:: PHP API
xcopy "%ROOT%\php-api\*" "%OUT%\php-api\" /s /e /q >nul
:: 安装脚本和 Nginx 配置
copy "%ROOT%\deploy\install.sh" "%OUT%\" >nul
copy "%ROOT%\deploy\check_db.sh" "%OUT%\" >nul
copy "%ROOT%\nginx.conf" "%OUT%\" >nul
:: PHP 插件（如有）
if exist "%ROOT%\go-api\deploy\auth_bridge.php" copy "%ROOT%\go-api\deploy\auth_bridge.php" "%OUT%\" >nul
if exist "%ROOT%\go-api\deploy\plugins" xcopy "%ROOT%\go-api\deploy\plugins\*" "%OUT%\plugins\" /s /e /q >nul
echo ✅ 文件复制完成

echo.
echo [4/4] 打包压缩...
cd /d "%OUT%"
tar -czf "%ROOT%\deploy\release.tar.gz" *
echo ✅ 打包完成: deploy\release.tar.gz

echo.
echo ========================================
echo   部署步骤：
echo   1. 上传 deploy\release.tar.gz 到服务器
echo   2. 在服务器执行：
echo      mkdir -p /opt/deploy ^&^& cd /opt/deploy
echo      tar -xzf release.tar.gz
echo      bash install.sh
echo ========================================
pause
